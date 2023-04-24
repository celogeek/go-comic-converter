package epubimageprocessing

import (
	"fmt"
	"image"
	_ "image/jpeg"
	_ "image/png"
	"io"
	"os"
	"path/filepath"
	"strings"
	"sync"

	epubfilters "github.com/celogeek/go-comic-converter/v2/internal/epub/filters"
	epubimage "github.com/celogeek/go-comic-converter/v2/internal/epub/image"
	epubimagedata "github.com/celogeek/go-comic-converter/v2/internal/epub/imagedata"
	epubprogress "github.com/celogeek/go-comic-converter/v2/internal/epub/progress"
	"github.com/disintegration/gift"
	_ "golang.org/x/image/webp"
)

type tasks struct {
	Id     int
	Reader io.Reader
	Path   string
	Name   string
}

func LoadImages(o *Options) ([]*epubimage.Image, error) {
	images := make([]*epubimage.Image, 0)

	fi, err := os.Stat(o.Input)
	if err != nil {
		return nil, err
	}

	var (
		imageCount int
		imageInput chan *tasks
	)

	if fi.IsDir() {
		imageCount, imageInput, err = o.loadDir()
	} else {
		switch ext := strings.ToLower(filepath.Ext(o.Input)); ext {
		case ".cbz", ".zip":
			imageCount, imageInput, err = o.loadCbz()
		case ".cbr", ".rar":
			imageCount, imageInput, err = o.loadCbr()
		case ".pdf":
			imageCount, imageInput, err = o.loadPdf()
		default:
			err = fmt.Errorf("unknown file format (%s): support .cbz, .zip, .cbr, .rar, .pdf", ext)
		}
	}
	if err != nil {
		return nil, err
	}

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
				// Decode image
				src, _, err := image.Decode(img.Reader)
				if err != nil {
					bar.Clear()
					fmt.Fprintf(os.Stderr, "error processing image %s%s: %s\n", img.Path, img.Name, err)
					os.Exit(1)
				}

				if o.Image.Crop {
					g := gift.New(gift.Crop(findMarging(src)))
					newSrc := image.NewNRGBA(g.Bounds(src.Bounds()))
					g.Draw(newSrc, src)
					src = newSrc
				}

				g := epubimage.NewGift(o.Image)

				// Convert image
				dst := image.NewGray(g.Bounds(src.Bounds()))
				g.Draw(dst, src)

				var raw image.Image
				if img.Id == 0 {
					raw = dst
				}

				imageOutput <- &epubimage.Image{
					Id:      img.Id,
					Part:    0,
					Raw:     raw,
					Data:    epubimagedata.New(img.Id, 0, dst, o.Image.Quality),
					Width:   dst.Bounds().Dx(),
					Height:  dst.Bounds().Dy(),
					IsCover: img.Id == 0,
					DoublePage: src.Bounds().Dx() > src.Bounds().Dy() &&
						src.Bounds().Dx() > o.Image.ViewHeight &&
						src.Bounds().Dy() > o.Image.ViewWidth,
					Path: img.Path,
					Name: img.Name,
				}

				// Auto split double page
				// Except for cover
				// Only if the src image have width > height and is bigger than the view
				if (!o.Image.HasCover || img.Id > 0) &&
					o.Image.AutoSplitDoublePage &&
					src.Bounds().Dx() > src.Bounds().Dy() &&
					src.Bounds().Dx() > o.Image.ViewHeight &&
					src.Bounds().Dy() > o.Image.ViewWidth {
					gifts := epubimage.NewGiftSplitDoublePage(o.Image)
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
		return nil, fmt.Errorf("image not found")
	}

	return images, nil
}

func LoadCoverData(img *epubimage.Image, title string, quality int) *epubimagedata.ImageData {
	// Create a blur version of the cover
	g := gift.New(epubfilters.CoverTitle(title))
	dst := image.NewGray(g.Bounds(img.Raw.Bounds()))
	g.Draw(dst, img.Raw)

	return epubimagedata.NewRaw("OEBPS/Images/title.jpg", dst, quality)
}
