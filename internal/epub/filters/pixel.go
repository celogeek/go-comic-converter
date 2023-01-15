package filters

import (
	"image"
	"image/color"
	"image/draw"

	"github.com/disintegration/gift"
)

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

func (p *pixel) Draw(dst draw.Image, src image.Image, options *gift.Options) {
	if dst.Bounds().Dx() == 1 && dst.Bounds().Dy() == 1 {
		dst.Set(0, 0, color.White)
	}
	draw.Draw(dst, dst.Bounds(), src, src.Bounds().Min, draw.Src)
}
