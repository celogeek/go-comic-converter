package comicconverter

import (
	"image"
	"image/color"
	"image/jpeg"
	"os"

	"golang.org/x/image/draw"
)

type ComicConverter struct {
	Options ComicConverterOptions
	img     image.Image
}

type ComicConverterOptions struct {
	Quality int
}

func New(opt ComicConverterOptions) *ComicConverter {
	return &ComicConverter{Options: opt}
}

func (c *ComicConverter) Load(file string) *ComicConverter {
	f, err := os.Open(file)
	if err != nil {
		panic(err)
	}
	defer f.Close()

	img, _, err := image.Decode(f)
	if err != nil {
		panic(err)
	}

	c.img = img

	return c
}

func (c *ComicConverter) GrayScale() *ComicConverter {
	if c.img == nil {
		panic("load image first")
	}

	grayImg := image.NewGray16(c.img.Bounds())

	draw.Draw(grayImg, grayImg.Bounds(), c.img, image.Point{}, draw.Src)

	c.img = grayImg

	return c
}

func (c *ComicConverter) CropMarging() *ComicConverter {
	if c.img == nil {
		panic("load image first")
	}

	imgArea := c.img.Bounds()
	colorLimit := uint(color.Gray16{60000}.Y)

LEFT:
	for x := imgArea.Min.X; x < imgArea.Max.X; x++ {
		for y := imgArea.Min.Y; y < imgArea.Max.Y; y++ {
			cc, _, _, _ := color.Gray16Model.Convert(c.img.At(x, y)).RGBA()
			if cc < uint32(colorLimit) {
				break LEFT
			}
		}
		imgArea.Min.X++
	}

UP:
	for y := imgArea.Min.Y; y < imgArea.Max.Y; y++ {
		for x := imgArea.Min.X; x < imgArea.Max.X; x++ {
			cc, _, _, _ := color.Gray16Model.Convert(c.img.At(x, y)).RGBA()
			if cc < uint32(colorLimit) {
				break UP
			}
		}
		imgArea.Min.Y++
	}

RIGHT:
	for x := imgArea.Max.X - 1; x >= imgArea.Min.X; x-- {
		for y := imgArea.Min.Y; y < imgArea.Max.Y; y++ {
			cc, _, _, _ := color.Gray16Model.Convert(c.img.At(x, y)).RGBA()
			if cc < uint32(colorLimit) {
				break RIGHT
			}
		}
		imgArea.Max.X--
	}

BOTTOM:
	for y := imgArea.Max.Y - 1; y >= imgArea.Min.Y; y-- {
		for x := imgArea.Min.X; x < imgArea.Max.X; x++ {
			cc, _, _, _ := color.Gray16Model.Convert(c.img.At(x, y)).RGBA()
			if cc < uint32(colorLimit) {
				break BOTTOM
			}
		}
		imgArea.Max.Y--
	}

	grayImg := image.NewGray16(image.Rectangle{
		Min: image.Point{0, 0},
		Max: image.Point{imgArea.Dx(), imgArea.Dy()},
	})

	draw.Draw(grayImg, grayImg.Bounds(), c.img, imgArea.Min, draw.Src)

	c.img = grayImg

	return c
}

func (c *ComicConverter) Resize(w, h int) *ComicConverter {
	if c.img == nil {
		panic("load image first")
	}

	dim := c.img.Bounds()
	origWidth := dim.Dx()
	origHeight := dim.Dy()

	width, heigth := origWidth*h/origHeight, origHeight*w/origWidth

	if width > w {
		width = w
	}
	if heigth > h {
		heigth = h
	}

	imgGray := image.NewGray16(image.Rectangle{
		Min: image.Point{0, 0},
		Max: image.Point{width, heigth},
	})

	draw.BiLinear.Scale(imgGray, imgGray.Bounds(), c.img, c.img.Bounds(), draw.Src, nil)

	c.img = imgGray

	return c
}

func (c *ComicConverter) Save(output string) *ComicConverter {
	if c.img == nil {
		panic("load image first")
	}
	o, err := os.Create(output)
	if err != nil {
		panic(err)
	}
	defer o.Close()

	quality := 75
	if c.Options.Quality > 0 {
		quality = c.Options.Quality
	}

	err = jpeg.Encode(o, c.img, &jpeg.Options{Quality: quality})
	if err != nil {
		panic(err)
	}

	return c
}
