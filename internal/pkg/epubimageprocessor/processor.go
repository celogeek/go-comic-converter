// Package epubimageprocessor extract and transform image into a compressed jpeg.
package epubimageprocessor

import (
	"fmt"
	"image"
	"image/color"
	"image/draw"
	"sync"

	"github.com/disintegration/gift"

	"github.com/celogeek/go-comic-converter/v3/internal/pkg/epubimage"
	"github.com/celogeek/go-comic-converter/v3/internal/pkg/epubimagefilters"
	"github.com/celogeek/go-comic-converter/v3/internal/pkg/epubprogress"
	"github.com/celogeek/go-comic-converter/v3/internal/pkg/epubzip"
	"github.com/celogeek/go-comic-converter/v3/internal/pkg/utils"
	"github.com/celogeek/go-comic-converter/v3/pkg/epuboptions"
)

type EPUBImageProcessor interface {
	Load() (images []epubimage.EPUBImage, err error)
	CoverTitleData(o CoverTitleDataOptions) ([]epubzip.Image, error)
}

type ePUBImageProcessor struct {
	epuboptions.EPUBOptions
}

func New(o epuboptions.EPUBOptions) EPUBImageProcessor {
	return ePUBImageProcessor{o}
}

// Load extract and convert images
func (e ePUBImageProcessor) Load() (images []epubimage.EPUBImage, err error) {
	images = make([]epubimage.EPUBImage, 0)
	imageCount, imageInput, err := e.load()
	if err != nil {
		return nil, err
	}

	// dry run, skip conversion
	if e.Dry {
		for img := range imageInput {
			if img.Fatal != nil {
				return nil, img.Fatal
			}
			images = append(images, epubimage.EPUBImage{
				Id:     img.Id,
				Path:   img.Path,
				Name:   img.Name,
				Format: e.Image.Format,
			})
		}

		return images, nil
	}

	imageOutput := make(chan epubimage.EPUBImage)

	// processing
	bar := epubprogress.New(epubprogress.Options{
		Quiet:       e.Quiet,
		Json:        e.Json,
		Max:         imageCount,
		Description: "Processing",
		CurrentJob:  1,
		TotalJob:    2,
	})
	wg := &sync.WaitGroup{}

	var source <-chan task = imageInput

	// 2 pages merged count as 2 in the progress
	if e.Image.CombineSplitDoublePage {
		source = e.combineSplitDoublePages(imageInput, func() { _ = bar.Add(1) })
	}

	imgStorage, err := epubzip.NewStorageImageWriter(e.ImgStorage(), e.Image.Format)
	if err != nil {
		_ = bar.Close()
		return nil, err
	}

	wr := 50
	if e.Image.Format == "png" {
		wr = 100
	}
	// The first error stops the conversion, the workers then only drain the input so the loaders can finish.
	var (
		failMu  sync.Mutex
		failure error
	)
	fail := func(err error) {
		failMu.Lock()
		defer failMu.Unlock()
		if failure == nil {
			failure = err
		}
	}
	failed := func() bool {
		failMu.Lock()
		defer failMu.Unlock()
		return failure != nil
	}

	for range e.WorkersRatio(wr) {
		wg.Add(1)
		go func() {
			defer wg.Done()

			for input := range source {
				if input.Fatal != nil {
					fail(input.Fatal)
				}
				if failed() {
					continue
				}
				if err := e.processImage(imgStorage, input, imageOutput); err != nil {
					fail(fmt.Errorf("error with %s: %w", input.Name, err))
				}
			}
		}()
	}

	go func() {
		wg.Wait()
		_ = imgStorage.Close()
		close(imageOutput)
	}()

	// The cover is the first image once the blank ones are removed, it is the only one that keeps its raw data.
	// Images come in any order: the raw data of the best candidate so far is released when a better one arrives.
	cover := -1
	for img := range imageOutput {
		if img.Part == 0 {
			_ = bar.Add(1)
		}
		if e.Image.NoBlankImage && img.IsBlank {
			continue
		}

		if cover >= 0 && !isBefore(img, images[cover]) {
			img.Raw = nil
			// not stored, and not a page: nothing to keep
			if e.isDroppedDoublePage(img) {
				continue
			}
			images = append(images, img)
			continue
		}

		if cover >= 0 {
			images[cover].Raw = nil
		}
		images = append(images, img)
		cover = len(images) - 1
	}
	_ = bar.Close()

	if failure != nil {
		return nil, failure
	}

	if len(images) == 0 {
		return nil, errNoImagesFound
	}

	// a double page that is not kept is only used as the cover, not as a page
	kept := images[:0]
	for i, img := range images {
		if i == cover || !e.isDroppedDoublePage(img) {
			kept = append(kept, img)
		}
	}

	return kept, nil
}

