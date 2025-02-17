package epubimagepassthrough

import (
	"archive/zip"
	"bytes"
	"errors"
	"fmt"
	"image"
	"image/jpeg"
	"image/png"
	"io"
	"io/fs"
	"os"
	"path/filepath"
	"slices"
	"sort"
	"strings"

	"github.com/nwaples/rardecode/v2"

	"github.com/celogeek/go-comic-converter/v3/internal/pkg/epubimage"
	"github.com/celogeek/go-comic-converter/v3/internal/pkg/epubimageprocessor"
	"github.com/celogeek/go-comic-converter/v3/internal/pkg/epubprogress"
	"github.com/celogeek/go-comic-converter/v3/internal/pkg/epubzip"
	"github.com/celogeek/go-comic-converter/v3/internal/pkg/sortpath"
	"github.com/celogeek/go-comic-converter/v3/pkg/epuboptions"
)

type ePUBImagePassthrough struct {
	epuboptions.EPUBOptions
}

func (e ePUBImagePassthrough) Load() (images []epubimage.EPUBImage, err error) {
	fi, err := os.Stat(e.Input)
	if err != nil {
		return
	}

	if fi.IsDir() {
		return e.loadDir()
	} else {
		switch ext := strings.ToLower(filepath.Ext(e.Input)); ext {
		case ".cbz", ".zip":
			return e.loadCbz()
		case ".cbr", ".rar":
			return e.loadCbr()
		default:
			return nil, fmt.Errorf("unknown file format (%s): support .cbz, .zip, .cbr, .rar", ext)
		}
	}
}

func (e ePUBImagePassthrough) CoverTitleData(o epubimageprocessor.CoverTitleDataOptions) (epubzip.Image, error) {
	return epubimageprocessor.New(e.EPUBOptions).CoverTitleData(o)
}

var errNoImagesFound = errors.New("no images found")

func New(o epuboptions.EPUBOptions) epubimageprocessor.EPUBImageProcessor {
	return ePUBImagePassthrough{o}
}

