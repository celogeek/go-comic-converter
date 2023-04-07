package epub

import "encoding/xml"

type TocTitle struct {
	XMLName xml.Name `xml:"a"`
	Value   string   `xml:",innerxml"`
	Link    string   `xml:"href,attr"`
}

type TocChildren struct {
	XMLName xml.Name `xml:"ol"`
	Tags    []*TocPart
}

type TocPart struct {
	XMLName  xml.Name `xml:"li"`
	Title    TocTitle
	Children *TocChildren `xml:",omitempty"`
}
