package epubimageprocessor

import (
	"archive/zip"
	"fmt"
	"image"
	"image/color"
	"image/draw"
	"image/jpeg"
	"os"
	"path/filepath"
	"reflect"
	"sort"
	"testing"

	"github.com/celogeek/go-comic-converter/v3/internal/pkg/fonts"
	"github.com/celogeek/go-comic-converter/v3/internal/pkg/utils"
	"github.com/celogeek/go-comic-converter/v3/pkg/epuboptions"
)

// An entry that cannot be opened (unsupported compression method) must be reported as a corrupted image.
func TestLoadCbzUnreadableEntry(t *testing.T) {
	path := filepath.Join(t.TempDir(), "book.cbz")
	if err := writeBrokenZip(path); err != nil {
		t.Fatal(err)
	}

	e := ePUBImageProcessor{epuboptions.EPUBOptions{Input: path, TrueTypeFont: fonts.Default}}
	total, output, err := e.loadCbz()
	if err != nil {
		t.Fatal(err)
	}
	if total != 1 {
		t.Fatalf("got %d images, want 1", total)
	}

	n := 0
	for task := range output {
		n++
		if task.Error == nil {
			t.Errorf("%s: expected an error", task.Name)
		}
		if task.Image == nil {
			t.Errorf("%s: expected the corrupted image", task.Name)
		}
	}
	if n != 1 {
		t.Errorf("got %d tasks, want 1", n)
	}
}

func writeBrokenZip(path string) error {
	f, err := os.Create(path)
	if err != nil {
		return err
	}
	defer f.Close()

	w := zip.NewWriter(f)
	// method 99 is not registered: Open() fails
	fw, err := w.CreateRaw(&zip.FileHeader{Name: "001.jpg", Method: 99, CompressedSize64: 4, UncompressedSize64: 4})
	if err != nil {
		return err
	}
	if _, err = fw.Write([]byte("junk")); err != nil {
		return err
	}
	return w.Close()
}

// writePages writes jpeg pages named 001.jpg, 002.jpg... "" is a blank page, "wide" a double page.
func writePages(t *testing.T, dir string, pages ...string) {
	t.Helper()
	for i, kind := range pages {
		w, h := 60, 90
		if kind == "wide" {
			w, h = 180, 90
		}
		img := image.NewGray(image.Rect(0, 0, w, h))
		draw.Draw(img, img.Bounds(), image.NewUniform(color.White), image.Point{}, draw.Src)
		if kind != "" {
			draw.Draw(img, image.Rect(10, 10, w-10, h-10), image.NewUniform(color.Black), image.Point{}, draw.Src)
		}
		f, err := os.Create(filepath.Join(dir, fmt.Sprintf("%03d.jpg", i+1)))
		if err != nil {
			t.Fatal(err)
		}
		if err := jpeg.Encode(f, img, nil); err != nil {
			t.Fatal(err)
		}
		_ = f.Close()
	}
}

// The cover is the first image that is not blank, and the only one that keeps its raw data.
func TestLoadCover(t *testing.T) {
	for _, c := range []struct {
		name      string
		pages     []string
		wantCover [2]int // id, part
		want      [][2]int
	}{
		{"blank first", []string{"", "page", "page"}, [2]int{1, 0}, [][2]int{{1, 0}, {2, 0}}},
		{"blank first twice", []string{"", "", "page"}, [2]int{2, 0}, [][2]int{{2, 0}}},
		{"blank then double page", []string{"", "wide", "page"}, [2]int{1, 0}, [][2]int{{1, 0}, {1, 1}, {1, 2}, {2, 0}}},
		{"double page first", []string{"wide", "wide", "page"}, [2]int{0, 0}, [][2]int{{0, 0}, {0, 1}, {0, 2}, {1, 1}, {1, 2}, {2, 0}}},
		{"no blank", []string{"page", "page"}, [2]int{0, 0}, [][2]int{{0, 0}, {1, 0}}},
	} {
		t.Run(c.name, func(t *testing.T) {
			dir := filepath.Join(t.TempDir(), "in")
			if err := os.Mkdir(dir, 0o755); err != nil {
				t.Fatal(err)
			}
			writePages(t, dir, c.pages...)

			e := New(epuboptions.EPUBOptions{
				Input:        dir,
				Output:       filepath.Join(t.TempDir(), "book.epub"),
				Workers:      4,
				TrueTypeFont: fonts.Default,
				Image:        epuboptions.Image{Format: "jpeg", Quality: 90, NoBlankImage: true},
			})
			images, err := e.Load()
			if err != nil {
				t.Fatal(err)
			}

			sort.Slice(images, func(i, j int) bool {
				if images[i].Id == images[j].Id {
					return images[i].Part < images[j].Part
				}
				return images[i].Id < images[j].Id
			})

			var got [][2]int
			for i, img := range images {
				got = append(got, [2]int{img.Id, img.Part})
				if hasRaw := img.Raw != nil; hasRaw != (i == 0) {
					t.Errorf("image %d part %d: raw data kept = %v", img.Id, img.Part, hasRaw)
				}
			}
			if !reflect.DeepEqual(got, c.want) {
				t.Errorf("got images %v, want %v", got, c.want)
			}
			if got[0] != c.wantCover {
				t.Errorf("got cover %v, want %v", got[0], c.wantCover)
			}
		})
	}
}

// The corrupted image is a full size page with the name of the file in a box, in the colors of the view.
func TestCorruptedImage(t *testing.T) {
	opts := epuboptions.EPUBOptions{TrueTypeFont: fonts.Default}
	opts.Image.ProfileColor = true
	opts.Image.View.Color = epuboptions.Color{Foreground: "FF0000", Background: "00FF00"}
	e := ePUBImageProcessor{opts}

	img := e.corruptedImage("Vol 1/Chapter 2/", "003.jpg")
	if b := img.Bounds(); b.Dx() != 1200 || b.Dy() != 1920 {
		t.Errorf("got %dx%d, want 1200x1920", b.Dx(), b.Dy())
	}

	// the page is the background color, the text box border and text are the foreground color
	if got := utils.ColorToHex(img.At(5, 5)); got != "00FF00" {
		t.Errorf("page color %s, want 00FF00", got)
	}
	foreground := 0
	for y := 0; y < img.Bounds().Dy(); y++ {
		for x := 0; x < img.Bounds().Dx(); x += 2 {
			if utils.ColorToHex(img.At(x, y)) == "FF0000" {
				foreground++
			}
		}
	}
	if foreground == 0 {
		t.Error("expected the text box to be drawn in the foreground color")
	}

	// grayscale keeps the page in gray
	opts.Image.GrayScale = true
	img = ePUBImageProcessor{opts}.corruptedImage("", "003.jpg")
	if r, g, b, _ := img.At(5, 5).RGBA(); r != g || g != b {
		t.Errorf("grayscale page color is not gray: %d %d %d", r, g, b)
	}
}
