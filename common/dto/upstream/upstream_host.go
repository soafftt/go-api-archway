package upstream

type UptreamHost struct {
	Host       string              `json:"host"`    // service host
	Request    []*UpstreamPath     `json:"request"` // proxy metadata
	pathRouter *UpStreamPathRouter // internal Trie router for fast path lookup
}

// InitializeRouter builds the Trie router from Request paths
func (u *UptreamHost) InitializeRouter() {
	u.pathRouter = NewUpStreamPathRouter()
	for _, p := range u.Request {
		u.pathRouter.Insert(p)
	}
}

// LookupPath finds a matching path using Trie-based routing
// Supports both exact matches and path variables (e.g., /users/{id})
func (u *UptreamHost) LookupPath(path string) *UpstreamPath {
	// Lazy initialization
	if u.pathRouter == nil {
		u.InitializeRouter()
	}

	return u.pathRouter.Search(path)
}
