package handle_based_on_tree

import (
	"awesomeProject2/web/context"
	"awesomeProject2/web/handler"
	"regexp"
)

// 优先级
var (
	nodeTypeStatic = 1
	nodeTypeReg    = 2
	nodeTypeParams = 3
	nodeTypeAny    = 4
)

// 匹配规则
var (
	patternAny = "*"
)

type Node struct {
	path       string // 用户查找路径节点
	handleFunc handler.HandleFunc
	children   []*Node
	nodeType   int
	pattern    string // route注册匹配规则
	Match      func(path string, c *context.Context) bool
}

func NewStaticNode(pattern string) *Node {
	return &Node{
		pattern:  pattern,
		nodeType: nodeTypeStatic,
		Match: func(path string, c *context.Context) bool {
			ok := path == pattern
			if ok && c != nil {
				c.MatchRoute = pattern
			}
			return ok
		},
	}
}

func NewAnyNode() *Node {
	return &Node{
		path:     patternAny,
		nodeType: nodeTypeAny,
		pattern:  patternAny,
		Match: func(path string, c *context.Context) bool {
			// path != "*" 防止用户输入*被匹配到；用户输入*（具体路径）不等价于我们的*（wildPath）
			ok := path != "*"
			if ok && c != nil {
				c.MatchRoute = patternAny
			}
			return ok
		},
	}
}

func NewParamNode(pattern string) *Node {
	return &Node{
		nodeType: nodeTypeParams,
		pattern:  pattern,
		Match: func(path string, c *context.Context) bool {
			ok := path != patternAny
			if ok && c != nil {
				//c.PathParams[":userId"] = "123"
				if c.PathParams == nil {
					c.PathParams = make(map[string]string, 1)
				}
				c.PathParams[pattern] = path
				c.MatchRoute = pattern
			}
			return ok
		},
	}
}

func NewRegNode(pattern string) *Node {
	return &Node{
		nodeType: nodeTypeReg,
		path:     pattern,
		Match: func(path string, c *context.Context) bool {
			re := regexp.MustCompile(pattern)
			return re.MatchString(path)
		},
	}
}
