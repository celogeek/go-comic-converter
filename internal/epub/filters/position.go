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
	return image.Rect(0, 0, p.viewWidth, p.viewHeight)
}

func (p *positionFilter) Draw(dst draw.Image, src image.Image, options *gift.Options) {
	draw.Draw(dst, dst.Bounds(), image.White, dst.Bounds().Min, draw.Over)

	srcBounds := src.Bounds()
	left, top := (p.viewWidth-srcBounds.Dx())/2, (p.viewHeight-srcBounds.Dy())/2
	if p.align == PositionLeft {
		left = 0
	}

	if p.align == PositionRight {
		left = p.viewWidth - srcBounds.Dx()
	}

	draw.Draw(
		dst,
		image.Rect(
			left,
			top,
			p.viewWidth,
			p.viewHeight,
		),
		src,
		srcBounds.Min,
		draw.Over,
	)
}
