package epubimageprocessor

import (
	"fmt"
	"image"
	"image/color"
	"image/jpeg"
	"os"
	"path/filepath"
	"sort"
	"testing"

	"github.com/celogeek/go-comic-converter/v3/internal/pkg/fonts"
	"github.com/celogeek/go-comic-converter/v3/pkg/epuboptions"
)

const halfW, halfH = 60, 90

// spread is a double page: the content goes through the middle.
func spread() *image.Gray {
	img := image.NewGray(image.Rect(0, 0, halfW*2, halfH))
	for y := range halfH {
		for x := range halfW * 2 {
			img.SetGray(x, y, color.Gray{Y: uint8((x*2 + y*3) % 200)})
		}
	}
	return img
}

// other is a page that doesn't continue the spread.
func other() *image.Gray {
	img := image.NewGray(image.Rect(0, 0, halfW, halfH))
	for y := range halfH {
		for x := range halfW {
			img.SetGray(x, y, color.Gray{Y: uint8(255 - (x*7+y)%100)})
		}
	}
	return img
}

// inverse is a page of the same kind as other, but with a very different edge.
func inverse() *image.Gray {
	o := other()
	for i, v := range o.Pix {
		o.Pix[i] = 255 - v
	}
	return o
}

// solid is a page with a solid edge, like a black page.
func solid(v uint8) *image.Gray {
	img := image.NewGray(image.Rect(0, 0, halfW, halfH))
	for i := range img.Pix {
		img.Pix[i] = v
	}
	return img
}

func writeJPEG(t *testing.T, dir string, i int, img image.Image) {
	t.Helper()
	f, err := os.Create(filepath.Join(dir, fmt.Sprintf("%03d.jpg", i)))
	if err != nil {
		t.Fatal(err)
	}
	defer f.Close()
	if err := jpeg.Encode(f, img, &jpeg.Options{Quality: 100}); err != nil {
		t.Fatal(err)
	}
}

func TestCombineSplitDoublePage(t *testing.T) {
	s := spread()
	leftHalf := s.SubImage(image.Rect(0, 0, halfW, halfH))
	rightHalf := s.SubImage(image.Rect(halfW, 0, halfW*2, halfH))

	for _, c := range []struct {
		name    string
		manga   bool
		combine bool
		pages   []image.Image
		want    []int // ids kept
		wide    []int // ids that are combined (double page)
	}{
		{"left to right", false, true, []image.Image{other(), leftHalf, rightHalf, other()}, []int{0, 1, 3}, []int{1}},
		{"manga", true, true, []image.Image{other(), rightHalf, leftHalf, other()}, []int{0, 1, 3}, []int{1}},
		{"wrong direction", false, true, []image.Image{other(), rightHalf, leftHalf, other()}, []int{0, 1, 2, 3}, nil},
		{"unrelated pages", false, true, []image.Image{other(), inverse(), leftHalf, other()}, []int{0, 1, 2, 3}, nil},
		{"solid black pages", false, true, []image.Image{other(), solid(13), solid(13), other()}, []int{0, 1, 2, 3}, nil},
		{"off by default", false, false, []image.Image{other(), leftHalf, rightHalf, other()}, []int{0, 1, 2, 3}, nil},
		{"3 halves", false, true, []image.Image{leftHalf, rightHalf, leftHalf}, []int{0, 2}, []int{0}},
	} {
		t.Run(c.name, func(t *testing.T) {
			dir := filepath.Join(t.TempDir(), "in")
			if err := os.Mkdir(dir, 0o755); err != nil {
				t.Fatal(err)
			}
			for i, p := range c.pages {
				writeJPEG(t, dir, i, p)
			}

			images, err := New(epuboptions.EPUBOptions{
				Input:        dir,
				Output:       filepath.Join(t.TempDir(), "book.epub"),
				Workers:      4,
				Quiet:        true,
				TrueTypeFont: fonts.Default,
				Image: epuboptions.Image{
					Format:                 "jpeg",
					Quality:                90,
					Manga:                  c.manga,
					DoublePageMode:         epuboptions.DoublePageRotate,
					CombineSplitDoublePage: c.combine,
				},
			}).Load()
			if err != nil {
				t.Fatal(err)
			}

			sort.Slice(images, func(i, j int) bool { return images[i].Id < images[j].Id })
			var got, wide []int
			for _, img := range images {
				got = append(got, img.Id)
				if img.DoublePage {
					wide = append(wide, img.Id)
				}
			}
			if fmt.Sprint(got) != fmt.Sprint(c.want) {
				t.Errorf("got images %v, want %v", got, c.want)
			}
			if fmt.Sprint(wide) != fmt.Sprint(c.wide) {
				t.Errorf("got double pages %v, want %v", wide, c.wide)
			}
		})
	}
}

// In split mode, the combined spread is cut in 2 parts of the same size, whatever the crop of each source page.
func TestCombineSplitDoublePageSplitMode(t *testing.T) {
	s := spread()
	dir := filepath.Join(t.TempDir(), "in")
	if err := os.Mkdir(dir, 0o755); err != nil {
		t.Fatal(err)
	}
	writeJPEG(t, dir, 0, s.SubImage(image.Rect(0, 0, halfW, halfH)))
	writeJPEG(t, dir, 1, s.SubImage(image.Rect(halfW, 0, halfW*2, halfH)))

	images, err := New(epuboptions.EPUBOptions{
		Input:        dir,
		Output:       filepath.Join(t.TempDir(), "book.epub"),
		Workers:      4,
		Quiet:        true,
		TrueTypeFont: fonts.Default,
		Image: epuboptions.Image{
			Format:                    "jpeg",
			Quality:                   90,
			DoublePageMode:            epuboptions.DoublePageSplit,
			KeepSplitDoublePageAspect: true,
			CombineSplitDoublePage:    true,
		},
	}).Load()
	if err != nil {
		t.Fatal(err)
	}

	var parts, widths []int
	for _, img := range images {
		if img.Part > 0 {
			parts = append(parts, img.Part)
			widths = append(widths, img.Width)
		}
	}
	sort.Ints(parts)
	if fmt.Sprint(parts) != "[1 2]" {
		t.Fatalf("got parts %v, want [1 2]", parts)
	}
	if widths[0] != widths[1] || widths[0] != halfW {
		t.Errorf("got widths %v, want 2 parts of %d", widths, halfW)
	}
}
