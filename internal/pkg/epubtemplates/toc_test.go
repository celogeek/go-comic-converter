package epubtemplates

import (
	"testing"

	"github.com/beevik/etree"
	"github.com/stretchr/testify/assert"

	"github.com/celogeek/go-comic-converter/v3/internal/pkg/epubimage"
)

func tocImages(paths ...string) []epubimage.EPUBImage {
	var images []epubimage.EPUBImage
	for id, path := range paths {
		images = append(images, epubimage.EPUBImage{Id: id, Path: path, Format: "jpeg"})
	}
	return images
}

type navPoint struct {
	level     int
	title     string
	src       string
	playOrder string
}

func ncxNavPoints(t *testing.T, ncx string) (*etree.Document, []navPoint) {
	doc := etree.NewDocument()
	assert.NoError(t, doc.ReadFromString(ncx))
	var points []navPoint
	var walk func(e *etree.Element, level int)
	walk = func(e *etree.Element, level int) {
		for _, p := range e.SelectElements("navPoint") {
			points = append(points, navPoint{
				level:     level,
				title:     p.FindElement("navLabel/text").Text(),
				src:       p.FindElement("content").SelectAttrValue("src", ""),
				playOrder: p.SelectAttrValue("playOrder", ""),
			})
			walk(p, level+1)
		}
	}
	walk(doc.FindElement("/ncx/navMap"), 1)
	return doc, points
}

// the NCX has the same entries as the navigation, the entries of the same page share their play order.
func TestTocNcx(t *testing.T) {
	toc := Toc{Title: "Book", Images: tocImages("v1/c1", "v1/c1", "v1/c2/p", "v2")}
	doc, points := ncxNavPoints(t, toc.Ncx("1234"))

	assert.Equal(t, "urn:uuid:1234", doc.FindElement("//meta[@name='dtb:uid']").SelectAttrValue("content", ""))
	assert.Equal(t, "3", doc.FindElement("//meta[@name='dtb:depth']").SelectAttrValue("content", ""))
	assert.Equal(t, "Book", doc.FindElement("/ncx/docTitle/text").Text())
	assert.Equal(t, []navPoint{
		{1, "Book", "Text/page_0_p0.xhtml", "1"},
		{1, "v1", "Text/page_0_p0.xhtml", "1"},
		{2, "c1", "Text/page_0_p0.xhtml", "1"},
		{2, "c2", "Text/page_2_p0.xhtml", "3"},
		{3, "p", "Text/page_2_p0.xhtml", "3"},
		{1, "v2", "Text/page_3_p0.xhtml", "4"},
	}, points)

	ids := map[string]bool{}
	for _, p := range doc.FindElements("//navPoint") {
		id := p.SelectAttrValue("id", "")
		assert.False(t, ids[id], "duplicate id %s", id)
		ids[id] = true
	}
}

// the Kindle supports only 2 levels: the deeper entries follow their parent, in the navigation and the NCX.
func TestTocMaxDepth(t *testing.T) {
	toc := Toc{Title: "Book", HasTitlePage: true, StripFirstDirectory: true, MaxDepth: KindleTocMaxDepth, Images: tocImages("s/v1/c1/a", "s/v1/c1/b", "s/v1/c2", "s/v2")}
	doc, points := ncxNavPoints(t, toc.Ncx("1234"))

	assert.Equal(t, "2", doc.FindElement("//meta[@name='dtb:depth']").SelectAttrValue("content", ""))
	assert.Equal(t, []navPoint{
		{1, "Book", "Text/title.xhtml", "1"},
		{1, "v1", "Text/page_0_p0.xhtml", "2"},
		{2, "c1", "Text/page_0_p0.xhtml", "2"},
		{2, "a", "Text/page_0_p0.xhtml", "2"},
		{2, "b", "Text/page_1_p0.xhtml", "3"},
		{2, "c2", "Text/page_2_p0.xhtml", "4"},
		{1, "v2", "Text/page_3_p0.xhtml", "5"},
	}, points)

	nav := etree.NewDocument()
	assert.NoError(t, nav.ReadFromString(toc.Nav()))
	assert.Len(t, nav.FindElements("//nav[@id='toc']/ol/li"), 3)
	assert.Len(t, nav.FindElements("//nav[@id='toc']/ol/li/ol/li"), 4)
	assert.Empty(t, nav.FindElements("//nav[@id='toc']/ol/li/ol/li/ol"))
}

// a page number for each image: the double page and its split parts are the same page, the title page has none.
func TestTocPageList(t *testing.T) {
	images := []epubimage.EPUBImage{
		{Id: 0, Format: "jpeg"},
		{Id: 1, Format: "jpeg", DoublePage: true},
		{Id: 1, Part: 1, Format: "jpeg"},
		{Id: 1, Part: 2, Format: "jpeg"},
		{Id: 2, Part: 1, Format: "jpeg"},
		{Id: 2, Part: 2, Format: "jpeg"},
		{Id: 3, Format: "jpeg"},
	}
	toc := Toc{Title: "Book", HasTitlePage: true, Images: images}
	doc, points := ncxNavPoints(t, toc.Ncx("1234"))

	assert.Equal(t, []navPoint{{1, "Book", "Text/title.xhtml", "1"}}, points)
	assert.Equal(t, "4", doc.FindElement("//meta[@name='dtb:totalPageCount']").SelectAttrValue("content", ""))
	assert.Equal(t, "4", doc.FindElement("//meta[@name='dtb:maxPageNumber']").SelectAttrValue("content", ""))

	var targets [][3]string
	for _, p := range doc.FindElements("/ncx/pageList/pageTarget") {
		targets = append(targets, [3]string{
			p.SelectAttrValue("value", ""),
			p.FindElement("content").SelectAttrValue("src", ""),
			p.SelectAttrValue("playOrder", ""),
		})
	}
	assert.Equal(t, [][3]string{
		{"1", "Text/page_0_p0.xhtml", "2"},
		{"2", "Text/page_1_p0.xhtml", "3"},
		{"3", "Text/page_2_p1.xhtml", "4"},
		{"4", "Text/page_3_p0.xhtml", "5"},
	}, targets)

	nav := etree.NewDocument()
	assert.NoError(t, nav.ReadFromString(toc.Nav()))
	var links [][2]string
	for _, a := range nav.FindElements("//nav[@epub:type='page-list']/ol/li/a") {
		links = append(links, [2]string{a.Text(), a.SelectAttrValue("href", "")})
	}
	assert.Equal(t, [][2]string{
		{"1", "Text/page_0_p0.xhtml"},
		{"2", "Text/page_1_p0.xhtml"},
		{"3", "Text/page_2_p1.xhtml"},
		{"4", "Text/page_3_p0.xhtml"},
	}, links)
}
