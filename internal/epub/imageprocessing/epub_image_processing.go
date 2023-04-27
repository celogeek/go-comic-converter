/*
Extract and transform image into a compressed jpeg.
*/
package epubimageprocessing

import (
	"fmt"
	"image"
	"path/filepath"
	"strings"
	"sync"

	epubimage "github.com/celogeek/go-comic-converter/v2/internal/epub/image"
	epubimagefilters "github.com/celogeek/go-comic-converter/v2/internal/epub/imagefilters"
	epuboptions "github.com/celogeek/go-comic-converter/v2/internal/epub/options"
	epubprogress "github.com/celogeek/go-comic-converter/v2/internal/epub/progress"
	epubzip "github.com/celogeek/go-comic-converter/v2/internal/epub/zip"
	"github.com/disintegration/gift"
)

type LoadedImage struct {
	Image    *epubimage.Image
	ZipImage *epubzip.ZipImage
}

type LoadedImages []*LoadedImage

func (l LoadedImages) Images() []*epubimage.Image {
	res := make([]*epubimage.Image, len(l))
	for i, v := range l {
		res[i] = v.Image
	}
	return res
}

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
func LoadImages(o *epuboptions.Options) (LoadedImages, error) {
	images := make(LoadedImages, 0)

	imageCount, imageInput, err := Load(o)
	if err != nil {
		return nil, err
	}

	// dry run, skip convertion
	if o.Dry {
		for img := range imageInput {
			images = append(images, &LoadedImage{
				Image: &epubimage.Image{
					Id:   img.Id,
					Path: img.Path,
					Name: img.Name,
				},
			})
		}

		return images, nil
	}

	imageOutput := make(chan *LoadedImage)

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

					imageOutput <- &LoadedImage{
						Image: &epubimage.Image{
							Id:         img.Id,
							Part:       part,
							Raw:        raw,
							Width:      dst.Bounds().Dx(),
							Height:     dst.Bounds().Dy(),
							IsCover:    img.Id == 0 && part == 0,
							DoublePage: part == 0 && src.Bounds().Dx() > src.Bounds().Dy(),
							Path:       img.Path,
							Name:       img.Name,
						},
						ZipImage: epubzip.CompressImage(fmt.Sprintf("OEBPS/Images/%d_p%d.jpg", img.Id, part), dst, o.Image.Quality),
					}
				}
			}
		}()
	}

	go func() {
		wg.Wait()
		close(imageOutput)
	}()

	for output := range imageOutput {
		if output.Image.Part == 0 {
			bar.Add(1)
		}
		if o.Image.NoBlankPage && output.Image.Width == 1 && output.Image.Height == 1 {
			continue
		}
		images = append(images, output)
	}
	bar.Close()

	if len(images) == 0 {
		return nil, errNoImagesFound
	}

	return images, nil
}

// create a title page with the cover
func CoverTitleData(img image.Image, title string, quality int) *epubzip.ZipImage {
	// Create a blur version of the cover
	g := gift.New(epubimagefilters.CoverTitle(title))
	dst := image.NewGray(g.Bounds(img.Bounds()))
	g.Draw(dst, img)

	return epubzip.CompressImage("OEBPS/Images/title.jpg", dst, quality)
}

// transform image into 1 or 3 images
// only doublepage with autosplit has 3 versions
func TransformImage(src image.Image, srcId int, o *epuboptions.Image) []image.Image {
	var filters, splitFilter []gift.Filter
	var images []image.Image

	if o.Crop.Enabled {
		f := epubimagefilters.AutoCrop(
			src,
			o.Crop.Left,
			o.Crop.Up,
			o.Crop.Right,
			o.Crop.Bottom,
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
		epubimagefilters.Resize(o.View.Width, o.View.Height, gift.LanczosResampling),
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
			epubimagefilters.Resize(o.View.Width, o.View.Height, gift.LanczosResampling),
		)
		dst := image.NewGray(g.Bounds(src.Bounds()))
		g.Draw(dst, src)
		images = append(images, dst)
	}

	return images
}
