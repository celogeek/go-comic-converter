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
	"sort"
	"strings"

	"github.com/gen2brain/jxl"
	"github.com/nwaples/rardecode/v2"

	"github.com/celogeek/go-comic-converter/v3/internal/pkg/epubimage"
	"github.com/celogeek/go-comic-converter/v3/internal/pkg/epubimageprocessor"
	"github.com/celogeek/go-comic-converter/v3/internal/pkg/epubzip"
	"github.com/celogeek/go-comic-converter/v3/internal/pkg/sortpath"
	"github.com/celogeek/go-comic-converter/v3/internal/pkg/utils"
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

func (e ePUBImagePassthrough) CoverTitleData(o epubimageprocessor.CoverTitleDataOptions) ([]epubzip.Image, error) {
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

		if !d.IsDir() && e.isSupportedImage(path) {
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
	total := len(imagesPath)

	for i, imgPath := range imagesPath {

		var img epubimage.EPUBImage
		img, err = e.copyRawDataToStorage(
			imgStorage,
			func() ([]byte, error) {
				f, err := os.Open(imgPath)
				if err != nil {
					return nil, err
				}
				defer f.Close()
				return io.ReadAll(f)
			},
			i,
			input,
			imgPath,
		)
		if err != nil {
			return
		}

		images = append(images, img)
		e.Processed(len(images), total)
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
		if !f.FileInfo().IsDir() && e.isSupportedImage(f.Name) {
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
	total := len(indexedNames)

	for _, imgZip := range imagesZip {
		if _, ok := indexedNames[imgZip.Name]; !ok {
			continue
		}

		var img epubimage.EPUBImage
		img, err = e.copyRawDataToStorage(
			imgStorage,
			func() ([]byte, error) {
				f, err := imgZip.Open()
				if err != nil {
					return nil, err
				}
				defer f.Close()

				return io.ReadAll(f)
			},
			indexedNames[imgZip.Name],
			"",
			imgZip.Name,
		)

		if err != nil {
			return
		}

		images = append(images, img)
		e.Processed(len(images), total)
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
		if !f.IsDir && e.isSupportedImage(f.Name) {
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
	total := len(names)

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

			var img epubimage.EPUBImage
			img, err = e.copyRawDataToStorage(
				imgStorage,
				func() ([]byte, error) {
					return io.ReadAll(r)
				},
				indexedNames[f.Name],
				"",
				f.Name,
			)

			if err != nil {
				return
			}

			images = append(images, img)
			e.Processed(len(images), total)
		}
	} else {
		for _, file := range files {
			if i, ok := indexedNames[file.Name]; ok {
				var img epubimage.EPUBImage
				img, err = e.copyRawDataToStorage(
					imgStorage,
					func() ([]byte, error) {
						f, err := file.Open()
						if err != nil {
							return nil, err
						}
						defer f.Close()
						return io.ReadAll(f)
					},
					i,
					"",
					file.Name,
				)

				if err != nil {
					return
				}

				images = append(images, img)
				e.Processed(len(images), total)
			}
		}
	}

	if len(images) == 0 {
		err = errNoImagesFound
	}

	return
}

func (e ePUBImagePassthrough) isSupportedImage(path string) bool {
	switch strings.ToLower(filepath.Ext(path)) {
	case ".jpg", ".jpeg", ".png", ".jxl":
		{
			return !strings.HasPrefix(filepath.Base(path), ".")
		}
	}
	return false
}

func (e ePUBImagePassthrough) copyRawDataToStorage(
	imgStorage epubzip.StorageImageWriter,
	getData func() ([]byte, error),
	id int,
	dirname string,
	filename string,
) (img epubimage.EPUBImage, err error) {
	if e.Dry {
		p, fn := utils.SplitPath(dirname, filename)
		return epubimage.EPUBImage{
			Id:     id,
			Path:   p,
			Name:   fn,
			Format: e.Image.Format,
		}, nil
	}

	// dirname is the input directory to strip, an archive has none
	p, fn := utils.SplitPath(dirname, filename)

	var (
		format       string
		decodeConfig func(r io.Reader) (image.Config, error)
		decode       func(r io.Reader) (image.Image, error)
	)

	// same case insensitive match as isSupportedImage
	switch ext := strings.ToLower(filepath.Ext(fn)); ext {
	case ".png":
		format = "png"
		decodeConfig = png.DecodeConfig
		decode = png.Decode
	case ".jpg", ".jpeg":
		format = "jpeg"
		decodeConfig = jpeg.DecodeConfig
		decode = jpeg.Decode
	case ".jxl":
		// the image is stored as is: the device must display it
		if err = e.Image.CheckFormat("jxl"); err != nil {
			err = fmt.Errorf("%w, it can't be kept in the copy mode: %s", err, fn)
			return
		}
		format = "jxl"
		decodeConfig = jxl.DecodeConfig
		decode = func(r io.Reader) (image.Image, error) { return jxl.Decode(r) }
	default:
		err = fmt.Errorf("unsupported image format (%s): %s", ext, fn)
		return
	}

	// an image that cannot be read is replaced by a page that tells it, and reported at the end
	var (
		uncompressedData []byte
		config           image.Config
		rawImage         image.Image
		readErr          error
	)
	uncompressedData, readErr = getData()
	if readErr == nil {
		config, readErr = decodeConfig(bytes.NewReader(uncompressedData))
	}
	// only the cover needs the decoded image
	if readErr == nil && id == 0 {
		rawImage, readErr = decode(bytes.NewReader(uncompressedData))
	}
	if readErr != nil {
		return e.storeCorruptedImage(imgStorage, id, p, fn, readErr)
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

// storeCorruptedImage stores the corrupted page of an unreadable image, the cause is kept to be reported.
func (e ePUBImagePassthrough) storeCorruptedImage(
	imgStorage epubzip.StorageImageWriter,
	id int,
	path, name string,
	cause error,
) (img epubimage.EPUBImage, err error) {
	page := epubimageprocessor.CorruptedImage(e.EPUBOptions, path, name)

	img = epubimage.EPUBImage{
		Id:                  id,
		Width:               page.Bounds().Dx(),
		Height:              page.Bounds().Dy(),
		Path:                path,
		Name:                name,
		Format:              "jpeg",
		OriginalAspectRatio: float64(page.Bounds().Dy()) / float64(page.Bounds().Dx()),
		Error:               cause,
	}
	// the first image is the cover
	if id == 0 {
		img.Raw = page
	}

	var buf bytes.Buffer
	if err = jpeg.Encode(&buf, page, &jpeg.Options{Quality: e.Image.Quality}); err != nil {
		return
	}
	err = imgStorage.AddRaw(img.EPUBImgPath(), buf.Bytes())

	return
}
