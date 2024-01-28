/*
Image helpers to transform image.
*/
package epubimage

import (
	"fmt"
	"image"
	"strings"
)

type Image struct {
	Id                  int
	Part                int
	Raw                 image.Image
	Width               int
	Height              int
	IsBlank             bool
	DoublePage          bool
	Path                string
	Name                string
	Position            string
	Format              string
	OriginalAspectRatio float64
	Error               error
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
	return fmt.Sprintf("Images/%s.%s", i.ImgKey(), i.Format)
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
	relWidth, relHeight := i.RelSize(viewWidth, viewHeight)
	marginW, marginH := float64(viewWidth-relWidth)/2, float64(viewHeight-relHeight)/2

	style := []string{}

	style = append(style, fmt.Sprintf("width:%dpx", relWidth))
	style = append(style, fmt.Sprintf("height:%dpx", relHeight))
	style = append(style, fmt.Sprintf("top:%.2f%%", marginH*100/float64(viewHeight)))
	if align == "" {
		switch i.Position {
		case "rendition:page-spread-left":
			style = append(style, "right:0")
		case "rendition:page-spread-right":
			style = append(style, "left:0")
		default:
			style = append(style, fmt.Sprintf("left:%.2f%%", marginW*100/float64(viewWidth)))
		}
	} else {
		style = append(style, align)
	}

	return strings.Join(style, "; ")
}

func (i *Image) RelSize(viewWidth, viewHeight int) (relWidth, relHeight int) {
	w, h := viewWidth, viewHeight
	srcw, srch := i.Width, i.Height

	if w <= 0 || h <= 0 || srcw <= 0 || srch <= 0 {
		return
	}

	wratio := float64(srcw) / float64(w)
	hratio := float64(srch) / float64(h)

	if wratio > hratio {
		relWidth = w
		relHeight = int(float64(srch)/wratio + 0.5)
	} else {
		relHeight = h
		relWidth = int(float64(srcw)/hratio + 0.5)
	}

	return
}
