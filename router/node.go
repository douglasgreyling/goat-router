package router

type Node struct {
	children      map[string]*Node
	isEnd         bool
	handlers      map[string]HandlerFunc
	parameterName string
}

func NewNode() *Node {
	return &Node{
		children: make(map[string]*Node),
		handlers: make(map[string]HandlerFunc),
	}
}
