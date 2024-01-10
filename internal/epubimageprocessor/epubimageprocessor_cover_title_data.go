package epubimageprocessor

import (
	"fmt"
	"image"
	"image/color"
	"image/draw"

	"github.com/celogeek/go-comic-converter/v2/internal/epubimagefilters"
	"github.com/celogeek/go-comic-converter/v2/internal/epubzip"
	"github.com/disintegration/gift"
)

type CoverTitleDataOptions struct {
	Src         image.Image
	Name        string
	Text        string
	Align       string
	PctWidth    int
	PctMargin   int
	MaxFontSize int
	BorderSize  int
}

// create a title page with the cover
func (e *EPUBImageProcessor) CoverTitleData(o *CoverTitleDataOptions) (*epubzip.EPUBZipImage, error) {
	// Create a blur version of the cover
	g := gift.New(epubimagefilters.CoverTitle(o.Text, o.Align, o.PctWidth, o.PctMargin, o.MaxFontSize, o.BorderSize))
	var dst draw.Image
	if o.Name == "cover" && e.options.Image.GrayScale {
		// 16 shade of gray
		dst = image.NewPaletted(o.Src.Bounds(), color.Palette{
			color.Gray{0x00},
			color.Gray{0x11},
			color.Gray{0x22},
			color.Gray{0x33},
			color.Gray{0x44},
			color.Gray{0x55},
			color.Gray{0x66},
			color.Gray{0x77},
			color.Gray{0x88},
			color.Gray{0x99},
			color.Gray{0xAA},
			color.Gray{0xBB},
			color.Gray{0xCC},
			color.Gray{0xDD},
			color.Gray{0xEE},
			color.Gray{0xFF},
		})
	} else {
		dst = e.createImage(o.Src, g.Bounds(o.Src.Bounds()))
	}
	g.Draw(dst, o.Src)

	return epubzip.CompressImage(
		fmt.Sprintf("OEBPS/Images/%s.%s", o.Name, e.options.Image.Format),
		e.options.Image.Format,
		dst,
		e.options.Image.Quality,
	)
}
