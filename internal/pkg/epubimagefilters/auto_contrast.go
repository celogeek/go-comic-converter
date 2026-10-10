package epubimagefilters

import (
	"image"
	"image/draw"

	"github.com/disintegration/gift"

	"github.com/celogeek/go-comic-converter/v3/internal/pkg/utils"
)

const (
	// share of the pixels ignored on each side of the histogram,
	// so a few specks or scanner noise don't prevent the stretch
	autoContrastClip = 0.005
	// highest black point (0 to 1): dark pages are already deep enough,
	// pushing blacks further only crushes the shadows
	autoContrastMaxBlack = 0.01
	// minimal spread (0 to 1) of the levels to bother stretching
	autoContrastMinRange = 0.05
)

// AutoContrast Automatically improve contrast
//
// It stretches the luminance levels so the lightest pixels (ignoring outliers)
// become white, and the darkest ones a bit deeper (the black point is capped).
// The luminance is computed with the same grayscale mode as the final image.
func AutoContrast(grayScaleMode int) gift.Filter {
	return autocontrast{grayScaleMode: grayScaleMode}
}

type autocontrast struct {
	grayScaleMode int
}

// levels compute the luminance bounds (between 0 and 1) holding all the pixels
// except the clipped share of darkest and lightest ones.
func (f autocontrast) levels(src image.Image) (lo, hi float32) {
	var hist [256]int
	b := src.Bounds()
	for y := b.Min.Y; y < b.Max.Y; y++ {
		for x := b.Min.X; x < b.Max.X; x++ {
			r, g, bl, _ := src.At(x, y).RGBA()
			l := utils.RGBToGray(f.grayScaleMode, float32(r)/0xffff, float32(g)/0xffff, float32(bl)/0xffff)
			hist[int(l*255+0.5)]++
		}
	}

	limit := int(float64(b.Dx()*b.Dy()) * autoContrastClip)

	acc := 0
	lowIdx := 0
	for i, n := range hist {
		acc += n
		if acc > limit {
			lowIdx = i
			break
		}
	}

	acc = 0
	highIdx := 255
	for i := 255; i >= 0; i-- {
		acc += hist[i]
		if acc > limit {
			highIdx = i
			break
		}
	}

	lo = min(float32(lowIdx)/255, autoContrastMaxBlack)
	return lo, float32(highIdx) / 255
}

// ensure value stay into 0 to 1 bound
func (f autocontrast) cap(v float32) float32 {
	if v < 0 {
		return 0
	}
	if v > 1 {
		return 1
	}
	return v
}

// Draw into the dst after applying the filter
func (f autocontrast) Draw(dst draw.Image, src image.Image, options *gift.Options) {
	lo, hi := f.levels(src)

	// already contrasted, or flat image: nothing to stretch
	if hi-lo < autoContrastMinRange {
		lo, hi = 0, 1
	}

	scale := 1 / (hi - lo)

	// same linear map on every channel to keep the hue
	gift.ColorFunc(func(r0, g0, b0, a0 float32) (r float32, g float32, b float32, a float32) {
		return f.cap((r0 - lo) * scale), f.cap((g0 - lo) * scale), f.cap((b0 - lo) * scale), a0
	}).Draw(dst, src, options)
}

// Bounds calculates the appropriate bounds of an image after applying the filter.
func (autocontrast) Bounds(srcBounds image.Rectangle) (dstBounds image.Rectangle) {
	dstBounds = srcBounds
	return
}
