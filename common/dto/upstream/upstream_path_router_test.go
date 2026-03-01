// author: copilot
package upstream

import "testing"

func TestPathRouter_ExactMatch(t *testing.T) {
	router := NewUpStreamPathRouter()

	path1 := &UpstreamPath{Path: "/users", Method: "GET"}
	path2 := &UpstreamPath{Path: "/users/profile", Method: "GET"}
	path3 := &UpstreamPath{Path: "/products", Method: "GET"}

	router.Insert(path1)
	router.Insert(path2)
	router.Insert(path3)

	tests := []struct {
		name     string
		path     string
		expected *UpstreamPath
	}{
		{"match users", "/users", path1},
		{"match users profile", "/users/profile", path2},
		{"match products", "/products", path3},
		{"no match", "/admins", nil},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := router.Search(tt.path)
			if result != tt.expected {
				t.Errorf("Expected %v, got %v", tt.expected, result)
			}
		})
	}
}

func TestPathRouter_PathVariables(t *testing.T) {
	router := NewUpStreamPathRouter()

	path1 := &UpstreamPath{Path: "/users/{id}", Method: "GET"}
	path2 := &UpstreamPath{Path: "/users/{userId}/posts/{postId}", Method: "GET"}
	path3 := &UpstreamPath{Path: "/products/{productId}", Method: "GET"}

	router.Insert(path1)
	router.Insert(path2)
	router.Insert(path3)

	tests := []struct {
		name     string
		path     string
		expected *UpstreamPath
	}{
		{"match user by id", "/users/123", path1},
		{"match user by id string", "/users/abc", path1},
		{"match user post", "/users/123/posts/456", path2},
		{"match product", "/products/999", path3},
		{"no match wrong depth", "/users/123/posts", nil},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := router.Search(tt.path)
			if result != tt.expected {
				t.Errorf("Expected %v, got %v", tt.expected, result)
			}
		})
	}
}

func TestPathRouter_MixedStaticAndDynamic(t *testing.T) {
	router := NewUpStreamPathRouter()

	staticPath := &UpstreamPath{Path: "/users/profile", Method: "GET"}
	dynamicPath := &UpstreamPath{Path: "/users/{id}", Method: "GET"}

	router.Insert(staticPath)
	router.Insert(dynamicPath)

	tests := []struct {
		name     string
		path     string
		expected *UpstreamPath
	}{
		{"exact match has priority", "/users/profile", staticPath},
		{"dynamic match", "/users/123", dynamicPath},
		{"dynamic match with string", "/users/john", dynamicPath},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := router.Search(tt.path)
			if result != tt.expected {
				t.Errorf("Expected %v, got %v", tt.expected, result)
			}
		})
	}
}

func TestUpStreamHost_LookupPath(t *testing.T) {
	host := &UptreamHost{
		Host: "example.com",
		Request: []*UpstreamPath{
			{Path: "/api/users", Method: "GET"},
			{Path: "/api/users/{id}", Method: "GET"},
			{Path: "/api/users/{userId}/posts/{postId}", Method: "GET"},
			{Path: "/api/products", Method: "GET"},
		},
	}

	host.InitializeRouter()

	tests := []struct {
		name         string
		requestPath  string
		shouldMatch  bool
		expectedPath string
	}{
		{"exact match", "/api/users", true, "/api/users"},
		{"path variable", "/api/users/123", true, "/api/users/{id}"},
		{"nested path variables", "/api/users/123/posts/456", true, "/api/users/{userId}/posts/{postId}"},
		{"no match", "/api/orders", false, ""},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := host.LookupPath(tt.requestPath)

			if tt.shouldMatch {
				if result == nil {
					t.Errorf("Expected to find match for %s", tt.requestPath)
				} else if result.Path != tt.expectedPath {
					t.Errorf("Expected path %s, got %s", tt.expectedPath, result.Path)
				}
			} else {
				if result != nil {
					t.Errorf("Expected no match for %s, but got %v", tt.requestPath, result)
				}
			}
		})
	}
}

func BenchmarkPathRouter_Search(b *testing.B) {
	router := NewUpStreamPathRouter()

	paths := []string{
		"/api/users",
		"/api/users/{id}",
		"/api/users/{id}/posts",
		"/api/users/{id}/posts/{postId}",
		"/api/products",
		"/api/products/{productId}",
		"/api/orders",
		"/api/orders/{orderId}",
	}

	for _, p := range paths {
		router.Insert(&UpstreamPath{Path: p, Method: "GET"})
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		router.Search("/api/users/123/posts/456")
	}
}
