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

	addToElement(spine, o.getSpineAuto)

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

		if !o.sideOnlyDoublePage() {
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

// Spaces images followed by a blank page, one per image id.
//
// The blank pages are shared by the manifest, the spine and the writer of the epub,
// so the spine never refers to a page that does not exist (a part can start or end in the middle of a double page).
func (o Content) Spaces() []epubimage.EPUBImage {
	if o.sideOnlyDoublePage() {
		return nil
	}

	need := map[string]bool{}

	// double page, first half of a split double page, and the last image
	lastImage := o.Images[len(o.Images)-1]
	for _, img := range o.Images {
		if img.DoublePage ||
			(!o.ImageOptions.KeepDoublePage() && img.Part == 1) ||
			(img.Part == 0 && lastImage.Part == 0 && img.Id == lastImage.Id) {
			need[img.SpaceKey()] = true
		}
	}

	// the ones needed to put the pages on the right side of the spread
	for _, t := range o.getSpineAuto() {
		if key, ok := t.attrs["idref"]; ok {
			need[key] = true
		}
	}

	spaces := make([]epubimage.EPUBImage, 0, len(need))
	for _, img := range o.Images {
		if key := img.SpaceKey(); need[key] {
			spaces = append(spaces, img)
			need[key] = false
		}
	}

	return spaces
}

// sideOnlyDoublePage the normal pages have no side, only the double pages have one.
//
// Kindle shows the "Preview double page" on every page with the left / right hints.
func (o Content) sideOnlyDoublePage() bool {
	return o.ImageOptions.Device == epuboptions.Kindle
}

// getSpineOnlyDoublePage the normal pages have no side:
// a double page is centered, and the 2 parts of a split double page are facing each other.
func (o Content) getSpineOnlyDoublePage() []tag {
	first, second := "page-spread-left", "page-spread-right"
	if o.ImageOptions.Manga {
		first, second = second, first
	}

	var spine []tag
	if o.HasTitlePage {
		spine = append(spine, tag{"itemref", tagAttrs{"idref": "page_title"}, ""})
	}
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

// spine part of the content
func (o Content) getSpineAuto() []tag {
	if o.sideOnlyDoublePage() {
		return o.getSpineOnlyDoublePage()
	}

	isOnTheRight := !o.ImageOptions.Manga
	if o.ImageOptions.Device == epuboptions.Apple {
		isOnTheRight = !isOnTheRight
	}
	getSpread := func(isDoublePage bool) string {
		isOnTheRight = !isOnTheRight
		if isDoublePage {
			// Center the double page then start back to comic mode (mange/normal)
			isOnTheRight = !o.ImageOptions.Manga
			return "page-spread-center"
		}
		// Kindle needs the EPUB 3 core values, it ignores the spread with the rendition: alias.
		if isOnTheRight {
			return "page-spread-right"
		} else {
			return "page-spread-left"
		}
	}
	getSpreadBlank := func() string {
		return getSpread(false) + " layout-blank"
	}

	var spine []tag
	if o.HasTitlePage {
		if o.ImageOptions.Device != epuboptions.Apple {
			spine = append(spine,
				tag{"itemref", tagAttrs{"idref": "space_title", "properties": getSpreadBlank()}, ""},
				tag{"itemref", tagAttrs{"idref": "page_title", "properties": getSpread(false)}, ""},
			)
		} else {
			spine = append(spine,
				tag{"itemref", tagAttrs{"idref": "page_title", "properties": getSpread(true)}, ""},
			)
		}
	}
	for _, img := range o.Images {
		if (img.DoublePage || img.Part == 1) && o.ImageOptions.Manga == isOnTheRight {
			spine = append(spine, tag{
				"itemref",
				tagAttrs{"idref": img.SpaceKey(), "properties": getSpreadBlank()},
				"",
			})
		}
		spine = append(spine, tag{
			"itemref",
			tagAttrs{"idref": img.PageKey(), "properties": getSpread(img.DoublePage)},
			"",
		})
	}
	if o.ImageOptions.Manga == isOnTheRight {
		spine = append(spine, tag{
			"itemref",
			tagAttrs{"idref": o.Images[len(o.Images)-1].SpaceKey(), "properties": getSpread(false)},
			"",
		})
	}

	return spine
}

// getGuide Section guide of the content
func (o Content) getGuide() []tag {
	return []tag{
		{"reference", tagAttrs{"type": "text", "title": "content", "href": o.Images[0].PagePath()}, ""},
	}
}
