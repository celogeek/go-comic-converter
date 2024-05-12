package epubimagefilters

import (
	"image"
	"image/draw"

	"github.com/disintegration/gift"
)

// CropSplitDoublePage Cut a double page in 2 part: left and right.
//
// This will cut in the middle of the page.
func CropSplitDoublePage(right bool) gift.Filter {
	return cropSplitDoublePage{right}
}

type cropSplitDoublePage struct {
	right bool
}

func (p cropSplitDoublePage) Bounds(srcBounds image.Rectangle) (dstBounds image.Rectangle) {
	if p.right {
		dstBounds = image.Rect(
			srcBounds.Max.X/2, srcBounds.Min.Y,
			srcBounds.Max.X, srcBounds.Max.Y,
		)
	} else {
		dstBounds = image.Rect(
			srcBounds.Min.X, srcBounds.Min.Y,
			srcBounds.Max.X/2, srcBounds.Max.Y,
		)
	}
	return
}

func (p cropSplitDoublePage) Draw(dst draw.Image, src image.Image, options *gift.Options) {
	gift.Crop(dst.Bounds()).Draw(dst, src, options)
}
