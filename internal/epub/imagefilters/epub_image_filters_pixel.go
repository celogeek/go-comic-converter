package epubimagefilters

import (
	"image"
	"image/color"
	"image/draw"

	"github.com/disintegration/gift"
)

// Pixel Generate a blank pixel 1x1, if the size of the image is 0x0.
//
// An image 0x0 is not a valid image, and failed to read.
func Pixel() gift.Filter {
	return &pixel{}
}

type pixel struct {
}

func (p *pixel) Bounds(srcBounds image.Rectangle) (dstBounds image.Rectangle) {
	if srcBounds.Dx() == 0 || srcBounds.Dy() == 0 {
		dstBounds = image.Rect(0, 0, 1, 1)
	} else {
		dstBounds = srcBounds
	}
	return
}

func (p *pixel) Draw(dst draw.Image, src image.Image, _ *gift.Options) {
	if dst.Bounds().Dx() == 1 && dst.Bounds().Dy() == 1 {
		dst.Set(0, 0, color.White)
		return
	}
	draw.Draw(dst, dst.Bounds(), src, src.Bounds().Min, draw.Src)
}
