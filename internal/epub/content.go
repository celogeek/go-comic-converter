package epub

import (
	"fmt"

	"github.com/beevik/etree"
)

type Content struct {
	doc *etree.Document
}

type TagAttrs map[string]string

type Tag struct {
	name  string
	attrs TagAttrs
	value string
}

func (e *ePub) getMeta(title string, part *epubPart, currentPart, totalPart int) []Tag {
	metas := []Tag{
		{"meta", TagAttrs{"property": "dcterms:modified"}, e.UpdatedAt},
		{"meta", TagAttrs{"property": "rendition:layout"}, "pre-paginated"},
		{"meta", TagAttrs{"property": "rendition:spread"}, "auto"},
		{"meta", TagAttrs{"property": "rendition:orientation"}, "auto"},
		{"meta", TagAttrs{"property": "ibooks:specified-fonts"}, "true"},
		{"meta", TagAttrs{"property": "schema:accessMode"}, "visual"},
		{"meta", TagAttrs{"property": "schema:accessModeSufficient"}, "visual"},
		{"meta", TagAttrs{"property": "schema:accessibilityHazard"}, "noFlashingHazard"},
		{"meta", TagAttrs{"property": "schema:accessibilityHazard"}, "noMotionSimulationHazard"},
		{"meta", TagAttrs{"property": "schema:accessibilityHazard"}, "noSoundHazard"},
		{"meta", TagAttrs{"name": "book-type", "content": "comic"}, ""},
		{"opf:meta", TagAttrs{"name": "fixed-layout", "content": "true"}, ""},
		{"opf:meta", TagAttrs{"name": "original-resolution", "content": fmt.Sprintf("%dx%d", e.ViewWidth, e.ViewHeight)}, ""},
		{"dc:title", TagAttrs{}, title},
		{"dc:identifier", TagAttrs{"id": "ean"}, fmt.Sprintf("urn:uuid:%s", e.UID)},
		{"dc:language", TagAttrs{}, "en"},
		{"dc:creator", TagAttrs{}, e.Author},
		{"dc:publisher", TagAttrs{}, e.Publisher},
		{"dc:contributor", TagAttrs{}, "Go Comic Convertor"},
		{"dc:date", TagAttrs{}, e.UpdatedAt},
	}

	if e.Manga {
		metas = append(metas, Tag{"meta", TagAttrs{"name": "primary-writing-mode", "content": "horizontal-rl"}, ""})
	} else {
		metas = append(metas, Tag{"meta", TagAttrs{"name": "primary-writing-mode", "content": "horizontal-lr"}, ""})
	}

	if part.Cover != nil {
		metas = append(metas, Tag{"meta", TagAttrs{"name": "cover", "content": part.Cover.Key("img")}, ""})
	}

	if totalPart > 1 {
		metas = append(
			metas,
			Tag{"meta", TagAttrs{"name": "calibre:series", "content": e.Title}, ""},
			Tag{"meta", TagAttrs{"name": "calibre:series_index", "content": fmt.Sprint(currentPart)}, ""},
		)
	}

	return metas
}

func (e *ePub) getManifest(title string, part *epubPart, currentPart, totalPart int) []Tag {
	iTag := func(img *Image) Tag {
		return Tag{"item", TagAttrs{"id": img.Key("img"), "href": img.ImgPath(), "media-type": "image/jpeg"}, ""}
	}
	hTag := func(img *Image) Tag {
		return Tag{"item", TagAttrs{"id": img.Key("page"), "href": img.TextPath(), "media-type": "application/xhtml+xml"}, ""}
	}
	sTag := func(img *Image) Tag {
		return Tag{"item", TagAttrs{"id": img.SpaceKey("page"), "href": img.SpacePath(), "media-type": "application/xhtml+xml"}, ""}
	}
	items := []Tag{
		{"item", TagAttrs{"id": "toc", "href": "toc.xhtml", "properties": "nav", "media-type": "application/xhtml+xml"}, ""},
		{"item", TagAttrs{"id": "css", "href": "Text/style.css", "media-type": "text/css"}, ""},
	}

	if part.Cover != nil {
		items = append(items, iTag(part.Cover), hTag(part.Cover))
	}

	for _, img := range part.Images {
		if img.Part == 1 {
			items = append(items, sTag(img))
		}
		items = append(items, iTag(img), hTag(img))
	}
	items = append(items, sTag(part.Images[len(part.Images)-1]))

	return items
}

