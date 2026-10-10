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

	stacked := EPUBImage{Width: 2000, Height: 3001, Stacked: true}
	got := stacked.ImgStyles(1000, 1500)
	want := []string{
		"width:1000px; height:750px; top:0px; left:0px",
		"width:1000px; height:750px; top:750px; left:0px",
	}
	if !reflect.DeepEqual(got, want) {
		t.Errorf("stacked image: got %v, want %v", got, want)
	}
}
