package epubimage

import (
	"reflect"
	"testing"
)

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
		if got := NeedStack(c.w, c.h); got != c.want {
			t.Errorf("NeedStack(%d, %d) = %v, want %v", c.w, c.h, got, c.want)
		}
	}
}

func TestImgStyles(t *testing.T) {
	single := EPUBImage{Width: 1000, Height: 1500}
	if got := single.ImgStyles(1000, 1500); !reflect.DeepEqual(got, []string{single.ImgStyle(1000, 1500)}) {
		t.Errorf("single image: got %v", got)
	}

	// an image copied as is has no grid (Cols and Rows unset): it must fill the same area as a single image
	for _, c := range []EPUBImage{
		{Width: 3000, Height: 4000},
		{Width: 1500, Height: 2400},
		{Width: 1000, Height: 1500},
	} {
		got := c.ImgStyles(1200, 1920)
		if want := []string{c.ImgStyle(1200, 1920)}; !reflect.DeepEqual(got, want) {
			t.Errorf("image %dx%d without grid: got %v, want %v", c.Width, c.Height, got, want)
		}
	}

	// 2 columns x 2 rows, cut every 1200 x 1920 of the full size image
	tiled := EPUBImage{Width: 2000, Height: 3001, Cols: 2, Rows: 2}
	got := tiled.ImgStyles(1000, 1500)
	want := []string{
		"width:600px; height:960px; top:0px; left:0px",
		"width:400px; height:960px; top:0px; left:600px",
		"width:600px; height:540px; top:960px; left:0px",
		"width:400px; height:540px; top:960px; left:600px",
	}
	if !reflect.DeepEqual(got, want) {
		t.Errorf("tiled image: got %v, want %v", got, want)
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
		cols, rows := Grid(c.w, c.h)
		if cols != c.cols || rows != c.rows {
			t.Errorf("Grid(%d, %d) = %d, %d, want %d, %d", c.w, c.h, cols, rows, c.cols, c.rows)
		}
		// every tile stays within the limit
		e := EPUBImage{Width: c.w, Height: c.h, Cols: cols, Rows: rows}
		for idx := range e.Tiles() {
			r := e.tileRect(idx, c.w, c.h)
			if cols*rows > 1 && (r.Dx() > MaxWidth || r.Dy() > MaxHeight) {
				t.Errorf("tile %d of %dx%d is %dx%d", idx, c.w, c.h, r.Dx(), r.Dy())
			}
			// tiles are at the max size, only the last column / row is smaller
			wantW := MaxWidth
			if idx%cols == cols-1 {
				wantW = c.w - (idx%cols)*MaxWidth
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
		w, h := FitSize(c.w, c.h)
		if w != c.wantW || h != c.wantH {
			t.Errorf("FitSize(%d, %d) = %d, %d, want %d, %d", c.w, c.h, w, h, c.wantW, c.wantH)
		}
	}
}
