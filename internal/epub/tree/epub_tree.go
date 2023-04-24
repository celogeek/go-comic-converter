package epubtree

import (
	"path/filepath"
	"strings"
)

type Tree struct {
	Nodes map[string]*Node
}

type Node struct {
	Value    string
	Children []*Node
}

func New() *Tree {
	return &Tree{map[string]*Node{
		".": {".", []*Node{}},
	}}
}

func (n *Tree) Root() *Node {
	return n.Nodes["."]
}

func (n *Tree) Add(filename string) {
	cn := n.Root()
	cp := ""
	for _, p := range strings.Split(filepath.Clean(filename), string(filepath.Separator)) {
		cp = filepath.Join(cp, p)
		if _, ok := n.Nodes[cp]; !ok {
			n.Nodes[cp] = &Node{Value: p, Children: []*Node{}}
			cn.Children = append(cn.Children, n.Nodes[cp])
		}
		cn = n.Nodes[cp]
	}
}

func (n *Node) ToString(indent string) string {
	r := strings.Builder{}
	if indent != "" {
		r.WriteString(indent)
		r.WriteString("- ")
		r.WriteString(n.Value)
		r.WriteString("\n")
	}
	indent += "  "
	for _, c := range n.Children {
		r.WriteString(c.ToString(indent))
	}
	return r.String()
}
