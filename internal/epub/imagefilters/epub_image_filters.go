package epubimagefilters

import (
	"image"

	epubimage "github.com/celogeek/go-comic-converter/v2/internal/epub/image"
	"github.com/disintegration/gift"
)

// create filter to apply to the source
func NewGift(img image.Image, options *epubimage.Options) *gift.GIFT {
	g := gift.New()
	g.SetParallelization(false)

	if options.Crop {
		g.Add(AutoCrop(
			img,
			options.CropRatioLeft,
			options.CropRatioUp,
			options.CropRatioRight,
			options.CropRatioBottom,
		))
	}
	if options.AutoRotate && img.Bounds().Dx() > img.Bounds().Dy() {
		g.Add(gift.Rotate90())
	}

	if options.Contrast != 0 {
		g.Add(gift.Contrast(float32(options.Contrast)))
	}

	if options.Brightness != 0 {
		g.Add(gift.Brightness(float32(options.Brightness)))
	}

	g.Add(
		Resize(options.ViewWidth, options.ViewHeight, gift.LanczosResampling),
		Pixel(),
	)
	return g
}

// create filters to cut image into 2 equal pieces
func NewGiftSplitDoublePage(options *epubimage.Options) []*gift.GIFT {
	gifts := make([]*gift.GIFT, 2)

	gifts[0] = gift.New(
		CropSplitDoublePage(options.Manga),
	)

	gifts[1] = gift.New(
		CropSplitDoublePage(!options.Manga),
	)

	for _, g := range gifts {
		g.SetParallelization(false)
		if options.Contrast != 0 {
			g.Add(gift.Contrast(float32(options.Contrast)))
		}
		if options.Brightness != 0 {
			g.Add(gift.Brightness(float32(options.Brightness)))
		}

		g.Add(
			Resize(options.ViewWidth, options.ViewHeight, gift.LanczosResampling),
		)
	}

	return gifts
}
