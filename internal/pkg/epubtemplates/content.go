package epubtemplates

import (
	"github.com/beevik/etree"

	"github.com/celogeek/go-comic-converter/v3/internal/pkg/epubimage"
	"github.com/celogeek/go-comic-converter/v3/internal/pkg/utils"
	"github.com/celogeek/go-comic-converter/v3/pkg/epuboptions"
)

type Content struct {
	Title        string
	HasTitlePage bool
	UID          string
	Author       string
	Publisher    string
	UpdatedAt    string
	ImageOptions epuboptions.Image
	Cover        epubimage.EPUBImage
	Images       []epubimage.EPUBImage
	Current      int
	Total        int
}

type tagAttrs map[string]string

type tag struct {
	name  string
	attrs tagAttrs
	value string
}

// Get create the content file
//
//goland:noinspection HttpUrlsUsage,HttpUrlsUsage,HttpUrlsUsage,HttpUrlsUsage
func (o Content) String() string {
	doc := etree.NewDocument()
	doc.CreateProcInst("xml", `version="1.0" encoding="UTF-8"`)

	pkg := doc.CreateElement("package")
	pkg.CreateAttr("xmlns", "http://www.idpf.org/2007/opf")
	pkg.CreateAttr("unique-identifier", "ean")
	pkg.CreateAttr("version", "3.0")
	pkg.CreateAttr("prefix", "rendition: http://www.idpf.org/vocab/rendition/#")

	addToElement := func(elm *etree.Element, meth func() []tag) {
		for _, p := range meth() {
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
	addToElement(metadata, o.getMeta)

	manifest := pkg.CreateElement("manifest")
	addToElement(manifest, o.getManifest)

	spine := pkg.CreateElement("spine")
	if o.ImageOptions.Manga {
		spine.CreateAttr("page-progression-direction", "rtl")
	} else {
		spine.CreateAttr("page-progression-direction", "ltr")
	}

	addToElement(spine, o.getSpine)

	guide := pkg.CreateElement("guide")
	addToElement(guide, o.getGuide)

	doc.Indent(2)
	r, _ := doc.WriteToString()

	return r
}

// metadata part of the content
func (o Content) getMeta() []tag {
	metas := []tag{
		{"meta", tagAttrs{"property": "dcterms:modified"}, o.UpdatedAt},
		{"meta", tagAttrs{"property": "schema:accessMode"}, "visual"},
		{"meta", tagAttrs{"property": "schema:accessModeSufficient"}, "visual"},
		{"meta", tagAttrs{"property": "schema:accessibilityHazard"}, "noFlashingHazard"},
		{"meta", tagAttrs{"property": "schema:accessibilityHazard"}, "noMotionSimulationHazard"},
		{"meta", tagAttrs{"property": "schema:accessibilityHazard"}, "noSoundHazard"},
		{"meta", tagAttrs{"name": "book-type", "content": "comic"}, ""},
		{"opf:meta", tagAttrs{"name": "fixed-layout", "content": "true"}, ""},
		{"opf:meta", tagAttrs{"name": "original-resolution", "content": o.ImageOptions.View.Dimension()}, ""},
		{"dc:title", tagAttrs{}, o.Title},
		{"dc:identifier", tagAttrs{"id": "ean"}, "urn:uuid:" + o.UID},
		{"dc:language", tagAttrs{}, "en"},
		{"dc:creator", tagAttrs{}, o.Author},
		{"dc:publisher", tagAttrs{}, o.Publisher},
		{"dc:contributor", tagAttrs{}, "Go Comic Convertor"},
		{"dc:date", tagAttrs{}, o.UpdatedAt},
	}

	metas = append(metas, []tag{
		{"meta", tagAttrs{"property": "rendition:layout"}, "pre-paginated"},
		{"meta", tagAttrs{"property": "rendition:spread"}, "auto"},
		{"meta", tagAttrs{"property": "rendition:orientation"}, "auto"},
	}...)

	if o.ImageOptions.Manga {
		metas = append(metas, tag{"meta", tagAttrs{"name": "primary-writing-mode", "content": "horizontal-rl"}, ""})
	} else {
		metas = append(metas, tag{"meta", tagAttrs{"name": "primary-writing-mode", "content": "horizontal-lr"}, ""})
	}

	metas = append(metas, tag{"meta", tagAttrs{"name": "cover", "content": "img_cover"}, ""})

	if o.Total > 1 {
		metas = append(
			metas,
			tag{"meta", tagAttrs{"name": "calibre:series", "content": o.Title}, ""},
			tag{"meta", tagAttrs{"name": "calibre:series_index", "content": utils.IntToString(o.Current)}, ""},
		)
	}

	return metas
}

// imageItems manifest items of the jpeg tiles of the cover or the title image
func imageItems(key, path string, img epubimage.EPUBImage) []tag {
	items := make([]tag, 0, img.Tiles())
	for idx := range img.Tiles() {
		items = append(items, tag{"item", tagAttrs{
			"id":         epubimage.TileName(key, idx),
			"href":       epubimage.TileName(path, idx) + ".jpeg",
			"media-type": "image/jpeg",
		}, ""})
	}

	return items
}

func (o Content) getManifest() []tag {
	var imageTags, pageTags, spaceTags []tag
	addTag := func(img epubimage.EPUBImage) {
		for idx := range img.Tiles() {
			imageTags = append(imageTags,
				tag{"item", tagAttrs{"id": img.TileImgKey(idx), "href": img.TileImgPath(idx), "media-type": img.MediaType()}, ""},
			)
		}
		pageTags = append(pageTags,
			tag{"item", tagAttrs{"id": img.PageKey(), "href": img.PagePath(), "media-type": "application/xhtml+xml"}, ""},
		)
	}

	items := []tag{
		{"item", tagAttrs{"id": "toc", "href": "toc.xhtml", "properties": "nav", "media-type": "application/xhtml+xml"}, ""},
		{"item", tagAttrs{"id": "css", "href": "Text/style.css", "media-type": "text/css"}, ""},
		// The cover is only the thumbnail of the book, it has no page.
		{"item", tagAttrs{"id": "img_cover", "href": "Images/cover.jpeg", "media-type": "image/jpeg", "properties": "cover-image"}, ""},
	}

	if o.ImageOptions.AddPanelView {
		items = append(items, tag{"item", tagAttrs{"id": "panelview_css", "href": "Text/panelview.css", "media-type": "text/css"}, ""})
	}

	if o.HasTitlePage {
		items = append(items,
			tag{"item", tagAttrs{"id": "page_title", "href": "Text/title.xhtml", "media-type": "application/xhtml+xml"}, ""},
		)
		items = append(items, imageItems("img_title", "Images/title", o.ImageOptions.Tile().Tiled(o.ImageOptions.View.Width, o.ImageOptions.View.Height))...)

		if o.ImageOptions.LandscapeMode() {
			items = append(items, tag{"item", tagAttrs{"id": "space_title", "href": "Text/space_title.xhtml", "media-type": "application/xhtml+xml"}, ""})
		}
	}

	for _, img := range o.Images {
		addTag(img)
	}
	for _, img := range o.Spaces() {
		spaceTags = append(spaceTags,
			tag{"item", tagAttrs{"id": img.SpaceKey(), "href": img.SpacePath(), "media-type": "application/xhtml+xml"}, ""},
		)
	}

	items = append(items, imageTags...)
	items = append(items, pageTags...)
	items = append(items, spaceTags...)

	return items
}

// Spaces images followed by a blank page, one per image id: the blank pages of the spine.
func (o Content) Spaces() []epubimage.EPUBImage {
	need := map[string]bool{}
	for _, t := range o.getSpine() {
		if key, ok := t.attrs["idref"]; ok {
			need[key] = true
		}
	}

	var spaces []epubimage.EPUBImage
	for _, img := range o.Images {
		if key := img.SpaceKey(); need[key] {
			spaces = append(spaces, img)
			need[key] = false
		}
	}

	return spaces
}

// getSpine spine part of the content, the way to spread the pages depends on the device.
func (o Content) getSpine() []tag {
	switch {
	case o.ImageOptions.LandscapeMode():
		return o.getSpineKindleLandscape()
	case o.ImageOptions.Device == epuboptions.Kindle:
		return o.getSpineKindle()
	case o.ImageOptions.Device == epuboptions.Apple:
		return o.getSpineApple()
	default:
		return o.getSpineKobo()
	}
}

// getSpineKindle Kindle is centered by default, and a side on every page shows the "Preview double page" on every page:
// only the double page is centered, and the 2 parts of a split double page are facing each other.
func (o Content) getSpineKindle() []tag {
	first, second := "page-spread-left", "page-spread-right"
	if o.ImageOptions.Manga {
		first, second = second, first
	}

	spine := o.getSpineTitle()
	for _, img := range o.Images {
		attrs := tagAttrs{"idref": img.PageKey()}
		switch {
		case img.DoublePage:
			attrs["properties"] = "page-spread-center"
		case img.Part == 1:
			attrs["properties"] = first
		case img.Part == 2:
			attrs["properties"] = second
		}
		spine = append(spine, tag{"itemref", attrs, ""})
	}

	return spine
}

// getSpineApple Apple Books doesn't support the center, it balances the left and right pages by itself:
// the pages have no hint, and a blank page fills the spread when a double page
// (or the first part of a split) would start on the wrong side.
//
// The first page is on the second side of the spread (left for manga), after the title.
func (o Content) getSpineApple() []tag {
	return o.getSpineFiller(true)
}

// getSpineKobo the same as Apple Books, but the first page is on the first side of the spread (right for manga).
func (o Content) getSpineKobo() []tag {
	return o.getSpineFiller(false)
}

// getSpineFiller the pages have no hint, the reader balances them: a blank page fills the spread
// when a double page (or the first part of a split) would start on the wrong side.
func (o Content) getSpineFiller(firstOnSecondSide bool) []tag {
	spine := o.getSpineTitle()
	onSecondSide := firstOnSecondSide
	for _, img := range o.Images {
		if onSecondSide && (img.DoublePage || img.Part == 1) {
			spine = append(spine, tag{"itemref", tagAttrs{"idref": img.SpaceKey()}, ""})
			onSecondSide = false
		}
		switch {
		case img.DoublePage, img.Part == 2:
			onSecondSide = false
		case img.Part == 1:
			onSecondSide = true
		default:
			onSecondSide = !onSecondSide
		}
		spine = append(spine, tag{"itemref", tagAttrs{"idref": img.PageKey()}, ""})
	}

	return spine
}

// getSpineTitle the title page, alone and without hint.
func (o Content) getSpineTitle() []tag {
	if !o.HasTitlePage {
		return nil
	}
	return []tag{{"itemref", tagAttrs{"idref": "page_title"}, ""}}
}

// getSpineKindleLandscape the pages are spread left and right, alternately, with a blank page when needed:
// the double page is centered, then the pages start back on the first side (right for manga).
// It is a Kindle Scribe only mode, with the blank pages of the Kindle (layout-blank).
func (o Content) getSpineKindleLandscape() []tag {
	isOnTheRight := !o.ImageOptions.Manga
	getSpread := func(isDoublePage bool) string {
		isOnTheRight = !isOnTheRight
		if isDoublePage {
			// Center the double page then start back to comic mode (manga/normal)
			isOnTheRight = !o.ImageOptions.Manga
			return "page-spread-center"
		}
		if isOnTheRight {
			return "page-spread-right"
		}
		return "page-spread-left"
	}
	getSpreadBlank := func() string {
		return getSpread(false) + " layout-blank"
	}

	var spine []tag
	if o.HasTitlePage {
		spine = append(spine,
			tag{"itemref", tagAttrs{"idref": "space_title", "properties": getSpreadBlank()}, ""},
			tag{"itemref", tagAttrs{"idref": "page_title", "properties": getSpread(false)}, ""},
		)
	}
	for _, img := range o.Images {
		if (img.DoublePage || img.Part == 1) && o.ImageOptions.Manga == isOnTheRight {
			spine = append(spine, tag{"itemref", tagAttrs{"idref": img.SpaceKey(), "properties": getSpreadBlank()}, ""})
		}
		spine = append(spine, tag{"itemref", tagAttrs{"idref": img.PageKey(), "properties": getSpread(img.DoublePage)}, ""})
	}
	if o.ImageOptions.Manga == isOnTheRight {
		spine = append(spine, tag{"itemref", tagAttrs{"idref": o.Images[len(o.Images)-1].SpaceKey(), "properties": getSpread(false)}, ""})
	}

	return spine
}

// getGuide Section guide of the content
func (o Content) getGuide() []tag {
	return []tag{
		{"reference", tagAttrs{"type": "text", "title": "content", "href": o.Images[0].PagePath()}, ""},
	}
}
