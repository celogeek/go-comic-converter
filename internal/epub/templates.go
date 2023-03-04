package epub

import _ "embed"

//go:embed "templates/container.xml.tmpl"
var containerTmpl string

//go:embed "templates/content.opf.tmpl"
var contentTmpl string

//go:embed "templates/toc.ncx.tmpl"
var tocTmpl string

//go:embed "templates/nav.xhtml.tmpl"
var navTmpl string

//go:embed "templates/style.css.tmpl"
var styleTmpl string

//go:embed "templates/panelview.css.tmpl"
var panelViewTmpl string

//go:embed "templates/part.xhtml.tmpl"
var partTmpl string

//go:embed "templates/text.xhtml.tmpl"
var textTmpl string

//go:embed "templates/textnopanel.xhtml.tmpl"
var textNoPanelTmpl string

//go:embed "templates/blank.xhtml.tmpl"
var blankTmpl string
