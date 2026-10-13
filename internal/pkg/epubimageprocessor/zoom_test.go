package epubimageprocessor

import "testing"

func TestZoomFactor(t *testing.T) {
	for _, c := range []struct {
		name                        string
		width, height, viewW, viewH int
		zoom, want                  float64
	}{
		{"no zoom", 3000, 4000, 1000, 1500, 1, 1},
		{"zoom", 3000, 4500, 1000, 1500, 1.5, 1.5},
		{"zoom limited by the size of the image", 1500, 2250, 1000, 1500, 2, 1.5},
		{"image as the view", 1000, 1500, 1000, 1500, 2, 1},
		{"image smaller than the view", 500, 750, 1000, 1500, 2, 1},
		{"limited by the width", 3000, 3000, 1000, 1500, 3, 3},
		{"invalid size", 0, 0, 1000, 1500, 2, 1},
	} {
		t.Run(c.name, func(t *testing.T) {
			if got := zoomFactor(c.width, c.height, c.viewW, c.viewH, c.zoom); got != c.want {
				t.Errorf("got %v, want %v", got, c.want)
			}
		})
	}
}
