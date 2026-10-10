package epubtemplates

import _ "embed"

// Image of a page: path relative to the OEBPS directory and its style.
type Image struct {
	Path  string
	Style string
}

//go:embed "text.xhtml.tmpl"
var Text string
