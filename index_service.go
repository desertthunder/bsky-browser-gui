package main

import (
	"context"
	"encoding/json"
	"fmt"
	"reflect"
	"sync"
	"sync/atomic"
	"time"

	"github.com/bluesky-social/indigo/api/bsky"
	"github.com/bluesky-social/indigo/atproto/auth/oauth"
	"github.com/bluesky-social/indigo/atproto/syntax"
	"github.com/wailsapp/wails/v2/pkg/runtime"
)

// IndexService provides indexing functionality via Wails bindings
type IndexService struct {
	ctx      context.Context
	indexing atomic.Bool
	stats    IndexStats
	statsMu  sync.RWMutex
}

// IndexStats tracks indexing progress
type IndexStats struct {
	Fetched  int `json:"fetched"`
	Inserted int `json:"inserted"`
	Errors   int `json:"errors"`
	Total    int `json:"total"`
}

// IndexResult contains the final indexing result
type IndexResult struct {
	Total   int           `json:"total"`
	Errors  int           `json:"errors"`
	Elapsed time.Duration `json:"elapsed"`
}

// PostResult carries either a Post or an error from fetching
type PostResult struct {
	Post  *Post
	Error error
}

// NewIndexService creates a new IndexService instance
func NewIndexService() *IndexService {
	return &IndexService{}
}

// SetContext sets the Wails context for event emission
func (s *IndexService) SetContext(ctx context.Context) {
	s.ctx = ctx
}

// IsIndexing returns true if indexing is currently in progress
func (s *IndexService) IsIndexing() bool {
	return s.indexing.Load()
}

// Refresh fetches bookmarks and likes concurrently and indexes them
func (s *IndexService) Refresh(limit int) error {
	if !s.indexing.CompareAndSwap(false, true) {
		return fmt.Errorf("indexing already in progress")
	}
	defer s.indexing.Store(false)

	start := time.Now()

	s.statsMu.Lock()
	s.stats = IndexStats{}
	s.statsMu.Unlock()

	s.emitEvent("index:started", map[string]any{})

	client, err := s.createClient()
	if err != nil {
		s.emitEvent("index:done", IndexResult{Errors: 1, Elapsed: time.Since(start)})
		return err
	}

	postCh := make(chan *PostResult, 100)
	batchSize := 10

	var wg sync.WaitGroup
	wg.Add(2)

	go func() {
		defer wg.Done()
		client.fetchBookmarks(limit, postCh, s)
	}()

	go func() {
		defer wg.Done()
		client.fetchLikes(limit, postCh, s)
	}()

	go func() {
		wg.Wait()
		close(postCh)
	}()

	successCount, errorCount := s.batchWriter(postCh, batchSize)

	result := IndexResult{
		Total:   successCount + errorCount,
		Errors:  errorCount,
		Elapsed: time.Since(start),
	}

	s.emitEvent("index:done", result)
	return nil
}

// emitEvent emits a Wails event with the given name and data
func (s *IndexService) emitEvent(name string, data any) {
	if s.ctx != nil {
		runtime.EventsEmit(s.ctx, name, data)
	}
}

// updateProgress updates stats and emits progress event
func (s *IndexService) updateProgress(fetched, inserted, errors int) {
	s.statsMu.Lock()
	s.stats.Fetched += fetched
	s.stats.Inserted += inserted
	s.stats.Errors += errors
	stats := s.stats
	s.statsMu.Unlock()

	s.emitEvent("index:progress", stats)
}

// createClient creates an authenticated Bluesky client
func (s *IndexService) createClient() (*BlueskyClient, error) {
	ctx := context.Background()

	auth, err := GetAuth()
	if err != nil {
		return nil, fmt.Errorf("failed to load auth: %w", err)
	}
	if auth == nil {
		return nil, fmt.Errorf("not authenticated")
	}

	if auth.SessionID == "" {
		return nil, fmt.Errorf("session not found")
	}

	did, err := syntax.ParseDID(auth.DID)
	if err != nil {
		return nil, fmt.Errorf("invalid DID: %w", err)
	}

	store := NewSQLiteOAuthStore()
	app := newOAuthApp(store)

	session, err := app.ResumeSession(ctx, did, auth.SessionID)
	if err != nil {
		return nil, fmt.Errorf("failed to resume session: %w", err)
	}

	return &BlueskyClient{
		session: session,
		auth:    auth,
	}, nil
}

