package unix

import "net/http"

type Route struct {
	Method  string
	Path    string
	Handler http.HandlerFunc
}

// 라우터를 받아서, 넣기 위함.
type Router interface {
	Routes() []Route
}
