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

	maxCropX, maxCropY := imgArea.Dx()*limit/100, imgArea.Dy()*limit/100

LEFT:
	for x, maxCrop := imgArea.Min.X, maxCropX; x < imgArea.Max.X && (limit == 0 || maxCrop > 0); x, maxCrop = x+1, maxCrop-1 {
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
		if limit > 0 && maxCrop == 1 && skipIfLimitReached {
			return bounds
		}
	}

UP:
	for y, maxCrop := imgArea.Min.Y, maxCropY; y < imgArea.Max.Y && (limit == 0 || maxCrop > 0); y, maxCrop = y+1, maxCrop-1 {
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
		if limit > 0 && maxCrop == 1 && skipIfLimitReached {
			return bounds
		}
	}

RIGHT:
	for x, maxCrop := imgArea.Max.X-1, maxCropX; x >= imgArea.Min.X && (limit == 0 || maxCrop > 0); x, maxCrop = x-1, maxCrop-1 {
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
		if limit > 0 && maxCrop == 1 && skipIfLimitReached {
			return bounds
		}
	}

BOTTOM:
	for y, maxCrop := imgArea.Max.Y-1, maxCropY; y >= imgArea.Min.Y && (limit == 0 || maxCrop > 0); y, maxCrop = y-1, maxCrop-1 {
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
		if limit > 0 && maxCrop == 1 && skipIfLimitReached {
			return bounds
		}
	}

	return imgArea
}