func (e ePUBImagePassthrough) loadDir() (images []epubimage.EPUBImage, err error) {
	imagesPath := make([]string, 0)

	input := filepath.Clean(e.Input)
	err = filepath.WalkDir(input, func(path string, d fs.DirEntry, err error) error {
		if err != nil {
			return err
		}

		if filterCopyPath(d.IsDir(), path) {
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

	sort.Sort(sortpath.By(imagesPath, e.SortPathMode))

	var imgStorage epubzip.StorageImageWriter
	imgStorage, err = epubzip.NewStorageImageWriter(e.ImgStorage(), e.Image.Format)
	if err != nil {
		return
	}
	defer imgStorage.Close()

	// processing
	bar := epubprogress.New(epubprogress.Options{
		Quiet:       e.Quiet,
		Json:        e.Json,
		Max:         len(imagesPath),
		Description: "Copying",
		CurrentJob:  1,
		TotalJob:    2,
	})
	defer bar.Close()

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
		f.Close()

		var img epubimage.EPUBImage
		img, err = copyRawDataToStorage(
			imgStorage,
			uncompressedData,
			i,
			input,
			imgPath,
		)
		if err != nil {
			return
		}

		images = append(images, img)
		_ = bar.Add(1)
	}

	if len(images) == 0 {
		err = errNoImagesFound
	}

	return

}

func (e ePUBImagePassthrough) loadCbz() (images []epubimage.EPUBImage, err error) {
	images = make([]epubimage.EPUBImage, 0)

	input := filepath.Clean(e.Input)
	r, err := zip.OpenReader(input)
	if err != nil {
		return
	}
	defer r.Close()

	imagesZip := make([]*zip.File, 0)
	for _, f := range r.File {
		if filterCopyPath(f.FileInfo().IsDir(), f.Name) {
			imagesZip = append(imagesZip, f)
		}
	}

	if len(imagesZip) == 0 {
		err = errNoImagesFound
		return
	}

	var names []string
	for _, img := range imagesZip {
		names = append(names, img.Name)
	}

	sort.Sort(sortpath.By(names, e.SortPathMode))

	indexedNames := make(map[string]int)
	for i, name := range names {
		indexedNames[name] = i
	}

	var imgStorage epubzip.StorageImageWriter
	imgStorage, err = epubzip.NewStorageImageWriter(e.ImgStorage(), e.Image.Format)
	if err != nil {
		return
	}
	defer imgStorage.Close()

	// processing
	bar := epubprogress.New(epubprogress.Options{
		Quiet:       e.Quiet,
		Json:        e.Json,
		Max:         len(imagesZip),
		Description: "Copying",
		CurrentJob:  1,
		TotalJob:    2,
	})
	defer bar.Close()

	for _, imgZip := range imagesZip {
		if _, ok := indexedNames[imgZip.Name]; !ok {
			continue
		}

		var f io.ReadCloser
		f, err = imgZip.Open()
		if err != nil {
			return
		}

		var uncompressedData []byte
		uncompressedData, err = io.ReadAll(f)
		if err != nil {
			return
		}
		f.Close()

		var img epubimage.EPUBImage
		img, err = copyRawDataToStorage(
			imgStorage,
			uncompressedData,
			indexedNames[imgZip.Name],
			"",
			imgZip.Name,
		)

		if err != nil {
			return
		}

		images = append(images, img)
		_ = bar.Add(1)
	}

	if len(images) == 0 {
		err = errNoImagesFound
	}

	return
}

func (e ePUBImagePassthrough) loadCbr() (images []epubimage.EPUBImage, err error) {
	images = make([]epubimage.EPUBImage, 0)

	var isSolid bool
	files, err := rardecode.List(e.Input)
	if err != nil {
		return
	}

	names := make([]string, 0)
	for _, f := range files {
		if filterCopyPath(f.IsDir, f.Name) {
			if f.Solid {
				isSolid = true
			}
			names = append(names, f.Name)
		}
	}

	if len(names) == 0 {
		err = errNoImagesFound
		return
	}

	sort.Sort(sortpath.By(names, e.SortPathMode))

	indexedNames := make(map[string]int)
	for i, name := range names {
		indexedNames[name] = i
	}

	var imgStorage epubzip.StorageImageWriter
	imgStorage, err = epubzip.NewStorageImageWriter(e.ImgStorage(), e.Image.Format)
	if err != nil {
		return
	}
	defer imgStorage.Close()

	// processing
	bar := epubprogress.New(epubprogress.Options{
		Quiet:       e.Quiet,
		Json:        e.Json,
		Max:         len(names),
		Description: "Copying",
		CurrentJob:  1,
		TotalJob:    2,
	})
	defer bar.Close()

	if isSolid {
		var r *rardecode.ReadCloser
		r, err = rardecode.OpenReader(e.Input)
		if err != nil {
			return
		}
		defer r.Close()

		for {
			f, rerr := r.Next()
			if rerr != nil {
				if rerr == io.EOF {
					break
				}
				err = rerr
				return
			}

			if _, ok := indexedNames[f.Name]; !ok {
				continue
			}

			var uncompressedData []byte
			uncompressedData, err = io.ReadAll(r)
			if err != nil {
				return
			}

			var img epubimage.EPUBImage
			img, err = copyRawDataToStorage(
				imgStorage,
				uncompressedData,
				indexedNames[f.Name],
				"",
				f.Name,
			)

			if err != nil {
				return
			}

			images = append(images, img)
			_ = bar.Add(1)
		}
	} else {
		for _, file := range files {
			if i, ok := indexedNames[file.Name]; ok {
				var f io.ReadCloser
				f, err = file.Open()
				if err != nil {
					return
				}

				var uncompressedData []byte
				uncompressedData, err = io.ReadAll(f)
				if err != nil {
					return
				}
				f.Close()

				var img epubimage.EPUBImage
				img, err = copyRawDataToStorage(
					imgStorage,
					uncompressedData,
					i,
					"",
					file.Name,
				)

				if err != nil {
					return
				}

				images = append(images, img)
				_ = bar.Add(1)
			}
		}
	}

	if len(images) == 0 {
		err = errNoImagesFound
	}

	return
}

func filterCopyPath(isDir bool, filename string) bool {
	return !isDir &&
		!strings.HasPrefix(filepath.Base(filename), ".") &&
		slices.Contains([]string{".jpeg", ".jpg", ".png"}, strings.ToLower(filepath.Ext(filename)))
}

func copyRawDataToStorage(
	imgStorage epubzip.StorageImageWriter,
	uncompressedData []byte,
	id int,
	dirname string,
	filename string,
) (img epubimage.EPUBImage, err error) {
	p, fn := filepath.Split(filepath.Clean(filename))
	if p == dirname {
		p = ""
	} else {
		p = p[len(dirname)+1:]
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
	if id == 0 {
		rawImage, err = decode(bytes.NewReader(uncompressedData))
		if err != nil {
			return
		}
	}

	img = epubimage.EPUBImage{
		Id:                  id,
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

	return
}
