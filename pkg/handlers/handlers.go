package handlers

import (
	"container/list"
	"net/http"
)

// MiddlewareType specifies middleware interface
type MiddlewareType func(http.ResponseWriter, *http.Request, func(http.ResponseWriter, *http.Request))

// MiddlewareMux implements multiplexer with middlewares stored in a list
type MiddlewareMux struct {
	http.ServeMux
	middlewares list.List
}

// AppendMiddleware adds specified middleware to the stack
func (mux *MiddlewareMux) AppendMiddleware(middleware func(http.ResponseWriter, *http.Request, func(http.ResponseWriter, *http.Request))) {
	// Append middleware to the end
	mux.middlewares.PushBack(MiddlewareType(middleware))
}

// PrependModule prepends middleware in the front
func (mux *MiddlewareMux) PrependMiddleware(middleware func(http.ResponseWriter, *http.Request, func(http.ResponseWriter, *http.Request))) {
	// Add middleware in the front
	mux.middlewares.PushFront(MiddlewareType(middleware))
}

// Iterate to the next middleware in a stack
func (mux *MiddlewareMux) nextMiddleware(el *list.Element) func(w http.ResponseWriter, req *http.Request) {
	if el != nil {
		return func(w http.ResponseWriter, req *http.Request) {
			el.Value.(MiddlewareType)(w, req, mux.nextMiddleware(el.Next()))
		}
	}
	// Return default multiplexer in the end
	return mux.ServeMux.ServeHTTP
}

// ServeHTTP will serve every request
func (mux *MiddlewareMux) ServeHTTP(w http.ResponseWriter, req *http.Request) {
	mux.nextMiddleware(mux.middlewares.Front())(w, req)
}
