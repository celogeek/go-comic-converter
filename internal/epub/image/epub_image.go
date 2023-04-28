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
}

// key name of the blank plage after the image
func (i *Image) SpaceKey() string {
	return fmt.Sprintf("space_%d", i.Id)
}

// path of the blank page
func (i *Image) SpacePath() string {
	return fmt.Sprintf("Text/%s.xhtml", i.SpaceKey())
}

// key for page
func (i *Image) PageKey() string {
	return fmt.Sprintf("page_%d_p%d", i.Id, i.Part)
}

// page path linked to the image
func (i *Image) PagePath() string {
	return fmt.Sprintf("Text/%s.xhtml", i.PageKey())
}

// key for image
func (i *Image) ImgKey() string {
	return fmt.Sprintf("img_%d_p%d", i.Id, i.Part)
}

// image path
func (i *Image) ImgPath() string {
	return fmt.Sprintf("Images/%s.jpg", i.ImgKey())
}

// style to apply to the image.
//
// center by default.
// align to left or right if it's part of the splitted double page.
func (i *Image) ImgStyle(viewWidth, viewHeight int, manga bool) string {
	marginW, marginH := float64(viewWidth-i.Width)/2, float64(viewHeight-i.Height)/2
	left, top := marginW*100/float64(viewWidth), marginH*100/float64(viewHeight)
	var align string
	switch i.Part {
	case 0:
		align = fmt.Sprintf("left:%.2f%%", left)
	case 1:
		if manga {
			align = "left:0"
		} else {
			align = "right:0"
		}
	case 2:
		if manga {
			align = "right:0"
		} else {
			align = "left:0"
		}
	}

	return fmt.Sprintf(
		"width:%dpx; height:%dpx; top:%.2f%%; %s;",
		i.Width,
		i.Height,
		top,
		align,
	)
}
