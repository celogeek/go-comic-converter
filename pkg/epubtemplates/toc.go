package epubtemplates

import (
	"path/filepath"
	"strings"

	"github.com/beevik/etree"

	"github.com/celogeek/go-comic-converter/v2/pkg/epubimage"
)

// Toc create toc
//
//goland:noinspection HttpUrlsUsage
func Toc(title string, hasTitle bool, stripFirstDirectoryFromToc bool, images []epubimage.EPUBImage) string {
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
			link.CreateAttr("href", img.PagePath())
			link.CreateText(path)
			paths[currentPath] = t.CreateElement("ol")
		}
	}

	if len(ol.ChildElements()) == 1 && stripFirstDirectoryFromToc {
		ol = ol.FindElement("/li/ol")
	}

	for _, v := range ol.FindElements("//ol") {
		if len(v.ChildElements()) == 0 {
			v.Parent().RemoveChild(v)
		}
	}

	beginning := etree.NewElement("li")
	beginningLink := beginning.CreateElement("a")
	if hasTitle {
		beginningLink.CreateAttr("href", "Text/title.xhtml")
	} else {
		beginningLink.CreateAttr("href", images[0].PagePath())
	}
	beginningLink.CreateText(title)
	ol.InsertChildAt(0, beginning)

	nav.AddChild(ol)

	doc.Indent(2)
	r, _ := doc.WriteToString()
	return r
}
