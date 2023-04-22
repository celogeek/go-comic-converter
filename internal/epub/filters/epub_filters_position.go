package filters

import (
	"image"
	"image/draw"

	"github.com/disintegration/gift"
)

const (
	PositionCenter = iota
	PositionLeft
	PositionRight
)

func Position(viewWidth, viewHeight int, align int) gift.Filter {
	return &positionFilter{
		viewWidth, viewHeight, align,
	}
}

type positionFilter struct {
	viewWidth, viewHeight, align int
}

func (p *positionFilter) Bounds(srcBounds image.Rectangle) image.Rectangle {
	w, h := p.viewWidth, p.viewHeight
	srcw, srch := srcBounds.Dx(), srcBounds.Dy()

	if w <= 0 || h <= 0 || srcw <= 0 || srch <= 0 {
		return image.Rect(0, 0, 0, 0)
	}

	return image.Rect(0, 0, w, h)
}

func (p *positionFilter) Draw(dst draw.Image, src image.Image, options *gift.Options) {
	if dst.Bounds().Dx() == 0 || dst.Bounds().Dy() == 0 {
		return
	}

	draw.Draw(dst, dst.Bounds(), image.White, dst.Bounds().Min, draw.Over)

	srcBounds := src.Bounds()
	left, top := 0, (dst.Bounds().Dy()-srcBounds.Dy())/2

	if p.align == PositionCenter {
		left = (dst.Bounds().Dx() - srcBounds.Dx()) / 2
	}

	if p.align == PositionRight {
		left = dst.Bounds().Dx() - srcBounds.Dx()
	}

	draw.Draw(
		dst,
		image.Rect(
			left,
			top,
			dst.Bounds().Dx(),
			dst.Bounds().Dy(),
		),
		src,
		srcBounds.Min,
		draw.Over,
	)
}