// processImage transforms the image, stores it and sends it to the output, with its parts when it is a split double page.
func (e ePUBImageProcessor) processImage(imgStorage epubzip.StorageImageWriter, input task, output chan<- epubimage.EPUBImage) error {
	img := e.transformImage(input, 0, e.Image.Manga)

	// A double page that is not kept is not stored, it is only sent in case it ends up as the cover:
	// the collector knows which image comes first once the blank ones are removed.
	if !e.isDroppedDoublePage(img) {
		if err := e.addImage(imgStorage, &img); err != nil {
			return err
		}
	}
	output <- img

	// DOUBLE PAGE
	if !e.Image.SplitDoublePage() || // No split required
		!img.DoublePage { // Not a double page
		return nil
	}

	for i, b := range []bool{e.Image.Manga, !e.Image.Manga} {
		img = e.transformImage(input, i+1, b)
		if err := e.addImage(imgStorage, &img); err != nil {
			return err
		}
		output <- img
	}

	return nil
}

// isDroppedDoublePage tells if the image is a double page, that is not kept as a page (its parts are).
func (e ePUBImageProcessor) isDroppedDoublePage(img epubimage.EPUBImage) bool {
	return img.Part == 0 && img.DoublePage && !e.Image.KeepDoublePage()
}

// isBefore tells if the image comes before the other one in the book.
func isBefore(a, b epubimage.EPUBImage) bool {
	return a.Id < b.Id || (a.Id == b.Id && a.Part < b.Part)
}

type subImager interface {
	SubImage(r image.Rectangle) image.Image
}

// addImage compress the image into the storage.
//
// An image bigger than the tile size of the device is stored as a grid of tiles, displayed together in the same page.
func (e ePUBImageProcessor) addImage(s epubzip.StorageImageWriter, img *epubimage.EPUBImage) error {
	img.Cols, img.Rows = 1, 1
	img.Tile = e.Image.Tile()
	sub, ok := img.Raw.(subImager)
	if ok {
		img.Cols, img.Rows = img.Tile.Grid(img.Width, img.Height)
	}
	if !ok || img.Tiles() == 1 {
		return s.Add(img.EPUBImgPath(), img.Raw, e.Image.Quality)
	}

	b := img.Raw.Bounds()
	for idx := range img.Tiles() {
		if err := s.Add(img.EPUBTileImgPath(idx), sub.SubImage(img.TileBounds(idx, b)), e.Image.Quality); err != nil {
			return err
		}
	}

	return nil
}

