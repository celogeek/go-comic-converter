package epubimage

import (
	"reflect"
	"testing"
)

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
	tiled := EPUBImage{Width: 2000, Height: 3001, Cols: 2, Rows: 2, Tile: TileKindle}
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