// batchWriter reads from channel and inserts posts in batches
func (s *IndexService) batchWriter(ch <-chan *PostResult, batchSize int) (int, int) {
	batch := make([]*Post, 0, batchSize)
	successCount := 0
	errorCount := 0

	flushBatch := func() {
		if len(batch) == 0 {
			return
		}

		for _, post := range batch {
			if err := InsertPost(post); err != nil {
				errorCount++
				s.updateProgress(0, 0, 1)
			} else {
				successCount++
				s.updateProgress(0, 1, 0)
			}
		}
		batch = batch[:0]
	}

	for result := range ch {
		if result.Error != nil {
			errorCount++
			s.updateProgress(0, 0, 1)
			continue
		}

		if result.Post != nil {
			batch = append(batch, result.Post)
			s.updateProgress(1, 0, 0)

			if len(batch) >= batchSize {
				flushBatch()
			}
		}
	}

	flushBatch()
	return successCount, errorCount
}

// BlueskyClient wraps an authenticated OAuth session
type BlueskyClient struct {
	session *oauth.ClientSession
	auth    *Auth
}

// fetchBookmarks writes bookmarks to the provided channel in batches
func (c *BlueskyClient) fetchBookmarks(maxPosts int, ch chan<- *PostResult, _ *IndexService) {
	ctx := context.Background()
	apiClient := c.session.APIClient()
	var cursor string
	batchSize := int64(100)
	count := 0

	for {
		resp, err := bsky.BookmarkGetBookmarks(ctx, apiClient, cursor, batchSize)
		if err != nil {
			ch <- &PostResult{Error: fmt.Errorf("failed to fetch bookmarks: %w", err)}
			return
		}

		for _, bookmark := range resp.Bookmarks {
			if bookmark.Item == nil {
				continue
			}

			if bookmark.Item.FeedDefs_PostView != nil {
				pv := bookmark.Item.FeedDefs_PostView

				exists, err := PostExists(pv.Uri)
				if err != nil {
					continue
				}
				if exists {
					continue
				}

				post := c.convertPostView(pv, "saved")
				if post != nil {
					ch <- &PostResult{Post: post}
					count++

					if maxPosts > 0 && count >= maxPosts {
						return
					}
				}
			}
		}

		if resp.Cursor == nil || *resp.Cursor == "" {
			break
		}
		cursor = *resp.Cursor
	}
}

// fetchLikes writes likes to the provided channel in batches
func (c *BlueskyClient) fetchLikes(maxPosts int, ch chan<- *PostResult, _ *IndexService) {
	ctx := context.Background()
	apiClient := c.session.APIClient()
	var cursor string
	batchSize := int64(100)
	count := 0

	for {
		resp, err := bsky.FeedGetActorLikes(ctx, apiClient, c.auth.DID, cursor, batchSize)
		if err != nil {
			ch <- &PostResult{Error: fmt.Errorf("failed to fetch likes: %w", err)}
			return
		}

		for _, feedView := range resp.Feed {
			if feedView.Post != nil {
				pv := feedView.Post

				exists, err := PostExists(pv.Uri)
				if err != nil {
					continue
				}
				if exists {
					continue
				}

				post := c.convertPostView(pv, "liked")
				if post != nil {
					ch <- &PostResult{Post: post}
					count++

					if maxPosts > 0 && count >= maxPosts {
						return
					}
				}
			}
		}

		if resp.Cursor == nil || *resp.Cursor == "" {
			break
		}
		cursor = *resp.Cursor
	}
}

