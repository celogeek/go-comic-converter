// Package epubimage EPUBImage helpers to transform image.
package epubimage

import (
	"image"
	"strings"

	"github.com/celogeek/go-comic-converter/v2/internal/pkg/utils"
)

type EPUBImage struct {
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

// SpaceKey key name of the blank page after the image
func (i EPUBImage) SpaceKey() string {
	return "space_" + utils.IntToString(i.Id)
}

// SpacePath path of the blank page
func (i EPUBImage) SpacePath() string {
	return "Text/" + i.SpaceKey() + ".xhtml"
}

// EPUBSpacePath path of the blank page into the EPUB
func (i EPUBImage) EPUBSpacePath() string {
	return "OEBPS/" + i.SpacePath()
}

func (i EPUBImage) PartKey() string {
	return utils.IntToString(i.Id) + "_p" + utils.IntToString(i.Part)
}

// PageKey key for page
func (i EPUBImage) PageKey() string {
	return "page_" + i.PartKey()
}

// PagePath page path linked to the image
func (i EPUBImage) PagePath() string {
	return "Text/" + i.PageKey() + ".xhtml"
}

// EPUBPagePath page path into the EPUB
func (i EPUBImage) EPUBPagePath() string {
	return "OEBPS/" + i.PagePath()
}

// ImgKey key for image
func (i EPUBImage) ImgKey() string {
	return "img_" + i.PartKey()
}

// ImgPath image path
func (i EPUBImage) ImgPath() string {
	return "Images/" + i.ImgKey() + "." + i.Format
}

// EPUBImgPath image path into the EPUB
func (i EPUBImage) EPUBImgPath() string {
	return "OEBPS/" + i.ImgPath()
}

// ImgStyle style to apply to the image.
//
// center by default.
// align to left or right if it's part of the split double page.
func (i EPUBImage) ImgStyle(viewWidth, viewHeight int, align string) string {
	relWidth, relHeight := i.RelSize(viewWidth, viewHeight)
	marginW, marginH := float64(viewWidth-relWidth)/2, float64(viewHeight-relHeight)/2

	style := make([]string, 0, 4)

	style = append(style, "width:"+utils.IntToString(relWidth)+"px")
	style = append(style, "height:"+utils.IntToString(relHeight)+"px")
	style = append(style, "top:"+utils.FloatToString(marginH*100/float64(viewHeight), 2)+"%")
	if align == "" {
		switch i.Position {
		case "rendition:page-spread-left":
			style = append(style, "right:0")
		case "rendition:page-spread-right":
			style = append(style, "left:0")
		default:
			style = append(style, "left:"+utils.FloatToString(marginW*100/float64(viewWidth), 2)+"%")
		}
	} else {
		style = append(style, align)
	}

	return strings.Join(style, "; ")
}

func (i EPUBImage) RelSize(viewWidth, viewHeight int) (relWidth, relHeight int) {
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
