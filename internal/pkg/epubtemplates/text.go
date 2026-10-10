package epubtemplates

import _ "embed"

// Image of a page: path relative to the OEBPS directory and its style.
type Image struct {
	Path  string
	Style string
}

// PanelCorner hotspot of the panel view: its key (TL, TR, BL, BR) and the reading order of its zoomed view.
type PanelCorner struct {
	Key     string
	Ordinal int
}

// PanelView overlay of an image page, to zoom on each corner of the image.
type PanelView struct {
	// Style position and size of the image into the page, the hotspots cover it.
	Style string
	// ZoomStyle size of the zoomed image, anchored to the corner of the page.
	ZoomStyle string
	// ZoomImages the files of the image (2 when it is stacked), with their style into the zoomed image.
	ZoomImages []Image
	Corners    []PanelCorner
}

//go:embed "text.xhtml.tmpl"
var Text string
