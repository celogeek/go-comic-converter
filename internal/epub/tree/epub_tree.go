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

type tree struct {
	Nodes map[string]*node
}

type node struct {
	Value    string
	Children []*node
}

// New initialize tree with a root node
func New() *tree {
	return &tree{map[string]*node{
		".": {".", []*node{}},
	}}
}

// Root root node
func (n *tree) Root() *node {
	return n.Nodes["."]
}

// Add the filename to the tree
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

// WriteString string version of the tree
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
