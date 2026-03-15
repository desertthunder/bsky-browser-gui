package main

// SearchService provides search functionality via Wails bindings
type SearchService struct{}

// NewSearchService creates a new SearchService instance
func NewSearchService() *SearchService {
	return &SearchService{}
}

// Search performs an FTS5 search with BM25 ranking and optional source filter
func (s *SearchService) Search(query string, source string) ([]SearchResult, error) {
	return SearchPosts(query, source)
}

// CountPosts returns the total number of indexed posts
func (s *SearchService) CountPosts() (int, error) {
	return CountPosts()
}
