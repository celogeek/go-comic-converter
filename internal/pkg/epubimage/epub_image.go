// Package epubimage EPUBImage helpers to transform image.
package epubimage

import (
	"image"
	"strings"

	"github.com/celogeek/go-comic-converter/v3/internal/pkg/utils"
)

// Maximum size of an image displayed as is (SR profile).
//
// A bigger image is stored in 2 halves, stacked in the same page,
// so each half stays close to what the device can handle without downscaling.
const (
	MaxWidth  = 1200
	MaxHeight = 1920
)

type EPUBImage struct {
	Id         int
	Part       int
	Raw        image.Image
	Width      int
	Height     int
	IsBlank    bool
	DoublePage bool
	// Stacked tells the image is stored in 2 files, the top half and the bottom half (see BottomImgPath).
	// Width and Height are the size of the whole image.
	Stacked             bool
	Path                string
	Name                string
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

// NeedStack tells if an image of this size is too big to be stored in one file.
func NeedStack(width, height int) bool {
	return height > 1 && (width > MaxWidth || height > MaxHeight)
}

// TopHeight height of the top half of a stacked image.
func (i EPUBImage) TopHeight() int {
	return i.Height / 2
}

// BottomImgKey key for the bottom half of a stacked image
func (i EPUBImage) BottomImgKey() string {
	return i.ImgKey() + "_b"
}

// BottomImgPath image path of the bottom half of a stacked image
func (i EPUBImage) BottomImgPath() string {
	return "Images/" + i.BottomImgKey() + "." + i.Format
}

// EPUBBottomImgPath image path of the bottom half of a stacked image into the EPUB
func (i EPUBImage) EPUBBottomImgPath() string {
	return "OEBPS/" + i.BottomImgPath()
}

// MediaType of the epub image
func (i EPUBImage) MediaType() string {
	return "image/" + i.Format
}

// ImgStyle style to apply to the image.
//
// Always centered: the same page is shown alone in portrait and in a spread in landscape,
// and Kindle devices handle only top/left reliably.
func (i EPUBImage) ImgStyle(viewWidth, viewHeight int) string {
	relWidth, relHeight := i.RelSize(viewWidth, viewHeight)
	marginW, marginH := float64(viewWidth-relWidth)/2, float64(viewHeight-relHeight)/2

	return imgStyle(viewWidth, viewHeight, relWidth, relHeight, marginW, marginH)
}

// ImgStyles style to apply to each file of the image: the top half then the bottom half
// when the image is stacked, so they are displayed as one image.
func (i EPUBImage) ImgStyles(viewWidth, viewHeight int) []string {
	if !i.Stacked {
		return []string{i.ImgStyle(viewWidth, viewHeight)}
	}

	relWidth, relHeight := i.RelSize(viewWidth, viewHeight)
	marginW, marginH := float64(viewWidth-relWidth)/2, float64(viewHeight-relHeight)/2

	topHeight := int(float64(relHeight)*float64(i.TopHeight())/float64(i.Height) + 0.5)

	return []string{
		imgStyle(viewWidth, viewHeight, relWidth, topHeight, marginW, marginH),
		imgStyle(viewWidth, viewHeight, relWidth, relHeight-topHeight, marginW, marginH+float64(topHeight)),
	}
}

func imgStyle(viewWidth, viewHeight, width, height int, left, top float64) string {
	style := make([]string, 0, 4)

	style = append(style, "width:"+utils.IntToString(width)+"px")
	style = append(style, "height:"+utils.IntToString(height)+"px")
	style = append(style, "top:"+utils.FloatToString(top*100/float64(viewHeight), 2)+"%")
	style = append(style, "left:"+utils.FloatToString(left*100/float64(viewWidth), 2)+"%")

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
