package router

type Group struct {
	router          *Router
	groupMiddleware *[]MiddlewareFunc
}

func NewGroup(router *Router, groupMiddleware *[]MiddlewareFunc) *Group {
	return &Group{
		router:          router,
		groupMiddleware: groupMiddleware,
	}
}

func (g *Group) Use(middleware MiddlewareFunc) {
	*g.groupMiddleware = append(*g.groupMiddleware, middleware)
}

func (g *Group) AddRoute(method string, routePath string, handler HandlerFunc) {
	g.router.AddRoute(method, routePath, handler)
}

func (g *Group) Group(basePath string, fn func(group *Group)) {
	g.router.Group(basePath, fn)
}

func (g *Group) Mount(basePath string, otherRouter *Router) {
	g.router.Mount(basePath, otherRouter)
}
