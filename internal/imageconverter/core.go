package imageconverter

import (
	"bytes"
	"image"
	"image/color"
	"image/jpeg"
	"io"

	"golang.org/x/image/draw"
)

func Load(reader io.ReadCloser, algo string, palette color.Palette) *image.Gray {
	defer reader.Close()
	img, _, err := image.Decode(reader)
	if err != nil {
		panic(err)
	}
	algoFunc, ok := ALGO_GRAY[algo]
	if !ok {
		panic("unknown algo")
	}

	grayImg := image.NewGray(img.Bounds())
	for y := img.Bounds().Min.Y; y < img.Bounds().Max.Y; y++ {
		for x := img.Bounds().Min.X; x < img.Bounds().Max.X; x++ {
			grayImg.SetGray(x, y, algoFunc(img.At(x, y), palette))
		}
	}

	return grayImg
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

func Get(img *image.Gray, quality int) []byte {
	b := bytes.NewBuffer([]byte{})
	err := jpeg.Encode(b, img, &jpeg.Options{Quality: quality})
	if err != nil {
		panic(err)
	}
	return b.Bytes()
}

func Convert(reader io.ReadCloser, crop bool, w, h int, quality int, algo string, palette color.Palette) ([]byte, int, int) {
	img := Load(reader, algo, palette)
	if crop {
		img = CropMarging(img)
	}
	img = Resize(img, w, h)
	return Get(img, quality), img.Bounds().Dx(), img.Bounds().Dy()
}