package comicconverter

import (
	"bytes"
	"image"
	"image/color"
	"image/jpeg"
	"io"
	"os"

	"golang.org/x/image/draw"
)

func Load(file string) *image.Gray {
	f, err := os.Open(file)
	if err != nil {
		panic(err)
	}
	defer f.Close()

	img, _, err := image.Decode(f)
	if err != nil {
		panic(err)
	}

	switch imgt := img.(type) {
	case *image.Gray:
		return imgt
	default:
		newImg := image.NewGray(img.Bounds())
		draw.Draw(newImg, newImg.Bounds(), img, image.Point{}, draw.Src)
		return newImg
	}
}

func isBlank(c color.Color) bool {
	r, g, b, _ := c.RGBA()
	return r > 60000 && g > 60000 && b > 60000
}

func CropMarging(img *image.Gray) *image.Gray {
	imgArea := img.Bounds()

LEFT:
	for x := imgArea.Min.X; x < imgArea.Max.X; x++ {
		for y := imgArea.Min.Y; y < imgArea.Max.Y; y++ {
			if !isBlank(img.At(x, y)) {
				break LEFT
			}
		}
		imgArea.Min.X++
	}

UP:
	for y := imgArea.Min.Y; y < imgArea.Max.Y; y++ {
		for x := imgArea.Min.X; x < imgArea.Max.X; x++ {
			if !isBlank(img.At(x, y)) {
				break UP
			}
		}
		imgArea.Min.Y++
	}

RIGHT:
	for x := imgArea.Max.X - 1; x >= imgArea.Min.X; x-- {
		for y := imgArea.Min.Y; y < imgArea.Max.Y; y++ {
			if !isBlank(img.At(x, y)) {
				break RIGHT
			}
		}
		imgArea.Max.X--
	}

BOTTOM:
	for y := imgArea.Max.Y - 1; y >= imgArea.Min.Y; y-- {
		for x := imgArea.Min.X; x < imgArea.Max.X; x++ {
			if !isBlank(img.At(x, y)) {
				break BOTTOM
			}
		}
		imgArea.Max.Y--
	}

	return img.SubImage(imgArea).(*image.Gray)
}

func Resize(img *image.Gray, w, h int) *image.Gray {
	dim := img.Bounds()
	origWidth := dim.Dx()
	origHeight := dim.Dy()

	if origWidth == 0 || origHeight == 0 {
		newImg := image.NewGray(image.Rectangle{
			image.Point{0, 0},
			image.Point{w, h},
		})
		draw.Draw(newImg, newImg.Bounds(), image.NewUniform(color.White), newImg.Bounds().Min, draw.Src)
		return newImg
	}

	width, height := origWidth*h/origHeight, origHeight*w/origWidth

	if width > w {
		width = w
	}
	if height > h {
		height = h
	}

	newImg := image.NewGray(image.Rectangle{
		Min: image.Point{0, 0},
		Max: image.Point{width, height},
	})

	draw.BiLinear.Scale(newImg, newImg.Bounds(), img, img.Bounds(), draw.Src, nil)

	return newImg
}

func Get(img *image.Gray, quality int) io.Reader {
	b := bytes.NewBuffer([]byte{})
	err := jpeg.Encode(b, img, &jpeg.Options{Quality: quality})
	if err != nil {
		panic(err)
	}
	return b
}

func Save(img *image.Gray, output string, quality int) {
	o, err := os.Create(output)
	if err != nil {
		panic(err)
	}
	defer o.Close()

	if quality == 0 {
		quality = 75
	}

	err = jpeg.Encode(o, img, &jpeg.Options{Quality: quality})
	if err != nil {
		panic(err)
	}
}

func Convert(path string, crop bool, w, h int, quality int) (io.Reader, int, int) {
	img := Load(path)
	if crop {
		img = CropMarging(img)
	}
	img = Resize(img, w, h)
	return Get(img, quality), img.Bounds().Dx(), img.Bounds().Dy()
}
