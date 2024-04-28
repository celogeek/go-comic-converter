package epubimagefilters

import (
	"image"
	"image/color"

	"github.com/disintegration/gift"
)

// AutoCrop Lookup for margin and crop
func AutoCrop(img image.Image, bounds image.Rectangle, cutRatioLeft, cutRatioUp, cutRatioRight, cutRatioBottom int, limit int) gift.Filter {
	return gift.Crop(
		findMargin(img, bounds, cutRatioOptions{cutRatioLeft, cutRatioUp, cutRatioRight, cutRatioBottom}, limit),
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

func findMargin(img image.Image, bounds image.Rectangle, cutRatio cutRatioOptions, limit int) image.Rectangle {
	imgArea := bounds

	maxCutX, maxCutY := imgArea.Dx()*limit/100, imgArea.Dy()*limit/100

LEFT:
	for x, maxCut := imgArea.Min.X, maxCutX; x < imgArea.Max.X && (maxCutX == 0 || maxCut > 0); x, maxCut = x+1, maxCut-1 {
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
	for y, maxCut := imgArea.Min.Y, maxCutY; y < imgArea.Max.Y && (maxCutY == 0 || maxCut > 0); y, maxCut = y+1, maxCut-1 {
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
	for x, maxCut := imgArea.Max.X-1, maxCutX; x >= imgArea.Min.X && (maxCutX == 0 || maxCut > 0); x, maxCut = x-1, maxCut-1 {
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
	for y, maxCut := imgArea.Max.Y-1, maxCutY; y >= imgArea.Min.Y && (maxCutY == 0 || maxCut > 0); y, maxCut = y-1, maxCut-1 {
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

	return imgArea
}
