package epubimagefilters

import (
	"image"
	"image/draw"

	"github.com/disintegration/gift"
)

// Resize image by keeping aspect ratio.
// This will reduce or enlarge image to fit into the viewWidth and viewHeight.
func Resize(viewWidth, viewHeight int, resampling gift.Resampling) gift.Filter {
	return &resizeFilter{
		viewWidth, viewHeight, resampling,
	}
}

type resizeFilter struct {
	viewWidth, viewHeight int
	resampling            gift.Resampling
}

func (p *resizeFilter) Bounds(srcBounds image.Rectangle) image.Rectangle {
	w, h := p.viewWidth, p.viewHeight
	srcw, srch := srcBounds.Dx(), srcBounds.Dy()

	if w <= 0 || h <= 0 || srcw <= 0 || srch <= 0 {
		return image.Rect(0, 0, 0, 0)
	}

	wratio := float64(srcw) / float64(w)
	hratio := float64(srch) / float64(h)

	var dstw, dsth int
	if wratio > hratio {
		dstw = w
		dsth = int(float64(srch)/wratio + 0.5)
	} else {
		dsth = h
		dstw = int(float64(srcw)/hratio + 0.5)
	}

	return image.Rect(0, 0, dstw, dsth)
}

func (p *resizeFilter) Draw(dst draw.Image, src image.Image, options *gift.Options) {
	gift.Resize(dst.Bounds().Dx(), dst.Bounds().Dy(), p.resampling).Draw(dst, src, options)
}
