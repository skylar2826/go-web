package handle_based_on_tree

import (
	"awesomeProject2/web/context"
	"awesomeProject2/web/custom_error"
	"awesomeProject2/web/handler"
	"log"
	"sort"
	"strings"
)

type BasedOnTree struct {
	tree *Node
}

func (b *BasedOnTree) ServeHTTP(c *context.Context) {
	oriPattern := c.R.URL.Path
	pattern := strings.Trim(oriPattern, "/")
	paths := strings.Split(pattern, "/")
	if paths[0] == "" {
		b.tree.handleFunc(c)
		return
	}
	node := b.tree
	for _, path := range paths {
		if child, ok := b.findMatchChild(node, path, c); !ok {
			err := c.NotFoundJson(oriPattern)
			if err != nil {
				log.Fatal(err)
			}
			return
		} else {
			node = child
		}
	}
	if node.handleFunc == nil {
		err := c.NotFoundJson(oriPattern)
		if err != nil {
			log.Fatal(err)
		}
		return
	}
	node.handleFunc(c)
}

// findMatchChild
// 支持静态路由匹配 /a/b
// 支持通配符匹配 /a/*
//
//	不支持 /a/*/b, 即/*必须在最后
//
// 支持路径参数匹配 /a/:userId /a/1
// 支持正则匹配
// 考虑：支持特有匹配规则
func (b *BasedOnTree) findMatchChild(node *Node, path string, c *context.Context) (*Node, bool) {
	candidates := make([]*Node, 0, 2)
	for _, child := range node.children {
		if child.Match(path, c) {
			candidates = append(candidates, child)
		}
	}
	if len(candidates) == 0 {
		return nil, false
	}
	sort.Slice(candidates, func(i, j int) bool {
		return candidates[i].nodeType > candidates[j].nodeType
	})
	return candidates[0], true
}

func (b *BasedOnTree) createSubTree(root *Node, pattern []string) *Node {
	node := root
	for _, path := range pattern {
		if node.children == nil {
			node.children = make([]*Node, 0)
		}
		var child *Node
		if path == "*" {
			child = NewAnyNode()
		} else if path[0] == ':' {
			child = NewParamNode(path)
		} else if path[0] == '[' && path[len(path)-1] == ']' {
			child = NewRegNode(path)
		} else {
			child = NewStaticNode(path)
		}
		node.children = append(node.children, child)
		node = child
	}
	return node
}

func (b *BasedOnTree) validPattern(pattern string) ([]string, error) {
	pattern = strings.Trim(pattern, "/")
	paths := strings.Split(pattern, "/")

	for i, path := range paths {
		if path == "*" && i != len(paths)-1 {
			return nil, custom_error.ErrorInvalidRouterPattern(pattern)
		}
	}
	return paths, nil
}

func (b *BasedOnTree) Route(method string, pattern string, handlerFunc handler.HandleFunc) {
	paths, err := b.validPattern(pattern)
	if err != nil {
		log.Fatal(err)
		return
	}

	node := b.tree
	if paths[0] == "" {
		node.handleFunc = handlerFunc
		return
	}
	for i, path := range paths {
		if child, ok := b.findMatchChild(node, path, nil); !ok {
			node = b.createSubTree(node, paths[i:])
			break
		} else {
			node = child
		}
	}
	node.handleFunc = handlerFunc
}

func NewHandleBasedOnTree() *BasedOnTree {
	return &BasedOnTree{
		tree: &Node{
			path:     "",
			children: make([]*Node, 0),
		},
	}
}
