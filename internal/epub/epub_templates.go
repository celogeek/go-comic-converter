package epub

import _ "embed"

//go:embed "templates/epub_templates_container.xml.tmpl"
var containerTmpl string

//go:embed "templates/epub_templates_applebooks.xml.tmpl"
var appleBooksTmpl string

//go:embed "templates/epub_templates_style.css.tmpl"
var styleTmpl string

//go:embed "templates/epub_templates_text.xhtml.tmpl"
var textTmpl string

//go:embed "templates/epub_templates_blank.xhtml.tmpl"
var blankTmpl string
