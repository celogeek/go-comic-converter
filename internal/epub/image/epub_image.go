/*
Image helpers to transform image.
*/
package epubimage

import (
	"fmt"
	"image"
)

type Image struct {
	Id         int
	Part       int
	Raw        image.Image
	Width      int
	Height     int
	IsCover    bool
	IsBlank    bool
	DoublePage bool
	Path       string
	Name       string
	Position   string
}

// key name of the blank plage after the image
func (i *Image) SpaceKey() string {
	return fmt.Sprintf("space_%d", i.Id)
}

// path of the blank page
func (i *Image) SpacePath() string {
	return fmt.Sprintf("Text/%s.xhtml", i.SpaceKey())
}

// path of the blank page into the EPUB
func (i *Image) EPUBSpacePath() string {
	return fmt.Sprintf("OEBPS/%s", i.SpacePath())
}

// key for page
func (i *Image) PageKey() string {
	return fmt.Sprintf("page_%d_p%d", i.Id, i.Part)
}

// page path linked to the image
func (i *Image) PagePath() string {
	return fmt.Sprintf("Text/%s.xhtml", i.PageKey())
}

// page path into the EPUB
func (i *Image) EPUBPagePath() string {
	return fmt.Sprintf("OEBPS/%s", i.PagePath())
}

// key for image
func (i *Image) ImgKey() string {
	return fmt.Sprintf("img_%d_p%d", i.Id, i.Part)
}

// image path
func (i *Image) ImgPath() string {
	return fmt.Sprintf("Images/%s.jpg", i.ImgKey())
}

// image path into the EPUB
func (i *Image) EPUBImgPath() string {
	return fmt.Sprintf("OEBPS/%s", i.ImgPath())
}

// style to apply to the image.
//
// center by default.
// align to left or right if it's part of the splitted double page.
func (i *Image) ImgStyle(viewWidth, viewHeight int, align string) string {
	marginW, marginH := float64(viewWidth-i.Width)/2, float64(viewHeight-i.Height)/2

	if align == "" {
		switch i.Position {
		case "rendition:page-spread-left":
			align = "right:0"
		case "rendition:page-spread-right":
			align = "left:0"
		default:
			align = fmt.Sprintf("left:%.2f%%", marginW*100/float64(viewWidth))
		}
	}

	return fmt.Sprintf(
		"width:%dpx; height:%dpx; top:%.2f%%; %s;",
		i.Width,
		i.Height,
		marginH*100/float64(viewHeight),
		align,
	)
}