func (e ePUBImageProcessor) createImage(src image.Image, r image.Rectangle) draw.Image {
	if e.EPUBOptions.Image.GrayScale {
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
// only doublepage with split + rotate has 3 versions
func (e ePUBImageProcessor) transformImage(input task, part int, right bool) epubimage.EPUBImage {
	g := gift.New()
	src := input.Image
	srcBounds := src.Bounds()

	// When the split halves don't need to keep the same aspect ratio,
	// we first cut, then crop.
	if part > 0 && !e.Image.KeepSplitDoublePageAspect {
		g.Add(epubimagefilters.CropSplitDoublePage(right))
	}

	// Lookup for margin if crop is enable or if we want to remove blank image
	if e.Image.Crop.Enabled || e.Image.NoBlankImage {
		f := epubimagefilters.AutoCrop(
			src,
			g.Bounds(src.Bounds()),
			e.Image.Crop.Left,
			e.Image.Crop.Up,
			e.Image.Crop.Right,
			e.Image.Crop.Bottom,
			e.Image.Crop.Limit,
			e.Image.Crop.SkipIfLimitReached,
		)

		// detect if blank image
		size := f.Bounds(srcBounds)
		isBlank := size.Dx() == 0 && size.Dy() == 0

		// crop is enable or if blank image with noblankimage options
		if e.Image.Crop.Enabled || (e.Image.NoBlankImage && isBlank) {
			g.Add(f)
		}
	}

	// With landscape support, we need to keep aspect ratio between each split
	// We first crop, then cut
	if part > 0 && e.Image.KeepSplitDoublePageAspect {
		g.Add(epubimagefilters.CropSplitDoublePage(right))
	}

	dstBounds := g.Bounds(src.Bounds())
	// Original && Cropped version need to landscape oriented
	// Only part 0 can be a double page
	isDoublePage := part == 0 && srcBounds.Dx() > srcBounds.Dy() && dstBounds.Dx() > dstBounds.Dy()

	if e.Image.RotateDoublePage() && isDoublePage {
		g.Add(gift.Rotate90())
	}

	if e.Image.AutoContrast {
		g.Add(epubimagefilters.AutoContrast(e.Image.GrayScaleMode))
	}

	if e.Image.Contrast != 0 {
		g.Add(gift.Contrast(float32(e.Image.Contrast)))
	}

	if e.Image.Brightness != 0 {
		g.Add(gift.Brightness(float32(e.Image.Brightness)))
	}

	if e.Image.Resize {
		g.Add(gift.ResizeToFit(e.Image.View.Width, e.Image.View.Height, gift.LanczosResampling))
	}

	if e.Image.GrayScale {
		g.Add(gift.ColorFunc(func(r0, g0, b0, a0 float32) (r float32, g float32, b float32, a float32) {
			y := utils.RGBToGray(e.Image.GrayScaleMode, r0, g0, b0)
			return y, y, y, a0
		}))
	}

	g.Add(epubimagefilters.Pixel())

	dst := e.createImage(src, g.Bounds(src.Bounds()))
	g.Draw(dst, src)

	return epubimage.EPUBImage{
		Id:                  input.Id,
		Part:                part,
		Raw:                 dst,
		Width:               dst.Bounds().Dx(),
		Height:              dst.Bounds().Dy(),
		IsBlank:             dst.Bounds().Dx() == 1 && dst.Bounds().Dy() == 1,
		DoublePage:          isDoublePage,
		Path:                input.Path,
		Name:                input.Name,
		Format:              e.Image.Format,
		OriginalAspectRatio: float64(src.Bounds().Dy()) / float64(src.Bounds().Dx()),
		Error:               input.Error,
	}

}

type CoverTitleDataOptions struct {
	Src   image.Image
	Name  string
	Title epubimagefilters.CoverTitle
}

func (e ePUBImageProcessor) cover16LevelOfGray(bounds image.Rectangle) draw.Image {
	return image.NewPaletted(bounds, color.Palette{
		color.Gray{},
		color.Gray{Y: 0x11},
		color.Gray{Y: 0x22},
		color.Gray{Y: 0x33},
		color.Gray{Y: 0x44},
		color.Gray{Y: 0x55},
		color.Gray{Y: 0x66},
		color.Gray{Y: 0x77},
		color.Gray{Y: 0x88},
		color.Gray{Y: 0x99},
		color.Gray{Y: 0xAA},
		color.Gray{Y: 0xBB},
		color.Gray{Y: 0xCC},
		color.Gray{Y: 0xDD},
		color.Gray{Y: 0xEE},
		color.Gray{Y: 0xFF},
	})
}

// CoverTitleData draws the title box on the source image (the cover or the title page background) and compresses it to jpeg.
//
// The title image bigger than the tile size of the device is stored as a grid of tiles, like the pages.
// The cover is only used as thumbnail, so it is never tiled, but it is always reduced to the max size,
// whatever the device or the resize option.
func (e ePUBImageProcessor) CoverTitleData(o CoverTitleDataOptions) ([]epubzip.Image, error) {
	g := gift.New()
	if o.Name == "cover" {
		sb := o.Src.Bounds()
		if w, h := e.Image.Tile().FitSize(sb.Dx(), sb.Dy()); w != sb.Dx() || h != sb.Dy() {
			g.Add(gift.Resize(w, h, gift.LanczosResampling))
		}
	}
	g.Add(o.Title)

	var dst draw.Image
	if o.Name == "cover" && e.Image.GrayScale {
		dst = e.cover16LevelOfGray(g.Bounds(o.Src.Bounds()))
	} else {
		dst = e.createImage(o.Src, g.Bounds(o.Src.Bounds()))
	}
	g.Draw(dst, o.Src)

	b := dst.Bounds()
	tiled := e.Image.Tile().Tiled(b.Dx(), b.Dy())
	sub, ok := dst.(subImager)
	if !ok || o.Name == "cover" {
		tiled.Cols, tiled.Rows = 1, 1
	}

	images := make([]epubzip.Image, 0, tiled.Tiles())
	for idx := range tiled.Tiles() {
		var part image.Image = dst
		if tiled.Tiles() > 1 {
			part = sub.SubImage(tiled.TileBounds(idx, b))
		}

		img, err := epubzip.CompressImage(
			"OEBPS/Images/"+epubimage.TileName(o.Name, idx)+".jpeg",
			"jpeg",
			part,
			e.Image.Quality,
		)
		if err != nil {
			return nil, err
		}
		images = append(images, img)
	}

	return images, nil
}
