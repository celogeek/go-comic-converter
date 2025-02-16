package epubimageprocessor

import (
	"bytes"
	"fmt"
	"image"
	"image/jpeg"
	"image/png"
	"io"
	"io/fs"
	"os"
	"path/filepath"
	"slices"
	"strings"

	"github.com/celogeek/go-comic-converter/v3/internal/pkg/epubimage"
	"github.com/celogeek/go-comic-converter/v3/internal/pkg/epubprogress"
	"github.com/celogeek/go-comic-converter/v3/internal/pkg/epubzip"
)

func (e EPUBImageProcessor) PassThrough() (images []epubimage.EPUBImage, err error) {
	fi, err := os.Stat(e.Input)
	if err != nil {
		return
	}

	if fi.IsDir() {
		return e.passThroughDir()
	} else {
		switch ext := strings.ToLower(filepath.Ext(e.Input)); ext {
		case ".cbz", ".zip":
			return e.passThroughCbz()
		case ".cbr", ".rar":
			return e.passThroughCbr()
		default:
			return nil, fmt.Errorf("unknown file format (%s): support .cbz, .zip, .cbr, .rar", ext)
		}
	}

}

func (e EPUBImageProcessor) passThroughDir() (images []epubimage.EPUBImage, err error) {
	imagesPath := make([]string, 0)

	input := filepath.Clean(e.Input)
	err = filepath.WalkDir(input, func(path string, d fs.DirEntry, err error) error {
		if err != nil {
			return err
		}

		// skip hidden files
		if strings.HasPrefix(filepath.Base(path), ".") {
			return nil
		}

		if d.IsDir() {
			return nil
		}

		if slices.Contains([]string{".jpeg", ".jpg", ".png"}, strings.ToLower(filepath.Ext(path))) {
			imagesPath = append(imagesPath, path)
		}
		return nil
	})

	if err != nil {
		return
	}

	if len(imagesPath) == 0 {
		err = errNoImagesFound
		return
	}

	var imgStorage epubzip.StorageImageWriter
	imgStorage, err = epubzip.NewStorageImageWriter(e.ImgStorage(), e.Image.Format)
	if err != nil {
		return
	}

	// processing
	bar := epubprogress.New(epubprogress.Options{
		Quiet:       e.Quiet,
		Json:        e.Json,
		Max:         len(imagesPath),
		Description: "Copying",
		CurrentJob:  1,
		TotalJob:    2,
	})

	for i, imgPath := range imagesPath {
		var f *os.File
		f, err = os.Open(imgPath)
		if err != nil {
			return
		}

		var uncompressedData []byte
		uncompressedData, err = io.ReadAll(f)
		if err != nil {
			return
		}

		err = f.Close()
		if err != nil {
			return
		}

		p, fn := filepath.Split(imgPath)
		if p == input {
			p = ""
		} else {
			p = p[len(input)+1:]
		}

		var (
			format       string
			decodeConfig func(r io.Reader) (image.Config, error)
			decode       func(r io.Reader) (image.Image, error)
		)

		switch filepath.Ext(fn) {
		case ".png":
			format = "png"
			decodeConfig = png.DecodeConfig
			decode = png.Decode
		case ".jpg", ".jpeg":
			format = "jpeg"
			decodeConfig = jpeg.DecodeConfig
			decode = jpeg.Decode
		}

		var config image.Config
		config, err = decodeConfig(bytes.NewReader(uncompressedData))
		if err != nil {
			return
		}

		var rawImage image.Image
		if i == 0 {
			rawImage, err = decode(bytes.NewReader(uncompressedData))
			if err != nil {
				return
			}
		}

		img := epubimage.EPUBImage{
			Id:                  i,
			Part:                0,
			Raw:                 rawImage,
			Width:               config.Width,
			Height:              config.Height,
			IsBlank:             false,
			DoublePage:          config.Width > config.Height,
			Path:                p,
			Name:                fn,
			Format:              format,
			OriginalAspectRatio: float64(config.Height) / float64(config.Width),
		}

		err = imgStorage.AddRaw(img.EPUBImgPath(), uncompressedData)
		if err != nil {
			return
		}

		images = append(images, img)
		_ = bar.Add(1)
	}

	err = imgStorage.Close()
	if err != nil {
		return
	}

	_ = bar.Close()

	if len(images) == 0 {
		err = errNoImagesFound
	}

	return

}

func (e EPUBImageProcessor) passThroughCbz() (images []epubimage.EPUBImage, err error) {
	images = make([]epubimage.EPUBImage, 0)
	err = errNoImagesFound
	return
}

func (e EPUBImageProcessor) passThroughCbr() (images []epubimage.EPUBImage, err error) {
	images = make([]epubimage.EPUBImage, 0)
	err = errNoImagesFound
	return
}
