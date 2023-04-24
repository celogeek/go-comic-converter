package epubimage

import (
	"fmt"
	"image"

	epubimagedata "github.com/celogeek/go-comic-converter/v2/internal/epub/imagedata"
)

type Image struct {
	Id         int
	Part       int
	Raw        image.Image
	Data       *epubimagedata.ImageData
	Width      int
	Height     int
	IsCover    bool
	DoublePage bool
	Path       string
	Name       string
}

func (i *Image) Key(prefix string) string {
	return fmt.Sprintf("%s_%d_p%d", prefix, i.Id, i.Part)
}

func (i *Image) SpaceKey(prefix string) string {
	return fmt.Sprintf("%s_%d_sp", prefix, i.Id)
}

func (i *Image) TextPath() string {
	return fmt.Sprintf("Text/%d_p%d.xhtml", i.Id, i.Part)
}

func (i *Image) ImgPath() string {
	return fmt.Sprintf("Images/%d_p%d.jpg", i.Id, i.Part)
}

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

func (i *Image) SpacePath() string {
	return fmt.Sprintf("Text/%d_sp.xhtml", i.Id)
}
