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
		filters.Position(options.ViewWidth, options.ViewHeight, filters.PositionCenter),
	)
	return g
}

func NewGiftSplitDoublePage(options *ImageOptions) []*gift.GIFT {
	gifts := make([]*gift.GIFT, 2)

	gifts[0] = gift.New(
		filters.CropSplitDoublePage(options.Manga),
	)

	gifts[1] = gift.New(
		filters.CropSplitDoublePage(!options.Manga),
	)

	for i, g := range gifts {
		if options.Contrast != 0 {
			g.Add(gift.Contrast(float32(options.Contrast)))
		}
		if options.Brightness != 0 {
			g.Add(gift.Brightness(float32(options.Brightness)))
		}

		position := filters.PositionLeft
		if (i == 1) == options.Manga {
			position = filters.PositionRight
		}

		g.Add(
			filters.Resize(options.ViewWidth, options.ViewHeight, gift.LanczosResampling),
			filters.Position(options.ViewWidth, options.ViewHeight, position),
		)
	}

	return gifts
}
