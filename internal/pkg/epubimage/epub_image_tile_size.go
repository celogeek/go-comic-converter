package epubimage

// TileSize maximum size of the file an image is stored in.
//
// A bigger image is stored as a grid of tiles, displayed side by side in the same page,
// so each file stays within what the device can handle without downscaling.
// The zero value has no constraint: the image is never tiled nor reduced.
type TileSize struct {
	Width, Height int
}

var (
	// TileKindle is the SR resolution, what the Kindle handles without downscaling.
	TileKindle = TileSize{Width: 1200, Height: 1920}
	// TileApple is under the 5.6 million pixels Apple Books accepts for an image (5.56 million).
	// It is tall enough for the native page of the iPad mini, iPad, iPad Air and iPad Pro 11".
	// Each side is a multiple of 16, the size of a JPEG block, so a cut never splits one.
	TileApple = TileSize{Width: 2288, Height: 2432}
)

// Limited tells if the device has a constraint on the size of an image.
func (t TileSize) Limited() bool {
	return t.Width > 0 && t.Height > 0
}

// NeedStack tells if an image of this size is too big to be stored in one file.
func (t TileSize) NeedStack(width, height int) bool {
	return t.Limited() && height > 1 && (width > t.Width || height > t.Height)
}

// FitSize size of an image reduced to the max size, keeping its aspect ratio.
// An image within the max size, or without constraint, is returned as is.
func (t TileSize) FitSize(width, height int) (int, int) {
	if !t.Limited() || (width <= t.Width && height <= t.Height) {
		return width, height
	}

	wratio := float64(width) / float64(t.Width)
	hratio := float64(height) / float64(t.Height)
	if wratio > hratio {
		return t.Width, max(1, min(t.Height, int(float64(height)/wratio+0.5)))
	}

	return max(1, min(t.Width, int(float64(width)/hratio+0.5))), t.Height
}

// Grid number of columns and rows needed to keep each tile of an image of this size within the max size.
func (t TileSize) Grid(width, height int) (cols, rows int) {
	cols, rows = 1, 1
	if t.NeedStack(width, height) {
		cols = (width + t.Width - 1) / t.Width
		rows = (height + t.Height - 1) / t.Height
	}
	return
}

// Tiled image of the given size, with the grid needed to store it.
func (t TileSize) Tiled(width, height int) EPUBImage {
	cols, rows := t.Grid(width, height)
	return EPUBImage{Width: width, Height: height, Cols: cols, Rows: rows, Tile: t}
}
