package epubtree

import (
	"path/filepath"
	"strings"
)

type tree struct {
	Nodes map[string]*node
}

type node struct {
	Value    string
	Children []*node
}

func New() *tree {
	return &tree{map[string]*node{
		".": {".", []*node{}},
	}}
}

func (n *tree) Root() *node {
	return n.Nodes["."]
}

func (n *tree) Add(filename string) {
	cn := n.Root()
	cp := ""
	for _, p := range strings.Split(filepath.Clean(filename), string(filepath.Separator)) {
		cp = filepath.Join(cp, p)
		if _, ok := n.Nodes[cp]; !ok {
			n.Nodes[cp] = &node{Value: p, Children: []*node{}}
			cn.Children = append(cn.Children, n.Nodes[cp])
		}
		cn = n.Nodes[cp]
	}
}

func (n *node) WriteString(indent string) string {
	r := strings.Builder{}
	if indent != "" {
		r.WriteString(indent)
		r.WriteString("- ")
		r.WriteString(n.Value)
		r.WriteString("\n")
	}
	indent += "  "
	for _, c := range n.Children {
		r.WriteString(c.WriteString(indent))
	}
	return r.String()
}
