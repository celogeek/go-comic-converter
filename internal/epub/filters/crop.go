package filters

import (
	"image"
	"image/draw"

	"github.com/disintegration/gift"
)

func CropSplitDoublePage(right bool) *cropSplitDoublePage {
	return &cropSplitDoublePage{right}
}

type cropSplitDoublePage struct {
	right bool
}

func (p *cropSplitDoublePage) Bounds(srcBounds image.Rectangle) (dstBounds image.Rectangle) {
	if p.right {
		dstBounds = image.Rectangle{
			Min: image.Point{srcBounds.Max.X / 2, srcBounds.Min.Y},
			Max: srcBounds.Max,
		}
	} else {
		dstBounds = image.Rectangle{
			Min: srcBounds.Min,
			Max: image.Point{srcBounds.Max.X / 2, srcBounds.Max.Y},
		}
	}
	return
}

func (p *cropSplitDoublePage) Draw(dst draw.Image, src image.Image, options *gift.Options) {
	gift.Crop(dst.Bounds()).Draw(dst, src, options)
}
