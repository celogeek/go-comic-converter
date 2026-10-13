package epubimage

import "testing"

func TestNeedStack(t *testing.T) {
	for _, c := range []struct {
		w, h int
		want bool
	}{
		{1200, 1920, false},
		{1201, 1920, true},
		{1200, 1921, true},
		{1860, 2480, true},
		{600, 800, false},
		{3000, 1, false},
	} {
		if got := TileKindle.NeedStack(c.w, c.h); got != c.want {
			t.Errorf("NeedStack(%d, %d) = %v, want %v", c.w, c.h, got, c.want)
		}
	}
}

func TestGrid(t *testing.T) {
	for _, c := range []struct {
		tile             TileSize
		w, h, cols, rows int
	}{
		{TileKindle, 1200, 1920, 1, 1},
		{TileKindle, 1860, 2480, 2, 2},
		{TileKindle, 1200, 3000, 1, 2},
		{TileKindle, 2400, 1920, 2, 1},
		{TileKindle, 3000, 1, 1, 1},
		{TileKindle, 3601, 5761, 4, 4},
		{TileApple, 2460, 3540, 2, 2},
		{TileApple, 2064, 2752, 1, 2},
	} {
		cols, rows := c.tile.Grid(c.w, c.h)
		if cols != c.cols || rows != c.rows {
			t.Errorf("Grid(%d, %d) = %d, %d, want %d, %d", c.w, c.h, cols, rows, c.cols, c.rows)
		}

		// tiles are contiguous, in the limit, and balanced: only the last column / row is shorter,
		// every cut is a multiple of 16 (a jpeg block)
		e := EPUBImage{Width: c.w, Height: c.h, Cols: cols, Rows: rows, Tile: c.tile}
		first := e.tileRect(0, c.w, c.h)
		for idx := range e.Tiles() {
			r := e.tileRect(idx, c.w, c.h)
			col, row := idx%cols, idx/cols
			if cols*rows > 1 && (r.Dx() > c.tile.Width || r.Dy() > c.tile.Height) {
				t.Errorf("tile %d of %dx%d is %dx%d", idx, c.w, c.h, r.Dx(), r.Dy())
			}
			if (col == 0 && r.Min.X != 0) || (col > 0 && r.Min.X != e.tileRect(idx-1, c.w, c.h).Max.X) ||
				(row == 0 && r.Min.Y != 0) || (row > 0 && r.Min.Y != e.tileRect(idx-cols, c.w, c.h).Max.Y) {
				t.Errorf("tile %d of %dx%d is not contiguous: %v", idx, c.w, c.h, r)
			}
			if col == cols-1 && r.Max.X != c.w || row == rows-1 && r.Max.Y != c.h {
				t.Errorf("tile %d of %dx%d ends at %v, not at the end of the image", idx, c.w, c.h, r.Max)
			}
			if col < cols-1 && (r.Dx() != first.Dx() || r.Dx()%16 != 0) {
				t.Errorf("tile %d of %dx%d has width %d, want %d, a multiple of 16", idx, c.w, c.h, r.Dx(), first.Dx())
			}
			if row < rows-1 && (r.Dy() != first.Dy() || r.Dy()%16 != 0) {
				t.Errorf("tile %d of %dx%d has height %d, want %d, a multiple of 16", idx, c.w, c.h, r.Dy(), first.Dy())
			}
			if col == cols-1 && cols > 1 && (r.Dx() > first.Dx() || first.Dx()-r.Dx() >= 16*cols) {
				t.Errorf("last column of %dx%d has width %d, first one has %d", c.w, c.h, r.Dx(), first.Dx())
			}
			if row == rows-1 && rows > 1 && (r.Dy() > first.Dy() || first.Dy()-r.Dy() >= 16*rows) {
				t.Errorf("last row of %dx%d has height %d, first one has %d", c.w, c.h, r.Dy(), first.Dy())
			}
		}
	}

	// exact cuts
	for _, c := range []struct {
		tile  TileSize
		w, h  int
		sizes [][2]int
	}{
		{TileKindle, 1860, 2480, [][2]int{{944, 1248}, {916, 1248}, {944, 1232}, {916, 1232}}},
		{TileKindle, 1200, 3000, [][2]int{{1200, 1504}, {1200, 1496}}},
		{TileApple, 2460, 3540, [][2]int{{1232, 1776}, {1228, 1776}, {1232, 1764}, {1228, 1764}}},
	} {
		cols, rows := c.tile.Grid(c.w, c.h)
		e := EPUBImage{Width: c.w, Height: c.h, Cols: cols, Rows: rows, Tile: c.tile}
		for idx, want := range c.sizes {
			if r := e.tileRect(idx, c.w, c.h); r.Dx() != want[0] || r.Dy() != want[1] {
				t.Errorf("tile %d of %dx%d is %dx%d, want %dx%d", idx, c.w, c.h, r.Dx(), r.Dy(), want[0], want[1])
			}
		}
	}
}

