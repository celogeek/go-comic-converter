package epub

import _ "embed"

//go:embed "templates/container.xml.tmpl"
var TEMPLATE_CONTAINER string

//go:embed "templates/content.opf.tmpl"
var TEMPLATE_CONTENT string

//go:embed "templates/toc.ncx.tmpl"
var TEMPLATE_TOC string

//go:embed "templates/nav.xhtml.tmpl"
var TEMPLATE_NAV string

//go:embed "templates/style.css.tmpl"
var TEMPLATE_STYLE string

//go:embed "templates/part.xhtml.tmpl"
var TEMPLATE_PART string

//go:embed "templates/text.xhtml.tmpl"
var TEMPLATE_TEXT string
