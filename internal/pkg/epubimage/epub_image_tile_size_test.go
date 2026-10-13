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
		w, h, cols, rows int
	}{
		{1200, 1920, 1, 1},
		{1860, 2480, 2, 2},
		{1200, 3000, 1, 2},
		{2400, 1920, 2, 1},
		{3000, 1, 1, 1},
	} {
		cols, rows := TileKindle.Grid(c.w, c.h)
		if cols != c.cols || rows != c.rows {
			t.Errorf("Grid(%d, %d) = %d, %d, want %d, %d", c.w, c.h, cols, rows, c.cols, c.rows)
		}
		// every tile stays within the limit
		e := EPUBImage{Width: c.w, Height: c.h, Cols: cols, Rows: rows, Tile: TileKindle}
		for idx := range e.Tiles() {
			r := e.tileRect(idx, c.w, c.h)
			if cols*rows > 1 && (r.Dx() > TileKindle.Width || r.Dy() > TileKindle.Height) {
				t.Errorf("tile %d of %dx%d is %dx%d", idx, c.w, c.h, r.Dx(), r.Dy())
			}
			// tiles are at the max size, only the last column / row is smaller
			wantW := TileKindle.Width
			if idx%cols == cols-1 {
				wantW = c.w - (idx%cols)*TileKindle.Width
			}
			if r.Dx() != wantW {
				t.Errorf("tile %d of %dx%d has width %d, want %d", idx, c.w, c.h, r.Dx(), wantW)
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
		// the native iPad size stays in one file
		{1640, 2360, 1, 1},
		{2366, 2366, 1, 1},
		{2367, 2366, 2, 1},
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
