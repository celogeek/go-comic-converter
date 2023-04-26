package epubimageprocessing

import (
	"image"
	"image/color"
	"path/filepath"
	"strings"
)

// only accept jpg, png and webp as source file
func isSupportedImage(path string) bool {
	switch strings.ToLower(filepath.Ext(path)) {
	case ".jpg", ".jpeg", ".png", ".webp":
		{
			return true
		}
	}
	return false
}

// check if the color is blank enough
func colorIsBlank(c color.Color) bool {
	g := color.GrayModel.Convert(c).(color.Gray)
	return g.Y >= 0xe0
}

// lookup for margin (blank) around the image
func findMarging(img image.Image, cutRatio int) image.Rectangle {
	imgArea := img.Bounds()

LEFT:
	for x := imgArea.Min.X; x < imgArea.Max.X; x++ {
		allowNonBlank := imgArea.Dy() * cutRatio / 100
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
		allowNonBlank := imgArea.Dx() * cutRatio / 100
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
		allowNonBlank := imgArea.Dy() * cutRatio / 100
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
		allowNonBlank := imgArea.Dx() * cutRatio / 100
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