func (e *ePub) getSpine(title string, part *epubPart, currentPart, totalPart int) []Tag {
	spine := []Tag{}
	isOnTheRight := !e.Manga
	getSpread := func(doublePageNoBlank bool) string {
		isOnTheRight = !isOnTheRight
		if doublePageNoBlank {
			// Center the double page then start back to comic mode (mange/normal)
			isOnTheRight = !e.Manga
			return "rendition:page-spread-center"
		}
		if isOnTheRight {
			return "rendition:page-spread-right"
		} else {
			return "rendition:page-spread-left"
		}
	}
	for _, img := range part.Images {
		spine = append(spine, Tag{
			"itemref",
			TagAttrs{"idref": img.Key("page"), "properties": getSpread(img.DoublePage && e.NoBlankPage)},
			"",
		})
		if img.DoublePage && isOnTheRight && !e.NoBlankPage {
			spine = append(spine, Tag{
				"itemref",
				TagAttrs{"idref": img.SpaceKey("page"), "properties": getSpread(false)},
				"",
			})
		}
	}
	if e.Manga == isOnTheRight {
		spine = append(spine, Tag{
			"itemref",
			TagAttrs{"idref": part.Images[len(part.Images)-1].SpaceKey("page"), "properties": getSpread(false)},
			"",
		})
	}

	return spine
}

func (e *ePub) getGuide(title string, part *epubPart, currentPart, totalPart int) []Tag {
	guide := []Tag{}
	if part.Cover != nil {
		guide = append(guide, Tag{"reference", TagAttrs{"type": "cover", "title": "cover", "href": part.Cover.TextPath()}, ""})
	}
	guide = append(guide, Tag{"reference", TagAttrs{"type": "text", "title": "content", "href": part.Images[0].TextPath()}, ""})
	return guide
}

func (e *ePub) getContent(title string, part *epubPart, currentPart, totalPart int) *Content {
	doc := etree.NewDocument()
	doc.CreateProcInst("xml", `version="1.0" encoding="UTF-8"`)

	pkg := doc.CreateElement("package")
	pkg.CreateAttr("xmlns", "http://www.idpf.org/2007/opf")
	pkg.CreateAttr("unique-identifier", "ean")
	pkg.CreateAttr("version", "3.0")
	pkg.CreateAttr("prefix", "rendition: http://www.idpf.org/vocab/rendition/# ibooks: http://vocabulary.itunes.apple.com/rdf/ibooks/vocabulary-extensions-1.0/")

	addToElement := func(elm *etree.Element, meth func(title string, part *epubPart, currentPart, totalPart int) []Tag) {
		for _, p := range meth(title, part, currentPart, totalPart) {
			meta := elm.CreateElement(p.name)
			for k, v := range p.attrs {
				meta.CreateAttr(k, v)
			}
			meta.SortAttrs()
			if p.value != "" {
				meta.CreateText(p.value)
			}
		}
	}

	metadata := pkg.CreateElement("metadata")
	metadata.CreateAttr("xmlns:dc", "http://purl.org/dc/elements/1.1/")
	metadata.CreateAttr("xmlns:opf", "http://www.idpf.org/2007/opf")
	addToElement(metadata, e.getMeta)

	manifest := pkg.CreateElement("manifest")
	addToElement(manifest, e.getManifest)

	spine := pkg.CreateElement("spine")
	if e.Manga {
		spine.CreateAttr("page-progression-direction", "rtl")
	} else {
		spine.CreateAttr("page-progression-direction", "ltr")
	}
	addToElement(spine, e.getSpine)

	guide := pkg.CreateElement("guide")
	addToElement(guide, e.getGuide)

	return &Content{
		doc,
	}
}

func (c *Content) String() string {
	c.doc.Indent(2)
	r, _ := c.doc.WriteToString()
	return r
}
