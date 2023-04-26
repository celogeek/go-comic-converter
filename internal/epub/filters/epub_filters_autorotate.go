/*
Rotate image if the source width > height.
*/
package epubfilters

import (
	"image"
	"image/draw"

	"github.com/disintegration/gift"
)

func AutoRotate() gift.Filter {
	return &autoRotateFilter{}
}

type autoRotateFilter struct {
}

func (p *autoRotateFilter) Bounds(srcBounds image.Rectangle) (dstBounds image.Rectangle) {
	if srcBounds.Dx() > srcBounds.Dy() {
		dstBounds = gift.Rotate90().Bounds(srcBounds)
	} else {
		dstBounds = srcBounds
	}
	return
}

func (p *autoRotateFilter) Draw(dst draw.Image, src image.Image, options *gift.Options) {
	if src.Bounds().Dx() > src.Bounds().Dy() {
		gift.Rotate90().Draw(dst, src, options)
	} else {
		draw.Draw(dst, dst.Bounds(), src, src.Bounds().Min, draw.Src)
	}
}
