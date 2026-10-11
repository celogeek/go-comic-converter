// Package epubimage EPUBImage helpers to transform image.
package epubimage

import (
	"image"

	"github.com/celogeek/go-comic-converter/v3/internal/pkg/utils"
)

// Maximum size of an image displayed as is (SR profile).
//
// A bigger image is stored as a grid of tiles, displayed side by side in the same page,
// so each file stays within what the device can handle without downscaling.
const (
	MaxWidth  = 1200
	MaxHeight = 1920
)

type EPUBImage struct {
	Id         int
	Part       int
	Raw        image.Image
	Width      int
	Height     int
	IsBlank    bool
	DoublePage bool
	// Cols and Rows tell the image is stored in Cols x Rows files (see TileImgPath), row by row.
	// Width and Height are the size of the whole image. 0 is the same as 1.
	Cols, Rows          int
	Path                string
	Name                string
	Format              string
	OriginalAspectRatio float64
	Error               error
}

// SpaceKey key name of the blank page after the image
func (i EPUBImage) SpaceKey() string {
	return "space_" + utils.IntToString(i.Id)
}

// SpacePath path of the blank page
func (i EPUBImage) SpacePath() string {
	return "Text/" + i.SpaceKey() + ".xhtml"
}

// EPUBSpacePath path of the blank page into the EPUB
func (i EPUBImage) EPUBSpacePath() string {
	return "OEBPS/" + i.SpacePath()
}

func (i EPUBImage) PartKey() string {
	return utils.IntToString(i.Id) + "_p" + utils.IntToString(i.Part)
}

// PageKey key for page
func (i EPUBImage) PageKey() string {
	return "page_" + i.PartKey()
}

// PagePath page path linked to the image
func (i EPUBImage) PagePath() string {
	return "Text/" + i.PageKey() + ".xhtml"
}

// EPUBPagePath page path into the EPUB
func (i EPUBImage) EPUBPagePath() string {
	return "OEBPS/" + i.PagePath()
}

// ImgKey key for image
func (i EPUBImage) ImgKey() string {
	return "img_" + i.PartKey()
}

// ImgPath image path
func (i EPUBImage) ImgPath() string {
	return "Images/" + i.ImgKey() + "." + i.Format
}

// EPUBImgPath image path into the EPUB
func (i EPUBImage) EPUBImgPath() string {
	return "OEBPS/" + i.ImgPath()
}

// NeedStack tells if an image of this size is too big to be stored in one file.
func NeedStack(width, height int) bool {
	return height > 1 && (width > MaxWidth || height > MaxHeight)
}

// FitSize size of an image reduced to the max size, keeping its aspect ratio.
// An image within the max size is returned as is.
func FitSize(width, height int) (int, int) {
	if width <= MaxWidth && height <= MaxHeight {
		return width, height
	}

	wratio := float64(width) / MaxWidth
	hratio := float64(height) / MaxHeight
	if wratio > hratio {
		return MaxWidth, max(1, min(MaxHeight, int(float64(height)/wratio+0.5)))
	}

	return max(1, min(MaxWidth, int(float64(width)/hratio+0.5))), MaxHeight
}

// Grid number of columns and rows needed to keep each tile of an image of this size within the max size.
func Grid(width, height int) (cols, rows int) {
	cols, rows = 1, 1
	if NeedStack(width, height) {
		cols = (width + MaxWidth - 1) / MaxWidth
		rows = (height + MaxHeight - 1) / MaxHeight
	}
	return
}

// Tiles number of files used to store the image.
func (i EPUBImage) Tiles() int {
	return max(i.Cols, 1) * max(i.Rows, 1)
}

// cut position of the n-th cut of a side of the given size, scaled to the displayed size.
//
// The image is cut every max pixels, the last piece being the remainder.
func cut(size, max, n, displayed int) int {
	c := min(n*max, size)
	return (c*displayed + size/2) / size
}

// tileRect area of the tile idx, scaled to an image of the given size.
//
// The image is cut every MaxWidth and every MaxHeight, so each tile is at the max size except the last ones.
func (i EPUBImage) tileRect(idx, width, height int) image.Rectangle {
	cols := max(i.Cols, 1)
	c, r := idx%cols, idx/cols

	return image.Rect(
		cut(i.Width, MaxWidth, c, width), cut(i.Height, MaxHeight, r, height),
		cut(i.Width, MaxWidth, c+1, width), cut(i.Height, MaxHeight, r+1, height),
	)
}

