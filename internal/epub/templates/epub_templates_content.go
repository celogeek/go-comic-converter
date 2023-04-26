package epubtemplates

import (
	"fmt"

	"github.com/beevik/etree"
	epubimage "github.com/celogeek/go-comic-converter/v2/internal/epub/image"
)

type ContentOptions struct {
	Title        string
	UID          string
	Author       string
	Publisher    string
	UpdatedAt    string
	ImageOptions *epubimage.Options
	Cover        *epubimage.Image
	Images       []*epubimage.Image
	Current      int
	Total        int
}

type tagAttrs map[string]string

type tag struct {
	name  string
	attrs tagAttrs
	value string
}

// create the content file
func Content(o *ContentOptions) string {
	doc := etree.NewDocument()
	doc.CreateProcInst("xml", `version="1.0" encoding="UTF-8"`)

	pkg := doc.CreateElement("package")
	pkg.CreateAttr("xmlns", "http://www.idpf.org/2007/opf")
	pkg.CreateAttr("unique-identifier", "ean")
	pkg.CreateAttr("version", "3.0")
	pkg.CreateAttr("prefix", "rendition: http://www.idpf.org/vocab/rendition/#")

	addToElement := func(elm *etree.Element, meth func(o *ContentOptions) []tag) {
		for _, p := range meth(o) {
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
	addToElement(metadata, getMeta)

	manifest := pkg.CreateElement("manifest")
	addToElement(manifest, getManifest)

	spine := pkg.CreateElement("spine")
	if o.ImageOptions.Manga {
		spine.CreateAttr("page-progression-direction", "rtl")
	} else {
		spine.CreateAttr("page-progression-direction", "ltr")
	}
	addToElement(spine, getSpine)

	guide := pkg.CreateElement("guide")
	addToElement(guide, getGuide)

	doc.Indent(2)
	r, _ := doc.WriteToString()

	return r
}

// metadata part of the content
func getMeta(o *ContentOptions) []tag {
	metas := []tag{
		{"meta", tagAttrs{"property": "dcterms:modified"}, o.UpdatedAt},
		{"meta", tagAttrs{"property": "rendition:layout"}, "pre-paginated"},
		{"meta", tagAttrs{"property": "rendition:spread"}, "auto"},
		{"meta", tagAttrs{"property": "rendition:orientation"}, "auto"},
		{"meta", tagAttrs{"property": "schema:accessMode"}, "visual"},
		{"meta", tagAttrs{"property": "schema:accessModeSufficient"}, "visual"},
		{"meta", tagAttrs{"property": "schema:accessibilityHazard"}, "noFlashingHazard"},
		{"meta", tagAttrs{"property": "schema:accessibilityHazard"}, "noMotionSimulationHazard"},
		{"meta", tagAttrs{"property": "schema:accessibilityHazard"}, "noSoundHazard"},
		{"meta", tagAttrs{"name": "book-type", "content": "comic"}, ""},
		{"opf:meta", tagAttrs{"name": "fixed-layout", "content": "true"}, ""},
		{"opf:meta", tagAttrs{"name": "original-resolution", "content": fmt.Sprintf("%dx%d", o.ImageOptions.ViewWidth, o.ImageOptions.ViewHeight)}, ""},
		{"dc:title", tagAttrs{}, o.Title},
		{"dc:identifier", tagAttrs{"id": "ean"}, fmt.Sprintf("urn:uuid:%s", o.UID)},
		{"dc:language", tagAttrs{}, "en"},
		{"dc:creator", tagAttrs{}, o.Author},
		{"dc:publisher", tagAttrs{}, o.Publisher},
		{"dc:contributor", tagAttrs{}, "Go Comic Convertor"},
		{"dc:date", tagAttrs{}, o.UpdatedAt},
	}

	if o.ImageOptions.Manga {
		metas = append(metas, tag{"meta", tagAttrs{"name": "primary-writing-mode", "content": "horizontal-rl"}, ""})
	} else {
		metas = append(metas, tag{"meta", tagAttrs{"name": "primary-writing-mode", "content": "horizontal-lr"}, ""})
	}

	if o.Cover != nil {
		metas = append(metas, tag{"meta", tagAttrs{"name": "cover", "content": o.Cover.Key("img")}, ""})
	}

	if o.Total > 1 {
		metas = append(
			metas,
			tag{"meta", tagAttrs{"name": "calibre:series", "content": o.Title}, ""},
			tag{"meta", tagAttrs{"name": "calibre:series_index", "content": fmt.Sprint(o.Current)}, ""},
		)
	}

	return metas
}

func getManifest(o *ContentOptions) []tag {
	itag := func(img *epubimage.Image) tag {
		return tag{"item", tagAttrs{"id": img.Key("img"), "href": img.ImgPath(), "media-type": "image/jpeg"}, ""}
	}
	htag := func(img *epubimage.Image) tag {
		return tag{"item", tagAttrs{"id": img.Key("page"), "href": img.TextPath(), "media-type": "application/xhtml+xml"}, ""}
	}
	stag := func(img *epubimage.Image) tag {
		return tag{"item", tagAttrs{"id": img.SpaceKey("page"), "href": img.SpacePath(), "media-type": "application/xhtml+xml"}, ""}
	}
	items := []tag{
		{"item", tagAttrs{"id": "toc", "href": "toc.xhtml", "properties": "nav", "media-type": "application/xhtml+xml"}, ""},
		{"item", tagAttrs{"id": "css", "href": "Text/style.css", "media-type": "text/css"}, ""},
		{"item", tagAttrs{"id": "page_title", "href": "Text/title.xhtml", "media-type": "application/xhtml+xml"}, ""},
		{"item", tagAttrs{"id": "img_title", "href": "Images/title.jpg", "media-type": "image/jpeg"}, ""},
	}

	if o.ImageOptions.HasCover || o.Current > 1 {
		items = append(items, itag(o.Cover), htag(o.Cover))
	}

	for _, img := range o.Images {
		if img.Part == 1 {
			items = append(items, stag(img))
		}
		items = append(items, itag(img), htag(img))
	}
	items = append(items, stag(o.Images[len(o.Images)-1]))

	return items
}

// spine part of the content
func getSpine(o *ContentOptions) []tag {
	isOnTheRight := !o.ImageOptions.Manga
	getSpread := func(doublePageNoBlank bool) string {
		isOnTheRight = !isOnTheRight
		if doublePageNoBlank {
			// Center the double page then start back to comic mode (mange/normal)
			isOnTheRight = !o.ImageOptions.Manga
			return "rendition:page-spread-center"
		}
		if isOnTheRight {
			return "rendition:page-spread-right"
		} else {
			return "rendition:page-spread-left"
		}
	}

	spine := []tag{
		{"itemref", tagAttrs{"idref": "page_title", "properties": getSpread(true)}, ""},
	}
	for _, img := range o.Images {
		spine = append(spine, tag{
			"itemref",
			tagAttrs{"idref": img.Key("page"), "properties": getSpread(img.DoublePage && o.ImageOptions.NoBlankPage)},
			"",
		})
		if img.DoublePage && isOnTheRight && !o.ImageOptions.NoBlankPage {
			spine = append(spine, tag{
				"itemref",
				tagAttrs{"idref": img.SpaceKey("page"), "properties": getSpread(false)},
				"",
			})
		}
	}
	if o.ImageOptions.Manga == isOnTheRight {
		spine = append(spine, tag{
			"itemref",
			tagAttrs{"idref": o.Images[len(o.Images)-1].SpaceKey("page"), "properties": getSpread(false)},
			"",
		})
	}

	return spine
}

// guide part of the content
func getGuide(o *ContentOptions) []tag {
	guide := []tag{}
	if o.Cover != nil {
		guide = append(guide, tag{"reference", tagAttrs{"type": "cover", "title": "cover", "href": o.Cover.TextPath()}, ""})
	}
	guide = append(guide, tag{"reference", tagAttrs{"type": "text", "title": "content", "href": o.Images[0].TextPath()}, ""})
	return guide
}