func TestFitSize(t *testing.T) {
	for _, c := range []struct {
		w, h, wantW, wantH int
	}{
		{1200, 1920, 1200, 1920},
		{600, 800, 600, 800},
		{1920, 3005, 1200, 1878},
		{2400, 3840, 1200, 1920},
		{3000, 1000, 1200, 400},
		{1000, 5000, 384, 1920},
	} {
		w, h := TileKindle.FitSize(c.w, c.h)
		if w != c.wantW || h != c.wantH {
			t.Errorf("FitSize(%d, %d) = %d, %d, want %d, %d", c.w, c.h, w, h, c.wantW, c.wantH)
		}
	}
}

// the tile size follows the device: the same image is tiled differently.
func TestTileSizeApple(t *testing.T) {
	for _, c := range []struct {
		w, h, cols, rows int
	}{
		// the native size of the iPad mini, iPad and iPad Pro 11" stays in one file
		{1488, 2266, 1, 1},
		{1640, 2360, 1, 1},
		{1668, 2420, 1, 1},
		{2288, 2432, 1, 1},
		{2289, 2432, 2, 1},
		{2288, 2433, 1, 2},
		// the iPad Pro 13" is over 5.6 million pixels
		{2064, 2752, 1, 2},
		{2460, 3540, 2, 2},
	} {
		cols, rows := TileApple.Grid(c.w, c.h)
		if cols != c.cols || rows != c.rows {
			t.Errorf("Grid(%d, %d) = %d, %d, want %d, %d", c.w, c.h, cols, rows, c.cols, c.rows)
		}
		e := TileApple.Tiled(c.w, c.h)
		for idx := range e.Tiles() {
			if r := e.tileRect(idx, c.w, c.h); r.Dx()*r.Dy() > 5_600_000 {
				t.Errorf("tile %d of %dx%d has %d pixels, Apple Books accepts 5.6M", idx, c.w, c.h, r.Dx()*r.Dy())
			}
		}
	}

	// the zero value has no constraint: never tiled, never reduced
	var zero TileSize
	if zero.Limited() || zero.NeedStack(5000, 8000) {
		t.Error("zero tile size must not be limited")
	}
	if cols, rows := zero.Grid(5000, 8000); cols != 1 || rows != 1 {
		t.Errorf("zero tile size: got %d x %d, want 1 x 1", cols, rows)
	}
	if w, h := zero.FitSize(5000, 8000); w != 5000 || h != 8000 {
		t.Errorf("zero tile size: got %dx%d, want 5000x8000", w, h)
	}
	if e := zero.Tiled(5000, 8000); e.Tiles() != 1 {
		t.Errorf("zero tile size: got %d tiles, want 1", e.Tiles())
	}
	if got := TileApple.Width * TileApple.Height; got >= 5_600_000 {
		t.Errorf("apple tile has %d pixels", got)
	}
}