// TileBounds area of the tile idx into the raw image.
func (i EPUBImage) TileBounds(idx int, b image.Rectangle) image.Rectangle {
	return i.tileRect(idx, i.Width, i.Height).Add(b.Min)
}

// Tiled image of the given size, with the grid needed to store it.
func Tiled(width, height int) EPUBImage {
	cols, rows := Grid(width, height)
	return EPUBImage{Width: width, Height: height, Cols: cols, Rows: rows}
}

// TileName name of the tile idx of an image, the first one being the image itself
func TileName(name string, idx int) string {
	if idx == 0 {
		return name
	}
	return name + "_" + utils.IntToString(idx)
}

// TileImgKey key for the tile idx of the image, the first one being the image itself
func (i EPUBImage) TileImgKey(idx int) string {
	return TileName(i.ImgKey(), idx)
}

// TileImgPath image path of the tile idx
func (i EPUBImage) TileImgPath(idx int) string {
	return "Images/" + i.TileImgKey(idx) + "." + i.Format
}

// EPUBTileImgPath image path of the tile idx into the EPUB
func (i EPUBImage) EPUBTileImgPath(idx int) string {
	return "OEBPS/" + i.TileImgPath(idx)
}

// MediaType of the epub image
func (i EPUBImage) MediaType() string {
	return "image/" + i.Format
}

// ImgStyle style to apply to the image.
//
// Always centered: the same page is shown alone in portrait and in a spread in landscape,
// and Kindle devices handle only top/left reliably.
func (i EPUBImage) ImgStyle(viewWidth, viewHeight int) string {
	relWidth, relHeight := i.RelSize(viewWidth, viewHeight)
	marginW, marginH := margins(viewWidth, viewHeight, relWidth, relHeight)

	return imgStyle(relWidth, relHeight, marginW, marginH)
}

// margins offset of an image centered into the view, in whole pixels.
func margins(viewWidth, viewHeight, width, height int) (left, top int) {
	return (viewWidth - width) / 2, (viewHeight - height) / 2
}

// ImgStyles style to apply to each file of the image, tile by tile,
// so they are displayed as one image.
func (i EPUBImage) ImgStyles(viewWidth, viewHeight int) []string {
	relWidth, relHeight := i.RelSize(viewWidth, viewHeight)
	marginW, marginH := margins(viewWidth, viewHeight, relWidth, relHeight)

	// whole pixels, so each tile starts exactly where the previous one ends (no seam)
	styles := make([]string, i.Tiles())
	for idx := range styles {
		r := i.tileRect(idx, relWidth, relHeight)
		styles[idx] = imgStyle(r.Dx(), r.Dy(), marginW+r.Min.X, marginH+r.Min.Y)
	}

	return styles
}

// PanelZoom size of the image zoomed by the panel view, and the style of each of its files (tile by tile).
//
// The zoom is relative to the size of the image displayed in the page.
func (i EPUBImage) PanelZoom(viewWidth, viewHeight int, zoom float64) (string, []string) {
	relWidth, relHeight := i.RelSize(viewWidth, viewHeight)
	width, height := int(float64(relWidth)*zoom+0.5), int(float64(relHeight)*zoom+0.5)
	box := "width:" + utils.IntToString(width) + "px; height:" + utils.IntToString(height) + "px"

	styles := make([]string, i.Tiles())
	for idx := range styles {
		r := i.tileRect(idx, width, height)
		styles[idx] = imgStyle(r.Dx(), r.Dy(), r.Min.X, r.Min.Y)
	}

	return box, styles
}

func imgStyle(width, height, left, top int) string {
	return "width:" + utils.IntToString(width) + "px; height:" + utils.IntToString(height) +
		"px; top:" + utils.IntToString(top) + "px; left:" + utils.IntToString(left) + "px"
}

func (i EPUBImage) RelSize(viewWidth, viewHeight int) (relWidth, relHeight int) {
	w, h := viewWidth, viewHeight
	srcw, srch := i.Width, i.Height

	if w <= 0 || h <= 0 || srcw <= 0 || srch <= 0 {
		return
	}

	wratio := float64(srcw) / float64(w)
	hratio := float64(srch) / float64(h)

	if wratio > hratio {
		relWidth = w
		relHeight = int(float64(srch)/wratio + 0.5)
	} else {
		relHeight = h
		relWidth = int(float64(srcw)/hratio + 0.5)
	}

	return
}
