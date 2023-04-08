package epub

import (
	"encoding/xml"
)

type TocTitle struct {
	XMLName xml.Name `xml:"a"`
	Value   string   `xml:",innerxml"`
	Link    string   `xml:"href,attr"`
}

type TocChildren struct {
	XMLName xml.Name `xml:"ol"`
	Tags    []*TocPart
}

func (t *TocChildren) MarshalYAML() (any, error) {
	return t.Tags, nil
}

type TocPart struct {
	XMLName  xml.Name `xml:"li"`
	Title    TocTitle
	Children *TocChildren `xml:",omitempty"`
}

func (t *TocPart) MarshalYAML() (any, error) {
	if t.Children == nil {
		return t.Title.Value, nil
	} else {
		return map[string]any{t.Title.Value: t.Children}, nil
	}
}
