/*
Templates use to create xml files of the EPUB.
*/
package epubtemplates

import _ "embed"

var (
	//go:embed "epubtemplates_container.xml.tmpl"
	Container string

	//go:embed "epubtemplates_applebooks.xml.tmpl"
	AppleBooks string

	//go:embed "epubtemplates_style.css.tmpl"
	Style string

	//go:embed "epubtemplates_text.xhtml.tmpl"
	Text string

	//go:embed "epubtemplates_blank.xhtml.tmpl"
	Blank string
)
