package filters

import (
	"image"
	"image/draw"

	"github.com/disintegration/gift"
)

func AutoRotate(viewWidth, viewHeight int) gift.Filter {
	return &autoRotateFilter{
		viewWidth, viewHeight,
	}
}

type autoRotateFilter struct {
	viewWidth, viewHeight int
}

func (p *autoRotateFilter) needRotate(srcBounds image.Rectangle) bool {
	width, height := srcBounds.Dx(), srcBounds.Dy()
	if width <= height {
		return false
	}
	if width <= p.viewWidth && height <= p.viewHeight {
		return false
	}
	return true
}

func (p *autoRotateFilter) Bounds(srcBounds image.Rectangle) (dstBounds image.Rectangle) {
	if p.needRotate(srcBounds) {
		dstBounds = gift.Rotate90().Bounds(srcBounds)
	} else {
		dstBounds = srcBounds
	}
	return
}

func (p *autoRotateFilter) Draw(dst draw.Image, src image.Image, options *gift.Options) {
	if p.needRotate(src.Bounds()) {
		gift.Rotate90().Draw(dst, src, options)
	} else {
		draw.Draw(dst, dst.Bounds(), src, src.Bounds().Min, draw.Src)
	}
}
