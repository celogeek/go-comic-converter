package comicconverter

import (
	"bytes"
	"image"
	"image/color"
	"image/jpeg"
	"io"
	"sort"

	"golang.org/x/image/draw"
)

var AlgoGray = map[string]func(color.Color) color.Color{
	"default": func(c color.Color) color.Color {
		return color.GrayModel.Convert(c)
	},
	"mean": func(c color.Color) color.Color {
		r, g, b, _ := c.RGBA()
		y := float64(r+g+b) / 3 * (255.0 / 65535)
		return color.Gray{uint8(y)}
	},
	"luma": func(c color.Color) color.Color {
		r, g, b, _ := c.RGBA()
		y := (0.2126*float64(r) + 0.7152*float64(g) + 0.0722*float64(b)) * (255.0 / 65535)
		return color.Gray{uint8(y)}
	},
	"luster": func(c color.Color) color.Color {
		r, g, b, _ := c.RGBA()
		arr := []float64{float64(r), float64(g), float64(b)}
		sort.Float64s(arr)
		y := (arr[0] + arr[2]) / 2 * (255.0 / 65535)
		return color.Gray{uint8(y)}
	},
}

func toGray(img image.Image, algo string) *image.Gray {
	grayImg := image.NewGray(img.Bounds())
	algoConv, ok := AlgoGray[algo]
	if !ok {
		panic("wrong gray algo")
	}

	for y := img.Bounds().Min.Y; y < img.Bounds().Max.Y; y++ {
		for x := img.Bounds().Min.X; x < img.Bounds().Max.X; x++ {
			grayImg.Set(x, y, algoConv(img.At(x, y)))
		}
	}
	return grayImg
}

func Load(reader io.ReadCloser, algo string) *image.Gray {
	defer reader.Close()
	img, _, err := image.Decode(reader)
	if err != nil {
		panic(err)
	}

	switch imgt := img.(type) {
	case *image.Gray:
		return imgt
	default:
		return toGray(img, algo)
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

func Get(img *image.Gray, quality int) []byte {
	b := bytes.NewBuffer([]byte{})
	err := jpeg.Encode(b, img, &jpeg.Options{Quality: quality})
	if err != nil {
		panic(err)
	}
	return b.Bytes()
}

func Convert(reader io.ReadCloser, crop bool, w, h int, quality int, algo string) ([]byte, int, int) {
	img := Load(reader, algo)
	if crop {
		img = CropMarging(img)
	}
	img = Resize(img, w, h)
	return Get(img, quality), img.Bounds().Dx(), img.Bounds().Dy()
}
