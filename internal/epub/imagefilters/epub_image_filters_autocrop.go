package epubimagefilters

import (
	"image"
	"image/color"

	"github.com/disintegration/gift"
)

// Lookup for margin and crop
func AutoCrop(img image.Image, bounds image.Rectangle, cutRatioLeft, cutRatioUp, cutRatioRight, cutRatioBottom int) gift.Filter {
	return gift.Crop(
		findMargin(img, bounds, cutRatioOptions{cutRatioLeft, cutRatioUp, cutRatioRight, cutRatioBottom}),
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

func findMargin(img image.Image, bounds image.Rectangle, cutRatio cutRatioOptions) image.Rectangle {
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

	return imgArea
}