// convertPostView converts a FeedDefs_PostView to our Post struct
func (c *BlueskyClient) convertPostView(pv *bsky.FeedDefs_PostView, source string) *Post {
	if pv == nil {
		return nil
	}

	record, facets, err := c.parsePostRecord(pv.Record)
	if err != nil {
		record = &postRecord{Text: "", CreatedAt: pv.IndexedAt}
	}

	var authorDID, authorHandle string
	if pv.Author != nil {
		authorDID = pv.Author.Did
		authorHandle = pv.Author.Handle
	}

	likeCount := 0
	if pv.LikeCount != nil {
		likeCount = int(*pv.LikeCount)
	}

	repostCount := 0
	if pv.RepostCount != nil {
		repostCount = int(*pv.RepostCount)
	}

	replyCount := 0
	if pv.ReplyCount != nil {
		replyCount = int(*pv.ReplyCount)
	}

	createdAt, err := syntax.ParseDatetimeLenient(record.CreatedAt)
	if err != nil {
		createdAt, _ = syntax.ParseDatetimeLenient(pv.IndexedAt)
	}

	return &Post{
		URI:          pv.Uri,
		CID:          pv.Cid,
		AuthorDID:    authorDID,
		AuthorHandle: authorHandle,
		Text:         record.Text,
		CreatedAt:    createdAt.Time(),
		LikeCount:    likeCount,
		RepostCount:  repostCount,
		ReplyCount:   replyCount,
		Source:       source,
		Facets:       facets,
	}
}

// postRecord represents the expected structure of a post record
type postRecord struct {
	Text      string `json:"text"`
	CreatedAt string `json:"createdAt"`
}

// parsePostRecord extracts post data and facets from the LexiconTypeDecoder
func (c *BlueskyClient) parsePostRecord(decoder any) (*postRecord, string, error) {
	if decoder == nil {
		return &postRecord{Text: "", CreatedAt: ""}, "", nil
	}

	type lexDecoder struct{ Val any }

	d, ok := decoder.(*lexDecoder)
	if !ok {
		switch v := decoder.(type) {
		case *bsky.FeedPost:
			facets := c.extractFacets(v)
			return &postRecord{
				Text:      v.Text,
				CreatedAt: v.CreatedAt,
			}, facets, nil
		case bsky.FeedPost:
			facets := c.extractFacets(&v)
			return &postRecord{
				Text:      v.Text,
				CreatedAt: v.CreatedAt,
			}, facets, nil
		default:
			return c.parsePostRecordWithReflection(decoder)
		}
	}

	if d.Val == nil {
		return &postRecord{Text: "", CreatedAt: ""}, "", nil
	}

	if feedPost, ok := d.Val.(*bsky.FeedPost); ok {
		facets := c.extractFacets(feedPost)
		return &postRecord{
			Text:      feedPost.Text,
			CreatedAt: feedPost.CreatedAt,
		}, facets, nil
	}

	return &postRecord{Text: "", CreatedAt: ""}, "", fmt.Errorf("unknown record type: %T", d.Val)
}

// extractFacets extracts and serializes facets from a FeedPost
func (c *BlueskyClient) extractFacets(feedPost *bsky.FeedPost) string {
	if feedPost == nil || len(feedPost.Facets) == 0 {
		return ""
	}

	facetsJSON, err := json.Marshal(feedPost.Facets)
	if err != nil {
		return ""
	}

	return string(facetsJSON)
}

// parsePostRecordWithReflection uses reflection to access the Val field
func (c *BlueskyClient) parsePostRecordWithReflection(decoder any) (*postRecord, string, error) {
	val := reflect.ValueOf(decoder)
	if val.Kind() == reflect.Pointer {
		val = val.Elem()
	}

	valField := val.FieldByName("Val")
	if !valField.IsValid() {
		return &postRecord{Text: "", CreatedAt: ""}, "", fmt.Errorf("no Val field found")
	}

	actualVal := valField.Interface()
	if actualVal == nil {
		return &postRecord{Text: "", CreatedAt: ""}, "", nil
	}

	if feedPost, ok := actualVal.(*bsky.FeedPost); ok {
		facets := c.extractFacets(feedPost)
		return &postRecord{
			Text:      feedPost.Text,
			CreatedAt: feedPost.CreatedAt,
		}, facets, nil
	}

	return &postRecord{Text: "", CreatedAt: ""}, "", fmt.Errorf("unknown record type in Val: %T", actualVal)
}
