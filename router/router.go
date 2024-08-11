package router

import (
	p "path"
	"strings"
)

type HTTPMethod = string

const (
	GET    HTTPMethod = "GET"
	POST   HTTPMethod = "POST"
	PUT    HTTPMethod = "PUT"
	DELETE HTTPMethod = "DELETE"
	PATCH  HTTPMethod = "PATCH"
)

type Router struct {
	root             *Node
	groupBasePath    string
	globalMiddleware []MiddlewareFunc
	groupMiddleware  []MiddlewareFunc
}

type HandlerParams = map[string]string

type MiddlewareFunc func(params HandlerParams, env map[string]interface{}, next HandlerFunc)

type HandlerFunc func(params HandlerParams, env map[string]interface{})

func NewRouter() *Router {
	return &Router{
		root: NewNode(),
	}
}

// Adds any route to the router
func (r *Router) AddRoute(method HTTPMethod, path string, handler HandlerFunc) {
	fullPath := p.Join(r.groupBasePath, path)
	node := r.root

	pathParts := strings.Split(fullPath, "/")

	var sanitizedParts []string

	// Iterate over the path parts and remove empty parts
	for _, part := range pathParts {
		if part != "" {
			sanitizedParts = append(sanitizedParts, part)
		}
	}

	// Traverse the path parts and create nodes as needed
	for _, part := range sanitizedParts {
		// Skip empty parts
		if part == "" {
			continue
		}

		// Handle parameter parts if they exist
		if strings.HasPrefix(part, ":") {
			parameterName := part[1:]
			part = ":param"
			node.parameterName = parameterName
		}

		// Add a new node in the tree if it doesn't exist
		if node.children[part] == nil {
			node.children[part] = NewNode()
		}

		// Link the node to the next part of the path
		node = node.children[part]
	}

	// Mark the node as the end of a route and add the handler
	node.isEnd = true

	// Register the handler for the method, and wrap it with middleware
	node.handlers[method] = r.wrapWithMiddleware(handler)
}

// HTTP Method Helpers

// Adds a GET route to the router
func (r *Router) GET(path string, handler HandlerFunc) {
	r.AddRoute(GET, path, handler)
}

// Adds a POST route to the router
func (r *Router) POST(path string, handler HandlerFunc) {
	r.AddRoute(POST, path, handler)
}

// Adds a PUT route to the router
func (r *Router) PUT(path string, handler HandlerFunc) {
	r.AddRoute(PUT, path, handler)
}

// Adds a DELETE route to the router
func (r *Router) DELETE(path string, handler HandlerFunc) {
	r.AddRoute(DELETE, path, handler)
}

// Adds a PATCH route to the router
func (r *Router) PATCH(path string, handler HandlerFunc) {
	r.AddRoute(PATCH, path, handler)
}

// Miscellaneous Helpers

// Adds a middleware to the router's global middleware
func (r *Router) Use(middleware MiddlewareFunc) {
	r.globalMiddleware = append(r.globalMiddleware, middleware)
}

func (r *Router) Group(basePath string, fn func(group *Group)) {
	originalGroupBasePath := r.groupBasePath
	originalGroupMiddleware := r.groupMiddleware

	r.groupBasePath = p.Join(r.groupBasePath, basePath)
	r.groupMiddleware = []MiddlewareFunc{}

	groupInstance := NewGroup(r, &r.groupMiddleware)
	fn(groupInstance)

	r.groupBasePath = originalGroupBasePath
	r.groupMiddleware = originalGroupMiddleware
}

func (r *Router) Mount(basePath string, otherRouter *Router) {
	originalGroupBasePath := r.groupBasePath
	originalGroupMiddleware := r.groupMiddleware

	r.groupBasePath = p.Join(r.groupBasePath, basePath)
	for _, route := range otherRouter.Routes() {
		method, path, handler := route[0].(string), route[1].(string), route[2].(HandlerFunc)
		r.AddRoute(method, path, handler)
	}

	r.groupBasePath = originalGroupBasePath
	r.groupMiddleware = originalGroupMiddleware
}

func (r *Router) FindHandler(method string, routePath string) (HandlerFunc, HandlerParams) {
	node := r.root
	pathParts := strings.Split(routePath, "/")

	var sanitizedParts []string

	// Iterate over the path parts and remove empty parts
	for _, part := range pathParts {
		if part != "" {
			sanitizedParts = append(sanitizedParts, part)
		}
	}

	params := make(HandlerParams)

	for _, part := range sanitizedParts {
		if part == "" {
			continue
		}
		if child, ok := node.children[part]; ok {
			node = child
		} else if child, ok := node.children[":param"]; ok {
			params[child.parameterName] = part
			node = child
		} else {
			return nil, params
		}
	}

	if node.isEnd {
		if handler, ok := node.handlers[method]; ok {
			return handler, params
		}
	}

	return nil, params
}

func (r *Router) Routes() [][]interface{} {
	return r.collectRoutes(r.root, "")
}

func (r *Router) collectRoutes(node *Node, basePath string) [][]interface{} {
	var routes [][]interface{}
	if node.isEnd {
		for method, handler := range node.handlers {
			routes = append(routes, []interface{}{method, basePath, handler})
		}
	}

	for part, childNode := range node.children {
		nextPath := p.Join(basePath, part)
		routes = append(routes, r.collectRoutes(childNode, nextPath)...)
	}

	return routes
}

func (r *Router) wrapWithMiddleware(handler HandlerFunc) HandlerFunc {
	allMiddleware := append(r.globalMiddleware, r.groupMiddleware...)
	for i := len(allMiddleware) - 1; i >= 0; i-- {
		middleware := allMiddleware[i]
		next := handler
		handler = func(params HandlerParams, env map[string]interface{}) {
			middleware(params, env, next)
		}
	}
	return handler
}
