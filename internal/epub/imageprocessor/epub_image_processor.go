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

	epubimage "github.com/celogeek/go-comic-converter/v2/internal/epub/image"
	epubimagefilters "github.com/celogeek/go-comic-converter/v2/internal/epub/imagefilters"
	epuboptions "github.com/celogeek/go-comic-converter/v2/internal/epub/options"
	epubprogress "github.com/celogeek/go-comic-converter/v2/internal/epub/progress"
	epubzip "github.com/celogeek/go-comic-converter/v2/internal/epub/zip"
	"github.com/disintegration/gift"
)

type EPUBImageProcessor struct {
	*epuboptions.Options
}

func New(o *epuboptions.Options) *EPUBImageProcessor {
	return &EPUBImageProcessor{o}
}

// extract and convert images
func (e *EPUBImageProcessor) Load() (images []*epubimage.Image, err error) {
	images = make([]*epubimage.Image, 0)
	imageCount, imageInput, err := e.load()
	if err != nil {
		return nil, err
	}

	// dry run, skip convertion
	if e.Dry {
		for img := range imageInput {
			images = append(images, &epubimage.Image{
				Id:     img.Id,
				Path:   img.Path,
				Name:   img.Name,
				Format: e.Image.Format,
			})
		}

		return images, nil
	}

	imageOutput := make(chan *epubimage.Image)

	// processing
	bar := epubprogress.New(epubprogress.Options{
		Quiet:       e.Quiet,
		Max:         imageCount,
		Description: "Processing",
		CurrentJob:  1,
		TotalJob:    2,
	})
	wg := &sync.WaitGroup{}

	imgStorage, err := epubzip.NewEPUBZipStorageImageWriter(e.ImgStorage(), e.Image.Format)
	if err != nil {
		bar.Close()
		return nil, err
	}

	wr := 50
	if e.Image.Format == "png" {
		wr = 100
	}
	for i := 0; i < e.WorkersRatio(wr); i++ {
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

					img := &epubimage.Image{
						Id:         input.Id,
						Part:       part,
						Raw:        raw,
						Width:      dst.Bounds().Dx(),
						Height:     dst.Bounds().Dy(),
						IsCover:    input.Id == 0 && part == 0,
						IsBlank:    dst.Bounds().Dx() == 1 && dst.Bounds().Dy() == 1,
						DoublePage: part == 0 && src.Bounds().Dx() > src.Bounds().Dy(),
						Path:       input.Path,
						Name:       input.Name,
						Format:     e.Image.Format,
					}

					if err = imgStorage.Add(img.EPUBImgPath(), dst, e.Image.Quality); err != nil {
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
		if e.Image.NoBlankImage && img.IsBlank {
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
	if e.Options.Image.GrayScale {
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
	var filters, splitFilter []gift.Filter
	var images []image.Image

	// Lookup for margin if crop is enable or if we want to remove blank image
	if e.Image.Crop.Enabled || e.Image.NoBlankImage {
		f := epubimagefilters.AutoCrop(
			src,
			e.Image.Crop.Left,
			e.Image.Crop.Up,
			e.Image.Crop.Right,
			e.Image.Crop.Bottom,
		)

		// detect if blank image
		size := f.Bounds(src.Bounds())
		isBlank := size.Dx() == 0 && size.Dy() == 0

		// crop is enable or if blank image with noblankimage options
		if e.Image.Crop.Enabled || (e.Image.NoBlankImage && isBlank) {
			filters = append(filters, f)
			splitFilter = append(splitFilter, f)
		}
	}

	if e.Image.AutoRotate && src.Bounds().Dx() > src.Bounds().Dy() {
		filters = append(filters, gift.Rotate90())
	}

	if e.Image.Contrast != 0 {
		f := gift.Contrast(float32(e.Image.Contrast))
		filters = append(filters, f)
		splitFilter = append(splitFilter, f)
	}

	if e.Image.Brightness != 0 {
		f := gift.Brightness(float32(e.Image.Brightness))
		filters = append(filters, f)
		splitFilter = append(splitFilter, f)
	}

	if e.Image.Resize {
		f := gift.ResizeToFit(e.Image.View.Width, e.Image.View.Height, gift.LanczosResampling)
		filters = append(filters, f)
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
	if !e.Image.AutoSplitDoublePage {
		return images
	}

	// portrait, no need to split
	if src.Bounds().Dx() <= src.Bounds().Dy() {
		return images
	}

	// cover
	if e.Image.HasCover && srcId == 0 {
		return images
	}

	// convert double page
	for _, b := range []bool{e.Image.Manga, !e.Image.Manga} {
		g := gift.New(splitFilter...)
		g.Add(epubimagefilters.CropSplitDoublePage(b))
		if e.Image.Resize {
			g.Add(gift.ResizeToFit(e.Image.View.Width, e.Image.View.Height, gift.LanczosResampling))
		}
		dst := e.createImage(src, g.Bounds(src.Bounds()))
		g.Draw(dst, src)
		images = append(images, dst)
	}

	return images
}

// create a title page with the cover
func (e *EPUBImageProcessor) CoverTitleData(src image.Image, title string) (*epubzip.ZipImage, error) {
	// Create a blur version of the cover
	g := gift.New(epubimagefilters.CoverTitle(title))
	dst := e.createImage(src, g.Bounds(src.Bounds()))
	g.Draw(dst, src)

	return epubzip.CompressImage(
		fmt.Sprintf("OEBPS/Images/title.%s", e.Image.Format),
		e.Image.Format,
		dst,
		e.Image.Quality,
	)
}
