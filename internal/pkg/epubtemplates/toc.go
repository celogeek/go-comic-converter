package epubtemplates

import (
	"path/filepath"
	"strings"

	"github.com/beevik/etree"

	"github.com/celogeek/go-comic-converter/v3/internal/pkg/epubimage"
	"github.com/celogeek/go-comic-converter/v3/internal/pkg/utils"
)

// KindleTocMaxDepth the Kindle supports only 2 levels in the table of contents.
const KindleTocMaxDepth = 2

// Toc the table of contents, from the directories of the images.
//
// It is written twice: the EPUB3 navigation (toc.xhtml), and the NCX (toc.ncx)
// for the readers that only read the NCX (Kindle, Kobo with the Adobe engine).
type Toc struct {
	Title               string
	HasTitlePage        bool
	StripFirstDirectory bool
	// MaxDepth the deeper entries are moved up to this level, 0 means no limit.
	MaxDepth int
	Images   []epubimage.EPUBImage
}

type tocEntry struct {
	title    string
	href     string
	children []*tocEntry
}

// entries the beginning of the book, then a tree of the directories linked to their first page.
func (o Toc) entries() []*tocEntry {
	root := &tocEntry{}
	paths := map[string]*tocEntry{".": root}
	for _, img := range o.Images {
		currentPath := "."
		for _, path := range strings.Split(img.Path, string(filepath.Separator)) {
			parentPath := currentPath
			currentPath = filepath.Join(currentPath, path)
			if _, ok := paths[currentPath]; ok {
				continue
			}
			e := &tocEntry{title: path, href: img.PagePath()}
			paths[parentPath].children = append(paths[parentPath].children, e)
			paths[currentPath] = e
		}
	}

	entries := root.children
	if len(entries) == 1 && o.StripFirstDirectory {
		entries = entries[0].children
	}

	beginning := &tocEntry{title: o.Title, href: o.Images[0].PagePath()}
	if o.HasTitlePage {
		beginning.href = "Text/title.xhtml"
	}

	return limitDepth(append([]*tocEntry{beginning}, entries...), o.MaxDepth)
}

// pages the page list: a number for each image, the parts of a double page are the same page.
// The title page and the blank pages have no number.
func (o Toc) pages() []*tocEntry {
	var pages []*tocEntry
	seen := map[int]bool{}
	for _, img := range o.Images {
		if seen[img.Id] {
			continue
		}
		seen[img.Id] = true
		pages = append(pages, &tocEntry{title: utils.IntToString(len(pages) + 1), href: img.PagePath()})
	}
	return pages
}

// playOrders the play order of the NCX, by reading order: the title page, then the pages.
//
// The entries of the table of contents link the first page of a directory, it is always a page of the list.
func (o Toc) playOrders(pages []*tocEntry) map[string]string {
	var hrefs []string
	if o.HasTitlePage {
		hrefs = append(hrefs, "Text/title.xhtml")
	}
	for _, p := range pages {
		hrefs = append(hrefs, p.href)
	}
	playOrders := map[string]string{}
	for i, href := range hrefs {
		playOrders[href] = utils.IntToString(i + 1)
	}
	return playOrders
}

// limitDepth the entries deeper than maxDepth follow their parent at the last level, in the reading order.
func limitDepth(entries []*tocEntry, maxDepth int) []*tocEntry {
	if maxDepth == 0 {
		return entries
	}
	var r []*tocEntry
	for _, e := range entries {
		if maxDepth == 1 {
			r = append(r, &tocEntry{title: e.title, href: e.href})
			r = append(r, limitDepth(e.children, 1)...)
		} else {
			r = append(r, &tocEntry{title: e.title, href: e.href, children: limitDepth(e.children, maxDepth-1)})
		}
	}
	return r
}

