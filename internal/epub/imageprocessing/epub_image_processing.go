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

	for i := 0; i < o.Workers; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()

			for img := range imageInput {
				src := img.Image

				g := epubimagefilters.NewGift(src, o.Image)
				// Convert image
				dst := image.NewGray(g.Bounds(src.Bounds()))
				g.Draw(dst, src)

				var raw image.Image
				if img.Id == 0 {
					raw = dst
				}

				imageOutput <- &epubimage.Image{
					Id:         img.Id,
					Part:       0,
					Raw:        raw,
					Data:       epubimagedata.New(img.Id, 0, dst, o.Image.Quality),
					Width:      dst.Bounds().Dx(),
					Height:     dst.Bounds().Dy(),
					IsCover:    img.Id == 0,
					DoublePage: src.Bounds().Dx() > src.Bounds().Dy(),
					Path:       img.Path,
					Name:       img.Name,
				}

				// Auto split double page
				// Except for cover
				// Only if the src image have width > height and is bigger than the view
				if (!o.Image.HasCover || img.Id > 0) &&
					o.Image.AutoSplitDoublePage &&
					src.Bounds().Dx() > src.Bounds().Dy() {
					gifts := epubimagefilters.NewGiftSplitDoublePage(o.Image)
					for i, g := range gifts {
						part := i + 1
						dst := image.NewGray(g.Bounds(src.Bounds()))
						g.Draw(dst, src)

						imageOutput <- &epubimage.Image{
							Id:         img.Id,
							Part:       part,
							Data:       epubimagedata.New(img.Id, part, dst, o.Image.Quality),
							Width:      dst.Bounds().Dx(),
							Height:     dst.Bounds().Dy(),
							IsCover:    false,
							DoublePage: false,
							Path:       img.Path,
							Name:       img.Name,
						}
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
