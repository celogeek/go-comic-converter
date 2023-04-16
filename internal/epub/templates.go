package epub

import _ "embed"

//go:embed "templates/container.xml.tmpl"
var containerTmpl string

//go:embed "templates/applebooks.xml.tmpl"
var appleBooksTmpl string

//go:embed "templates/style.css.tmpl"
var styleTmpl string

//go:embed "templates/part.xhtml.tmpl"
var partTmpl string

//go:embed "templates/text.xhtml.tmpl"
var textTmpl string

//go:embed "templates/blank.xhtml.tmpl"
var blankTmpl string
