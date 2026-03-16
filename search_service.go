package main

// SearchService provides search functionality via Wails bindings
type SearchService struct{}

// NewSearchService creates a new SearchService instance
func NewSearchService() *SearchService {
	return &SearchService{}
}

// Search performs an FTS5 search with BM25 ranking, optional source filter, and server-side sorting
func (s *SearchService) Search(query string, source string, limit int, sortColumn string, sortDirection string) ([]SearchResult, error) {
	return SearchPosts(query, source, limit, sortColumn, sortDirection)
}

// CountPosts returns the total number of indexed posts
func (s *SearchService) CountPosts() (int, error) {
	return CountPosts()
}
