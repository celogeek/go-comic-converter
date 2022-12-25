package comicconverter

import (
	"image"
	"image/jpeg"
	"os"

	"golang.org/x/image/draw"
)

type ComicConverter struct {
	Options ComicConverterOptions
	img     *image.Gray
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

	switch imgt := img.(type) {
	case *image.Gray:
		c.img = imgt
	default:
		newImg := image.NewGray(img.Bounds())
		draw.Draw(newImg, newImg.Bounds(), img, image.Point{}, draw.Src)
		c.img = newImg
	}

	return c
}

func (c *ComicConverter) isBlank(x, y int) bool {
	r, g, b, _ := c.img.At(x, y).RGBA()
	return r > 60000 && g > 60000 && b > 60000
}

func (c *ComicConverter) CropMarging() *ComicConverter {
	if c.img == nil {
		panic("load image first")
	}

	imgArea := c.img.Bounds()

LEFT:
	for x := imgArea.Min.X; x < imgArea.Max.X; x++ {
		for y := imgArea.Min.Y; y < imgArea.Max.Y; y++ {
			if !c.isBlank(x, y) {
				break LEFT
			}
		}
		imgArea.Min.X++
	}

UP:
	for y := imgArea.Min.Y; y < imgArea.Max.Y; y++ {
		for x := imgArea.Min.X; x < imgArea.Max.X; x++ {
			if !c.isBlank(x, y) {
				break UP
			}
		}
		imgArea.Min.Y++
	}

RIGHT:
	for x := imgArea.Max.X - 1; x >= imgArea.Min.X; x-- {
		for y := imgArea.Min.Y; y < imgArea.Max.Y; y++ {
			if !c.isBlank(x, y) {
				break RIGHT
			}
		}
		imgArea.Max.X--
	}

BOTTOM:
	for y := imgArea.Max.Y - 1; y >= imgArea.Min.Y; y-- {
		for x := imgArea.Min.X; x < imgArea.Max.X; x++ {
			if !c.isBlank(x, y) {
				break BOTTOM
			}
		}
		imgArea.Max.Y--
	}

	newImg := image.NewGray(image.Rectangle{
		Min: image.Point{0, 0},
		Max: image.Point{imgArea.Dx(), imgArea.Dy()},
	})

	draw.Draw(newImg, newImg.Bounds(), c.img, imgArea.Min, draw.Src)

	c.img = c.img.SubImage(imgArea).(*image.Gray)

	return c
}

func (c *ComicConverter) Resize(w, h int) *ComicConverter {
	if c.img == nil {
		panic("load image first")
	}

	dim := c.img.Bounds()
	origWidth := dim.Dx()
	origHeight := dim.Dy()

	width, height := 1, 1

	if origHeight > 0 && origWidth > 0 {
		width, height = origWidth*h/origHeight, origHeight*w/origWidth
	}

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

	draw.BiLinear.Scale(newImg, newImg.Bounds(), c.img, c.img.Bounds(), draw.Src, nil)

	c.img = newImg

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
