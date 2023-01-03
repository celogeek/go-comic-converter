package epub

import (
	"image"
	"image/draw"

	"github.com/disintegration/gift"
)

func NewGift(options *ImageOptions) *gift.GIFT {
	g := gift.New()
	g.SetParallelization(false)

	if options.AutoRotate {
		g.Add(&autoRotateFilter{})
	}
	if options.Contrast != 0 {
		g.Add(gift.Contrast(float32(options.Contrast)))
	}
	if options.Brightness != 0 {
		g.Add(gift.Brightness(float32(options.Brightness)))
	}
	g.Add(
		gift.ResizeToFit(options.ViewWidth, options.ViewHeight, gift.LanczosResampling),
	)
	return g
}

type autoRotateFilter struct{}

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
