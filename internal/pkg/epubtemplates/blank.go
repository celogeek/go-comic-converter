// Package epubtemplates Templates use to create xml files of the EPUB.
package epubtemplates

import _ "embed"

//go:embed "blank.xhtml.tmpl"
var Blank string
