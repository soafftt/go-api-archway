package upstream

import "strings"

// UpStreamPathRouter is a Trie-based path router that supports path variables
type UpStreamPathRouter struct {
	root *upstreamPathNode
}

// trie 알고리즘으로 구현된 UpStreamPathRouter는 다음과 같은 특징을 가집니다:
// 1. 경로를 '/' 기준으로 분할하여 트리 구조로 저장합니다.
// 2. 각 노드는 고유한 경로 세그먼트를 나타내며, path variable (예: {id})도 지원합니다.
// 3. 검색 시에는 먼저 정확한 경로 세그먼트 매칭을 시도하고, 그 다음 path variable 매칭을 시도합니다.
// 4. 경로 변수는 중괄호로 감싸진 형태로 표현되며, 예를 들어 "/users/{id}"는 "id"라는 path variable을 가집니다.
// 5. 라우터는 UpStreamPath 객체를 리프 노드에 저장하여, 요청이 일치하는 경우 해당 UpStreamPath 정보를 반환할 수 있습니다.
type upstreamPathNode struct {
	child        map[string]*upstreamPathNode // match path 를 / 기준으로 분할하여 저장 (root 는 빈공백)
	pathVariable string                       // 현재 노드의 path variable 이름 (예: {id} -> id)
	childParam   *upstreamPathNode            // path variable 이 존재 하는 경우 하위의 노드데이터
	path         *UpstreamPath                // root 노드에만 저장되는 UpStreamPath 데이터
}

// NewUpStreamPathRouter creates a new PathRouter
func NewUpStreamPathRouter() *UpStreamPathRouter {
	return &UpStreamPathRouter{
		root: &upstreamPathNode{
			child: make(map[string]*upstreamPathNode),
		},
	}
}

// Insert adds a path to the router
func (uspr *UpStreamPathRouter) Insert(upstreamPath *UpstreamPath) {
	segments := splitPath(upstreamPath.Path)
	node := uspr.root

	for _, segment := range segments {
		if isPathVariable(segment) {
			// Handle path variable {id}, {userId}, etc.
			paramName := extractParamName(segment)

			if node.childParam == nil {
				node.childParam = &upstreamPathNode{
					child:        make(map[string]*upstreamPathNode),
					pathVariable: paramName,
				}
			}
			node = node.childParam
		} else {
			// Handle static segment
			if _, exists := node.child[segment]; !exists {
				node.child[segment] = &upstreamPathNode{
					child: make(map[string]*upstreamPathNode),
				}
			}
			node = node.child[segment]
		}
	}

	// Set the upstream path at the leaf node
	node.path = upstreamPath
}

// Search finds a matching path in the router
func (uspr *UpStreamPathRouter) Search(requestPath string) *UpstreamPath {
	if uspr.root == nil {
		return nil
	}

	segments := splitPath(requestPath)
	return uspr.search(uspr.root, segments, 0)
}

// search recursively searches for a matching path
func (uspr *UpStreamPathRouter) search(node *upstreamPathNode, segments []string, index int) *UpstreamPath {
	// Base case: reached the end of segments
	if index == len(segments) {
		return node.path
	}

	segment := segments[index]

	// Try exact match first (higher priority)
	if child, exists := node.child[segment]; exists {
		if result := uspr.search(child, segments, index+1); result != nil {
			return result
		}
	}

	// Try path variable match
	if node.childParam != nil {
		if result := uspr.search(node.childParam, segments, index+1); result != nil {
			return result
		}
	}

	return nil
}

// splitPath splits a path into segments, removing empty strings
func splitPath(path string) []string {
	parts := strings.Split(strings.Trim(path, "/"), "/")
	segments := make([]string, 0, len(parts))

	for _, part := range parts {
		if part != "" {
			segments = append(segments, part)
		}
	}

	return segments
}

// isPathVariable checks if a segment is a path variable (e.g., {id})
func isPathVariable(segment string) bool {
	return len(segment) > 2 && segment[0] == '{' && segment[len(segment)-1] == '}'
}

// extractParamName extracts the parameter name from a path variable
// e.g., "{id}" -> "id"
func extractParamName(segment string) string {
	if isPathVariable(segment) {
		return segment[1 : len(segment)-1]
	}
	return ""
}
