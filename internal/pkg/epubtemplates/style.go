package epubtemplates

import _ "embed"

//go:embed "style.css.tmpl"
var Style string

//go:embed "panelview.css"
var PanelViewStyle string
