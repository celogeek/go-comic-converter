// Package epubimage EPUBImage helpers to transform image.
package epubimage

import (
	"image"

	"github.com/celogeek/go-comic-converter/v3/internal/pkg/utils"
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
	Cols, Rows int
	// Tile is the max size of each file of the image, unset when the device has no constraint.
	Tile                TileSize
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

// Tiles number of files used to store the image.
func (i EPUBImage) Tiles() int {
	return max(i.Cols, 1) * max(i.Rows, 1)
}

// cut position of the n-th cut of a side of the given size, scaled to the displayed size.
//
// The side is cut every step pixels (see TileSize.step), the last piece being the remainder.
func cut(size, step, n, displayed int) int {
	c := min(n*step, size)
	return (c*displayed + size/2) / size
}

// edge position of the n-th cut of a side cut into count pieces, scaled to the displayed size.
//
// The last piece always ends at the end of the side, even when the image is not tiled
// (Cols and Rows unset, as for an image copied as is) and bigger than the max size.
func edge(size, step, n, count, displayed int) int {
	if n >= count {
		return displayed
	}
	return cut(size, step, n, displayed)
}

// tileRect area of the tile idx, scaled to an image of the given size.
//
// The tiles of a side have the same length, a multiple of 16 (see TileSize.step), except the last one.
func (i EPUBImage) tileRect(idx, width, height int) image.Rectangle {
	cols, rows := max(i.Cols, 1), max(i.Rows, 1)
	c, r := idx%cols, idx/cols
	stepW := i.Tile.step(i.Width, cols, i.Tile.Width)
	stepH := i.Tile.step(i.Height, rows, i.Tile.Height)

	return image.Rect(
		edge(i.Width, stepW, c, cols, width), edge(i.Height, stepH, r, rows, height),
		edge(i.Width, stepW, c+1, cols, width), edge(i.Height, stepH, r+1, rows, height),
	)
}

// TileBounds area of the tile idx into the raw image.
func (i EPUBImage) TileBounds(idx int, b image.Rectangle) image.Rectangle {
	return i.tileRect(idx, i.Width, i.Height).Add(b.Min)
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
