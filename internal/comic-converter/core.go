package comicconverter

import (
	"fmt"
	"image"
	"image/color"
	"image/jpeg"
	"os"

	"golang.org/x/image/draw"
)

type ComicConverter struct {
	Options ComicConverterOptions
	img     image.Image
	grayImg *image.Gray16
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

	c.grayImg = image.NewGray16(c.img.Bounds())

	draw.Draw(c.grayImg, c.grayImg.Bounds(), c.img, image.Point{}, draw.Src)

	return c
}

func (c *ComicConverter) CropMarging() *ComicConverter {
	if c.grayImg == nil {
		panic("grayscale first")
	}

	imgArea := c.grayImg.Bounds()
	colorLimit := color.Gray16{60000}

LEFT:
	for x := imgArea.Min.X; x < imgArea.Max.X; x++ {
		for y := imgArea.Min.Y; y < imgArea.Max.Y; y++ {
			if c.grayImg.Gray16At(x, y).Y < colorLimit.Y {
				break LEFT
			}
		}
		imgArea.Min.X++
	}

UP:
	for y := imgArea.Min.Y; y < imgArea.Max.Y; y++ {
		for x := imgArea.Min.X; x < imgArea.Max.X; x++ {
			if c.grayImg.Gray16At(x, y).Y < colorLimit.Y {
				break UP
			}
		}
		imgArea.Min.Y++
	}

RIGHT:
	for x := imgArea.Max.X - 1; x >= imgArea.Min.X; x-- {
		for y := imgArea.Min.Y; y < imgArea.Max.Y; y++ {
			if c.grayImg.Gray16At(x, y).Y < colorLimit.Y {
				break RIGHT
			}
		}
		imgArea.Max.X--
	}

BOTTOM:
	for y := imgArea.Max.Y - 1; y >= imgArea.Min.Y; y-- {
		for x := imgArea.Min.X; x < imgArea.Max.X; x++ {
			if c.grayImg.Gray16At(x, y).Y < colorLimit.Y {
				break BOTTOM
			}
		}
		imgArea.Max.Y--
	}

	fmt.Println("CROP", imgArea)

	c.grayImg = c.grayImg.SubImage(imgArea).(*image.Gray16)

	return c
}

func (c *ComicConverter) Resize(w, h int) *ComicConverter {
	if c.grayImg == nil {
		panic("grayscale first")
	}

	dim := c.grayImg.Bounds()
	origWidth := dim.Dx()
	origHeight := dim.Dy()

	width, heigth := origWidth*h/origHeight, origHeight*w/origWidth

	fmt.Println("W:", origWidth, width, h)
	fmt.Println("H:", origHeight, heigth, w)
	if width > origWidth {
		width = origWidth
	} else if heigth > origHeight {
		heigth = origHeight
	}

	imgGray := image.NewGray16(image.Rectangle{
		Min: image.Point{0, 0},
		Max: image.Point{width, heigth},
	})

	fmt.Println("RESIZE", imgGray.Bounds())

	draw.BiLinear.Scale(imgGray, imgGray.Bounds(), c.grayImg, c.grayImg.Bounds(), draw.Src, nil)

	c.grayImg = imgGray

	return c
}

func (c *ComicConverter) Save(output string) *ComicConverter {
	if c.grayImg == nil {
		panic("grayscale first")
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

	err = jpeg.Encode(o, c.grayImg, &jpeg.Options{Quality: quality})
	if err != nil {
		panic(err)
	}

	return c
}
