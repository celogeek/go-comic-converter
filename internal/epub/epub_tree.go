package epub

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

func NewTree() *Tree {
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

func (n *Node) toString(indent string) string {
	r := strings.Builder{}
	if indent != "" {
		r.WriteString(indent)
		r.WriteString("- ")
		r.WriteString(n.Value)
		r.WriteString("\n")
	}
	indent += "  "
	for _, c := range n.Children {
		r.WriteString(c.toString(indent))
	}
	return r.String()
}

func (e *ePub) getTree(images []*Image, skip_files bool) string {
	t := NewTree()
	for _, img := range images {
		if skip_files {
			t.Add(img.Path)
		} else {
			t.Add(filepath.Join(img.Path, img.Name))
		}
	}
	c := t.Root()
	if skip_files && e.StripFirstDirectoryFromToc && len(c.Children) == 1 {
		c = c.Children[0]
	}

	return c.toString("")
}
