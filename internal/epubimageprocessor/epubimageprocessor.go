/*
Extract and transform image into a compressed jpeg.
*/
package epubimageprocessor

import (
	"fmt"
	"image"
	"image/draw"
	"os"
	"sync"

	"github.com/celogeek/go-comic-converter/v2/internal/epubimage"
	"github.com/celogeek/go-comic-converter/v2/internal/epubimagefilters"
	"github.com/celogeek/go-comic-converter/v2/internal/epubprogress"
	"github.com/celogeek/go-comic-converter/v2/internal/epubzip"
	"github.com/celogeek/go-comic-converter/v2/pkg/epuboptions"
	"github.com/disintegration/gift"
)

type EPUBImageProcessor struct {
	options *epuboptions.EPUBOptions
}

func New(o *epuboptions.EPUBOptions) *EPUBImageProcessor {
	return &EPUBImageProcessor{o}
}

// extract and convert images
func (e *EPUBImageProcessor) Load() (images []*epubimage.EPUBImage, err error) {
	images = make([]*epubimage.EPUBImage, 0)
	imageCount, imageInput, err := e.load()
	if err != nil {
		return nil, err
	}

	// dry run, skip convertion
	if e.options.Dry {
		for img := range imageInput {
			images = append(images, &epubimage.EPUBImage{
				Id:     img.Id,
				Path:   img.Path,
				Name:   img.Name,
				Format: e.options.Image.Format,
			})
		}

		return images, nil
	}

	imageOutput := make(chan *epubimage.EPUBImage)

	// processing
	bar := epubprogress.New(epubprogress.Options{
		Quiet:       e.options.Quiet,
		Json:        e.options.Json,
		Max:         imageCount,
		Description: "Processing",
		CurrentJob:  1,
		TotalJob:    2,
	})
	wg := &sync.WaitGroup{}

	imgStorage, err := epubzip.NewImageWriter(e.options.Temp(), e.options.Image.Format)
	if err != nil {
		bar.Close()
		return nil, err
	}

	wr := 50
	if e.options.Image.Format == "png" {
		wr = 100
	}
	for i := 0; i < e.options.WorkersRatio(wr); i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()

			for input := range imageInput {
				src := input.Image

				for part, dst := range e.transformImage(src, input.Id) {
					var raw image.Image
					if input.Id == 0 && part == 0 {
						raw = dst
					}

					img := &epubimage.EPUBImage{
						Id:                  input.Id,
						Part:                part,
						Raw:                 raw,
						Width:               dst.Bounds().Dx(),
						Height:              dst.Bounds().Dy(),
						IsCover:             input.Id == 0 && part == 0,
						IsBlank:             dst.Bounds().Dx() == 1 && dst.Bounds().Dy() == 1,
						DoublePage:          part == 0 && src.Bounds().Dx() > src.Bounds().Dy(),
						Path:                input.Path,
						Name:                input.Name,
						Format:              e.options.Image.Format,
						OriginalAspectRatio: float64(src.Bounds().Dy()) / float64(src.Bounds().Dx()),
						Error:               input.Error,
					}

					// do not keep double page if requested
					if !img.IsCover &&
						img.DoublePage &&
						e.options.Image.AutoSplitDoublePage &&
						!e.options.Image.KeepDoublePageIfSplitted {
						continue
					}

					if err = imgStorage.Add(img.EPUBImgPath(), dst, e.options.Image.Quality); err != nil {
						bar.Close()
						fmt.Fprintf(os.Stderr, "error with %s: %s", input.Name, err)
						os.Exit(1)
					}
					imageOutput <- img
				}
			}
		}()
	}

	go func() {
		wg.Wait()
		imgStorage.Close()
		close(imageOutput)
	}()

	for img := range imageOutput {
		if img.Part == 0 {
			bar.Add(1)
		}
		if e.options.Image.NoBlankImage && img.IsBlank {
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

func (e *EPUBImageProcessor) createImage(src image.Image, r image.Rectangle) draw.Image {
	if e.options.Image.GrayScale {
		return image.NewGray(r)
	}

	switch t := src.(type) {
	case *image.Gray:
		return image.NewGray(r)
	case *image.Gray16:
		return image.NewGray16(r)
	case *image.RGBA:
		return image.NewRGBA(r)
	case *image.RGBA64:
		return image.NewRGBA64(r)
	case *image.NRGBA:
		return image.NewNRGBA(r)
	case *image.NRGBA64:
		return image.NewNRGBA64(r)
	case *image.Alpha:
		return image.NewAlpha(r)
	case *image.Alpha16:
		return image.NewAlpha16(r)
	case *image.CMYK:
		return image.NewCMYK(r)
	case *image.Paletted:
		return image.NewPaletted(r, t.Palette)
	default:
		return image.NewNRGBA64(r)
	}
}

// transform image into 1 or 3 images
// only doublepage with autosplit has 3 versions
func (e *EPUBImageProcessor) transformImage(src image.Image, srcId int) []image.Image {
	var filters, splitFilters []gift.Filter
	var images []image.Image

	// Lookup for margin if crop is enable or if we want to remove blank image
	if e.options.Image.Crop.Enabled || e.options.Image.NoBlankImage {
		f := epubimagefilters.AutoCrop(
			src,
			e.options.Image.Crop.Left,
			e.options.Image.Crop.Up,
			e.options.Image.Crop.Right,
			e.options.Image.Crop.Bottom,
		)

		// detect if blank image
		size := f.Bounds(src.Bounds())
		isBlank := size.Dx() == 0 && size.Dy() == 0

		// crop is enable or if blank image with noblankimage options
		if e.options.Image.Crop.Enabled || (e.options.Image.NoBlankImage && isBlank) {
			filters = append(filters, f)
			splitFilters = append(splitFilters, f)
		}
	}

	if e.options.Image.AutoRotate && src.Bounds().Dx() > src.Bounds().Dy() {
		filters = append(filters, gift.Rotate90())
	}

	if e.options.Image.AutoContrast {
		f := epubimagefilters.AutoContrast()
		filters = append(filters, f)
		splitFilters = append(splitFilters, f)
	}

	if e.options.Image.Contrast != 0 {
		f := gift.Contrast(float32(e.options.Image.Contrast))
		filters = append(filters, f)
		splitFilters = append(splitFilters, f)
	}

	if e.options.Image.Brightness != 0 {
		f := gift.Brightness(float32(e.options.Image.Brightness))
		filters = append(filters, f)
		splitFilters = append(splitFilters, f)
	}

	if e.options.Image.Resize {
		f := gift.ResizeToFit(e.options.Image.View.Width, e.options.Image.View.Height, gift.LanczosResampling)
		filters = append(filters, f)
	}

	if e.options.Image.GrayScale {
		var f gift.Filter
		switch e.options.Image.GrayScaleMode {
		case 1: // average
			f = gift.ColorFunc(func(r0, g0, b0, a0 float32) (r float32, g float32, b float32, a float32) {
				y := (r0 + g0 + b0) / 3
				return y, y, y, a0
			})
		case 2: // luminance
			f = gift.ColorFunc(func(r0, g0, b0, a0 float32) (r float32, g float32, b float32, a float32) {
				y := 0.2126*r0 + 0.7152*g0 + 0.0722*b0
				return y, y, y, a0
			})
		default:
			f = gift.Grayscale()
		}
		filters = append(filters, f)
		splitFilters = append(splitFilters, f)
	}

	filters = append(filters, epubimagefilters.Pixel())

	// convert
	{
		g := gift.New(filters...)
		dst := e.createImage(src, g.Bounds(src.Bounds()))
		g.Draw(dst, src)
		images = append(images, dst)
	}

	// auto split off
	if !e.options.Image.AutoSplitDoublePage {
		return images
	}

	// portrait, no need to split
	if src.Bounds().Dx() <= src.Bounds().Dy() {
		return images
	}

	// cover
	if e.options.Image.HasCover && srcId == 0 {
		return images
	}

	// convert double page
	for _, b := range []bool{e.options.Image.Manga, !e.options.Image.Manga} {
		g := gift.New(splitFilters...)
		g.Add(epubimagefilters.CropSplitDoublePage(b))
		if e.options.Image.Resize {
			g.Add(gift.ResizeToFit(e.options.Image.View.Width, e.options.Image.View.Height, gift.LanczosResampling))
		}
		dst := e.createImage(src, g.Bounds(src.Bounds()))
		g.Draw(dst, src)
		images = append(images, dst)
	}

	return images
}
