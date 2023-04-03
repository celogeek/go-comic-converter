package epub

import (
	"github.com/celogeek/go-comic-converter/v2/internal/epub/filters"
	"github.com/disintegration/gift"
)

func NewGift(options *ImageOptions) *gift.GIFT {
	g := gift.New()
	g.SetParallelization(false)

	if options.AutoRotate {
		g.Add(filters.AutoRotate(options.ViewWidth, options.ViewHeight))
	}
	if options.Contrast != 0 {
		g.Add(gift.Contrast(float32(options.Contrast)))
	}
	if options.Brightness != 0 {
		g.Add(gift.Brightness(float32(options.Brightness)))
	}
	g.Add(
		filters.Resize(options.ViewWidth, options.ViewHeight, gift.LanczosResampling),
		filters.Pixel(),
	)
	return g
}

func NewGiftSplitDoublePage(options *ImageOptions) []*gift.GIFT {
	gifts := make([]*gift.GIFT, 2)

	rightFirst := options.Manga

	gifts[0] = gift.New(
		filters.CropSplitDoublePage(rightFirst),
	)

	gifts[1] = gift.New(
		filters.CropSplitDoublePage(!rightFirst),
	)

	for _, g := range gifts {
		if options.Contrast != 0 {
			g.Add(gift.Contrast(float32(options.Contrast)))
		}
		if options.Brightness != 0 {
			g.Add(gift.Brightness(float32(options.Brightness)))
		}
		g.Add(
			gift.ResizeToFit(options.ViewWidth, options.ViewHeight, gift.LanczosResampling),
		)
	}

	return gifts
}
