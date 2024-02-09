/*
Package epubtree Organize a list of filename with their path into a tree of directories.

Example:
  - A/B/C/D.jpg
  - A/B/C/E.jpg
  - A/B/F/G.jpg

This is transformed like:

	A
	 B
	  C
	   D.jpg
	   E.jpg
	 F
	  G.jpg
*/
package epubtree

import (
	"path/filepath"
	"strings"
)

type Tree struct {
	nodes map[string]*Node
}

type Node struct {
	value    string
	children []*Node
}

// New initialize tree with a root node
func New() *Tree {
	return &Tree{map[string]*Node{
		".": {".", []*Node{}},
	}}
}

// Root root node
func (n *Tree) Root() *Node {
	return n.nodes["."]
}

// Add the filename to the tree
func (n *Tree) Add(filename string) {
	cn := n.Root()
	cp := ""
	for _, p := range strings.Split(filepath.Clean(filename), string(filepath.Separator)) {
		cp = filepath.Join(cp, p)
		if _, ok := n.nodes[cp]; !ok {
			n.nodes[cp] = &Node{value: p, children: []*Node{}}
			cn.children = append(cn.children, n.nodes[cp])
		}
		cn = n.nodes[cp]
	}
}

func (n *Node) ChildCount() int {
	return len(n.children)
}

func (n *Node) FirstChild() *Node {
	return n.children[0]
}

// WriteString string version of the tree
func (n *Node) WriteString(indent string) string {
	r := strings.Builder{}
	if indent != "" {
		r.WriteString(indent)
		r.WriteString("- ")
		r.WriteString(n.value)
		r.WriteString("\n")
	}
	indent += "  "
	for _, c := range n.children {
		r.WriteString(c.WriteString(indent))
	}
	return r.String()
}
