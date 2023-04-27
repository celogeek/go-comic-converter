/*
Extract and transform image into a compressed jpeg.
*/
package epubimageprocessing

import (
	"image"
	"path/filepath"
	"strings"
	"sync"

	epubimage "github.com/celogeek/go-comic-converter/v2/internal/epub/image"
	epubimagedata "github.com/celogeek/go-comic-converter/v2/internal/epub/imagedata"
	epubimagefilters "github.com/celogeek/go-comic-converter/v2/internal/epub/imagefilters"
	epubprogress "github.com/celogeek/go-comic-converter/v2/internal/epub/progress"
	"github.com/disintegration/gift"
)

// only accept jpg, png and webp as source file
func isSupportedImage(path string) bool {
	switch strings.ToLower(filepath.Ext(path)) {
	case ".jpg", ".jpeg", ".png", ".webp":
		{
			return true
		}
	}
	return false
}

// extract and convert images
func LoadImages(o *Options) ([]*epubimage.Image, error) {
	images := make([]*epubimage.Image, 0)

	imageCount, imageInput, err := o.Load()
	if err != nil {
		return nil, err
	}

	// dry run, skip convertion
	if o.Dry {
		for img := range imageInput {
			images = append(images, &epubimage.Image{
				Id:   img.Id,
				Path: img.Path,
				Name: img.Name,
			})
		}

		return images, nil
	}

	imageOutput := make(chan *epubimage.Image)

	// processing
	bar := epubprogress.New(epubprogress.Options{
		Quiet:       o.Quiet,
		Max:         imageCount,
		Description: "Processing",
		CurrentJob:  1,
		TotalJob:    2,
	})
	wg := &sync.WaitGroup{}

	for i := 0; i < o.WorkersRatio(50); i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()

			for img := range imageInput {
				src := img.Image

				for part, dst := range TransformImage(src, img.Id, o.Image) {
					var raw image.Image
					if img.Id == 0 && part == 0 {
						raw = dst
					}

					imageOutput <- &epubimage.Image{
						Id:         img.Id,
						Part:       part,
						Raw:        raw,
						Data:       epubimagedata.New(img.Id, part, dst, o.Image.Quality),
						Width:      dst.Bounds().Dx(),
						Height:     dst.Bounds().Dy(),
						IsCover:    img.Id == 0 && part == 0,
						DoublePage: part == 0 && src.Bounds().Dx() > src.Bounds().Dy(),
						Path:       img.Path,
						Name:       img.Name,
					}
				}
			}
		}()
	}

	go func() {
		wg.Wait()
		close(imageOutput)
	}()

	for img := range imageOutput {
		if img.Part == 0 {
			bar.Add(1)
		}
		if o.Image.NoBlankPage && img.Width == 1 && img.Height == 1 {
			continue
		}
		images = append(images, img)
	}
	bar.Close()

	if len(images) == 0 {
		return nil, errNoImagesFound
	}

	return images, nil
}

// create a title page with the cover
func LoadCoverTitleData(img *epubimage.Image, title string, quality int) *epubimagedata.ImageData {
	// Create a blur version of the cover
	g := gift.New(epubimagefilters.CoverTitle(title))
	dst := image.NewGray(g.Bounds(img.Raw.Bounds()))
	g.Draw(dst, img.Raw)

	return epubimagedata.NewRaw("OEBPS/Images/title.jpg", dst, quality)
}

// transform image into 1 or 3 images
// only doublepage with autosplit has 3 versions
func TransformImage(src image.Image, srcId int, o *epubimage.Options) []image.Image {
	var filters, splitFilter []gift.Filter
	var images []image.Image

	if o.Crop {
		f := epubimagefilters.AutoCrop(
			src,
			o.CropRatioLeft,
			o.CropRatioUp,
			o.CropRatioRight,
			o.CropRatioBottom,
		)
		filters = append(filters, f)
		splitFilter = append(splitFilter, f)
	}

	if o.AutoRotate && src.Bounds().Dx() > src.Bounds().Dy() {
		filters = append(filters, gift.Rotate90())
	}

	if o.Contrast != 0 {
		f := gift.Contrast(float32(o.Contrast))
		filters = append(filters, f)
		splitFilter = append(splitFilter, f)
	}

	if o.Brightness != 0 {
		f := gift.Brightness(float32(o.Brightness))
		filters = append(filters, f)
		splitFilter = append(splitFilter, f)
	}

	filters = append(filters,
		epubimagefilters.Resize(o.ViewWidth, o.ViewHeight, gift.LanczosResampling),
		epubimagefilters.Pixel(),
	)

	// convert
	{
		g := gift.New(filters...)
		dst := image.NewGray(g.Bounds(src.Bounds()))
		g.Draw(dst, src)
		images = append(images, dst)
	}

	// auto split off
	if !o.AutoSplitDoublePage {
		return images
	}

	// portrait, no need to split
	if src.Bounds().Dx() <= src.Bounds().Dy() {
		return images
	}

	// cover
	if o.HasCover && srcId == 0 {
		return images
	}

	// convert double page
	for _, b := range []bool{o.Manga, !o.Manga} {
		g := gift.New(splitFilter...)
		g.Add(
			epubimagefilters.CropSplitDoublePage(b),
			epubimagefilters.Resize(o.ViewWidth, o.ViewHeight, gift.LanczosResampling),
		)
		dst := image.NewGray(g.Bounds(src.Bounds()))
		g.Draw(dst, src)
		images = append(images, dst)
	}

	return images
}
