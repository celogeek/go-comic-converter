package epub

import (
	"path/filepath"
	"strings"

	"github.com/beevik/etree"
)

func (e *ePub) getToc(title string, images []*Image) string {
	doc := etree.NewDocument()
	doc.CreateProcInst("xml", `version="1.0" encoding="UTF-8"`)
	doc.CreateDirective("DOCTYPE html")

	html := doc.CreateElement("html")
	html.CreateAttr("xmlns", "http://www.w3.org/1999/xhtml")
	html.CreateAttr("xmlns:epub", "http://www.idpf.org/2007/ops")

	html.CreateElement("head").CreateElement("title").CreateText(title)
	body := html.CreateElement("body")
	nav := body.CreateElement("nav")
	nav.CreateAttr("epub:type", "toc")
	nav.CreateAttr("id", "toc")
	nav.CreateElement("h2").CreateText(title)

	ol := etree.NewElement("ol")
	paths := map[string]*etree.Element{".": ol}
	for _, img := range images {
		currentPath := "."
		for _, path := range strings.Split(img.Path, string(filepath.Separator)) {
			parentPath := currentPath
			currentPath = filepath.Join(currentPath, path)
			if _, ok := paths[currentPath]; ok {
				continue
			}
			t := paths[parentPath].CreateElement("li")
			link := t.CreateElement("a")
			link.CreateAttr("href", img.TextPath())
			link.CreateText(path)
			paths[currentPath] = t
		}
	}

	if len(ol.ChildElements()) == 1 && e.StripFirstDirectoryFromToc {
		ol = ol.ChildElements()[0]
	}

	beginning := etree.NewElement("li")
	beginningLink := beginning.CreateElement("a")
	beginningLink.CreateAttr("href", images[0].TextPath())
	beginningLink.CreateText("Start of the book")
	ol.InsertChildAt(0, beginning)

	nav.AddChild(ol)

	doc.Indent(2)
	r, _ := doc.WriteToString()
	return r
}
