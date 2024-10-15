package epubimagefilters

import (
	"image"
	"image/color"

	"github.com/disintegration/gift"
)

// AutoCrop Lookup for margin and crop
func AutoCrop(img image.Image, bounds image.Rectangle, cutRatioLeft, cutRatioUp, cutRatioRight, cutRatioBottom int, limit int, skipIfLimitReached bool) gift.Filter {
	return gift.Crop(
		findMargin(img, bounds, cutRatioOptions{cutRatioLeft, cutRatioUp, cutRatioRight, cutRatioBottom}, limit, skipIfLimitReached),
	)
}

// check if the color is blank enough
func colorIsBlank(c color.Color) bool {
	g := color.GrayModel.Convert(c).(color.Gray)
	return g.Y >= 0xe0
}

// lookup for margin (blank) around the image
type cutRatioOptions struct {
	Left, Up, Right, Bottom int
}

func findMargin(img image.Image, bounds image.Rectangle, cutRatio cutRatioOptions, limit int, skipIfLimitReached bool) image.Rectangle {
	imgArea := bounds

LEFT:
	for x := imgArea.Min.X; x < imgArea.Max.X; x++ {
		allowNonBlank := imgArea.Dy() * cutRatio.Left / 100
		for y := imgArea.Min.Y; y < imgArea.Max.Y; y++ {
			if !colorIsBlank(img.At(x, y)) {
				allowNonBlank--
				if allowNonBlank <= 0 {
					break LEFT
				}
			}
		}
		imgArea.Min.X++
	}

UP:
	for y := imgArea.Min.Y; y < imgArea.Max.Y; y++ {
		allowNonBlank := imgArea.Dx() * cutRatio.Up / 100
		for x := imgArea.Min.X; x < imgArea.Max.X; x++ {
			if !colorIsBlank(img.At(x, y)) {
				allowNonBlank--
				if allowNonBlank <= 0 {
					break UP
				}
			}
		}
		imgArea.Min.Y++
	}

RIGHT:
	for x := imgArea.Max.X - 1; x >= imgArea.Min.X; x-- {
		allowNonBlank := imgArea.Dy() * cutRatio.Right / 100
		for y := imgArea.Min.Y; y < imgArea.Max.Y; y++ {
			if !colorIsBlank(img.At(x, y)) {
				allowNonBlank--
				if allowNonBlank <= 0 {
					break RIGHT
				}
			}
		}
		imgArea.Max.X--
	}

BOTTOM:
	for y := imgArea.Max.Y - 1; y >= imgArea.Min.Y; y-- {
		allowNonBlank := imgArea.Dx() * cutRatio.Bottom / 100
		for x := imgArea.Min.X; x < imgArea.Max.X; x++ {
			if !colorIsBlank(img.At(x, y)) {
				allowNonBlank--
				if allowNonBlank <= 0 {
					break BOTTOM
				}
			}
		}
		imgArea.Max.Y--
	}

	// no limit or blankImage
	if limit == 0 || imgArea.Dx() == 0 || imgArea.Dy() == 0 {
		return imgArea
	}

	exceedX, exceedY := limitExceed(bounds, imgArea, limit)
	if skipIfLimitReached && (exceedX > 0 || exceedY > 0) {
		return bounds
	}

	imgArea.Min.X, imgArea.Max.X = correctLine(imgArea.Min.X, imgArea.Max.X, bounds.Min.X, bounds.Max.X, exceedX)
	imgArea.Min.Y, imgArea.Max.Y = correctLine(imgArea.Min.Y, imgArea.Max.Y, bounds.Min.Y, bounds.Max.Y, exceedY)

	return imgArea
}

func limitExceed(bounds, newBounds image.Rectangle, limit int) (int, int) {
	return bounds.Dx() - newBounds.Dx() - bounds.Dx()*limit/100, bounds.Dy() - newBounds.Dy() - bounds.Dy()*limit/100
}

func correctLine(min, max, bMin, bMax, exceed int) (int, int) {
	if exceed <= 0 {
		return min, max
	}

	min -= exceed / 2
	max += exceed / 2
	if min < bMin {
		max += bMin - min
		min = bMin
	}
	if max > bMax {
		min -= max - bMax
		max = bMax
	}
	return min, max
}