// Nav create the EPUB3 navigation document
//
//goland:noinspection HttpUrlsUsage
func (o Toc) Nav() string {
	doc := etree.NewDocument()
	doc.CreateProcInst("xml", `version="1.0" encoding="UTF-8"`)
	doc.CreateDirective("DOCTYPE html")

	html := doc.CreateElement("html")
	html.CreateAttr("xmlns", "http://www.w3.org/1999/xhtml")
	html.CreateAttr("xmlns:epub", "http://www.idpf.org/2007/ops")

	html.CreateElement("head").CreateElement("title").CreateText(o.Title)
	body := html.CreateElement("body")
	nav := body.CreateElement("nav")
	nav.CreateAttr("epub:type", "toc")
	nav.CreateAttr("id", "toc")
	nav.CreateElement("h2").CreateText(o.Title)

	var add func(parent *etree.Element, entries []*tocEntry)
	add = func(parent *etree.Element, entries []*tocEntry) {
		ol := parent.CreateElement("ol")
		for _, e := range entries {
			li := ol.CreateElement("li")
			link := li.CreateElement("a")
			link.CreateAttr("href", e.href)
			link.CreateText(e.title)
			if len(e.children) > 0 {
				add(li, e.children)
			}
		}
	}
	add(nav, o.entries())

	pageList := body.CreateElement("nav")
	pageList.CreateAttr("epub:type", "page-list")
	pageList.CreateAttr("hidden", "")
	add(pageList, o.pages())

	doc.Indent(2)
	r, _ := doc.WriteToString()
	return r
}

// Ncx create the NCX, the uid is the one of the content.
//
// The entries linked to the same page share their play order, in the table of contents and the page list.
//
//goland:noinspection HttpUrlsUsage
func (o Toc) Ncx(uid string) string {
	doc := etree.NewDocument()
	doc.CreateProcInst("xml", `version="1.0" encoding="UTF-8"`)

	ncx := doc.CreateElement("ncx")
	ncx.CreateAttr("xmlns", "http://www.daisy.org/z3986/2005/ncx/")
	ncx.CreateAttr("version", "2005-1")

	entries := o.entries()
	depth := 0
	var maxDepth func(entries []*tocEntry, level int)
	maxDepth = func(entries []*tocEntry, level int) {
		for _, e := range entries {
			depth = max(depth, level)
			maxDepth(e.children, level+1)
		}
	}
	maxDepth(entries, 1)

	pages := o.pages()
	playOrders := o.playOrders(pages)

	head := ncx.CreateElement("head")
	for _, m := range [][2]string{
		{"dtb:uid", "urn:uuid:" + uid},
		{"dtb:depth", utils.IntToString(depth)},
		{"dtb:totalPageCount", utils.IntToString(len(pages))},
		{"dtb:maxPageNumber", utils.IntToString(len(pages))},
	} {
		meta := head.CreateElement("meta")
		meta.CreateAttr("name", m[0])
		meta.CreateAttr("content", m[1])
	}
	ncx.CreateElement("docTitle").CreateElement("text").CreateText(o.Title)

	id := 0
	var add func(parent *etree.Element, entries []*tocEntry)
	add = func(parent *etree.Element, entries []*tocEntry) {
		for _, e := range entries {
			id++
			navPoint := parent.CreateElement("navPoint")
			navPoint.CreateAttr("id", "nav_"+utils.IntToString(id))
			navPoint.CreateAttr("playOrder", playOrders[e.href])
			navPoint.CreateElement("navLabel").CreateElement("text").CreateText(e.title)
			navPoint.CreateElement("content").CreateAttr("src", e.href)
			add(navPoint, e.children)
		}
	}
	add(ncx.CreateElement("navMap"), entries)

	pageList := ncx.CreateElement("pageList")
	pageList.CreateElement("navLabel").CreateElement("text").CreateText(o.Title)
	for _, p := range pages {
		pageTarget := pageList.CreateElement("pageTarget")
		pageTarget.CreateAttr("id", "page_"+p.title)
		pageTarget.CreateAttr("type", "normal")
		pageTarget.CreateAttr("value", p.title)
		pageTarget.CreateAttr("playOrder", playOrders[p.href])
		pageTarget.CreateElement("navLabel").CreateElement("text").CreateText(p.title)
		pageTarget.CreateElement("content").CreateAttr("src", p.href)
	}

	doc.Indent(2)
	r, _ := doc.WriteToString()
	return r
}
