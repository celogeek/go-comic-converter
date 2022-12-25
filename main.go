package main

import (
	"fmt"
	"image"
	"image/jpeg"
	"os"

	"golang.org/x/image/draw"
)

type KindleSpec struct {
	Width  int
	Height int
}

func (k *KindleSpec) Bounds() image.Rectangle {
	return image.Rectangle{
		Min: image.Point{0, 0},
		Max: image.Point{k.Width, k.Height},
	}
}

func (k *KindleSpec) ConvertGray16(img image.Image) image.Image {
	r := detectMarging(img)
	width, heigth := r.Dx()*k.Height/r.Dy(), r.Dy()*k.Width/r.Dx()
	if width > k.Width {
		width = k.Width
	} else {
		heigth = k.Height
	}

	imgGray := image.NewGray16(image.Rectangle{
		Min: image.Point{0, 0},
		Max: image.Point{width, heigth},
	})

	draw.BiLinear.Scale(imgGray, imgGray.Bounds(), img, r, draw.Src, nil)

	return imgGray
}

var KindleScribe = &KindleSpec{1860, 2480}

func lineIsBlank(img image.Image, y int) bool {
	for x := 0; x < img.Bounds().Max.X; x++ {
		r, _, _, _ := img.At(x, y).RGBA()
		if r < 0xfff {
			return false
		}
	}
	return true
}

func colIsBlank(img image.Image, x int) bool {
	for y := 0; y < img.Bounds().Max.Y; y++ {
		r, _, _, _ := img.At(x, y).RGBA()
		if r < 0xfff { // allow a light gray, white = 0xffff
			return false
		}
	}
	return true
}

func detectMarging(img image.Image) image.Rectangle {
	rect := img.Bounds()

	var xmin, xmax = rect.Min.X, rect.Max.X - 1
	var ymin, ymax = rect.Min.Y, rect.Max.Y - 1

	for ; ymin < ymax && lineIsBlank(img, ymin); ymin++ {
		rect.Min.Y++
	}
	for ; ymin < ymax && lineIsBlank(img, ymax); ymax-- {
		rect.Max.Y--
	}
	for ; xmin < xmax && colIsBlank(img, xmin); xmin++ {
		rect.Min.X++
	}
	for ; xmin < xmax && colIsBlank(img, xmax); xmax-- {
		rect.Max.X--
	}

	fmt.Println(rect)
	return rect
}

func main() {
	f, err := os.Open("/Users/vincent/Downloads/00001.jpg")
	if err != nil {
		panic(err)
	}
	defer f.Close()
	img, _, err := image.Decode(f)
	if err != nil {
		panic(err)
	}

	imgGray := KindleScribe.ConvertGray16(img)

	o, err := os.Create("/Users/vincent/Downloads/00001_gray.jpg")
	if err != nil {
		panic(err)
	}
	defer o.Close()
	err = jpeg.Encode(o, imgGray, &jpeg.Options{Quality: 90})
	if err != nil {
		panic(err)
	}

}
