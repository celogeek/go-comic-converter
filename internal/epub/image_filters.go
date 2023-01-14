package epub

import (
	"github.com/celogeek/go-comic-converter/internal/epub/filters"
	"github.com/disintegration/gift"
)

func NewGift(options *ImageOptions) *gift.GIFT {
	g := gift.New()
	g.SetParallelization(false)

	if options.AutoRotate {
		g.Add(filters.AutoRotate())
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
