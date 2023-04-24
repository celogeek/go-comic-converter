package epubtemplates

import _ "embed"

var (
	//go:embed "epub_templates_container.xml.tmpl"
	Container string

	//go:embed "epub_templates_applebooks.xml.tmpl"
	AppleBooks string

	//go:embed "epub_templates_style.css.tmpl"
	Style string

	//go:embed "epub_templates_text.xhtml.tmpl"
	Text string

	//go:embed "epub_templates_blank.xhtml.tmpl"
	Blank string
)
