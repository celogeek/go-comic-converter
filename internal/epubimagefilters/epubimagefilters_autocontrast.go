package epubimagefilters

import (
	"image"
	"image/color"
	"image/draw"

	"github.com/disintegration/gift"
)

// Automatically improve contrast
func AutoContrast() *autocontrast {
	return &autocontrast{}
}

type autocontrast struct {
}

// compute the color number between 0 and 1 that hold half of the pixel
func (f *autocontrast) mean(src image.Image) float32 {
	bucket := map[uint32]int{}
	for x := src.Bounds().Min.X; x < src.Bounds().Max.X; x++ {
		for y := src.Bounds().Min.Y; y < src.Bounds().Max.Y; y++ {
			v, _, _, _ := color.GrayModel.Convert(src.At(x, y)).RGBA()
			bucket[v]++
		}
	}

	// calculate color idx
	var colorIdx uint32
	{
		// limit to half of the pixel
		limit := src.Bounds().Dx() * src.Bounds().Dy() / 2
		// loop on all color from 0 to 65536
		for colorIdx = 0; colorIdx < 1<<16; colorIdx++ {
			if limit-bucket[colorIdx] < 0 {
				break
			}
			limit -= bucket[colorIdx]
		}
	}
	// return the color idx between 0 and 1
	return float32(colorIdx) / (1 << 16)
}

// ensure value stay into 0 to 1 bound
func (f *autocontrast) cap(v float32) float32 {
	if v < 0 {
		return 0
	}
	if v > 1 {
		return 1
	}
	return v
}

// power of 2 for float32
func (f *autocontrast) pow2(v float32) float32 {
	return v * v
}

// Draw into the dst after applying the filter
func (f *autocontrast) Draw(dst draw.Image, src image.Image, options *gift.Options) {
	// half of the pixel has this color idx
	colorMean := f.mean(src)

	// if colorMean > 0.5, it means the color is mostly clear
	// in that case we will add a lot more darkness other light
	// compute dark factor
	d := f.pow2(colorMean)
	// compute light factor
	l := f.pow2(1 - colorMean)

	gift.ColorFunc(func(r0, g0, b0, a0 float32) (r float32, g float32, b float32, a float32) {
		// convert to gray color the source RGB
		y := 0.299*r0 + 0.587*g0 + 0.114*b0

		// compute a curve from dark and light factor applying to the color
		c := (1 - d) + (d+l)*y

		// applying the coef
		return f.cap(r0 * c), f.cap(g0 * c), f.cap(b0 * c), a0
	}).Draw(dst, src, options)
}

// Bounds calculates the appropriate bounds of an image after applying the filter.
func (*autocontrast) Bounds(srcBounds image.Rectangle) (dstBounds image.Rectangle) {
	dstBounds = srcBounds
	return
}
