package epubimagepassthrough

import (
	"bytes"
	"errors"
	"image"
	"image/jpeg"
	"path/filepath"
	"testing"

	"github.com/gen2brain/jxl"

	"github.com/celogeek/go-comic-converter/v3/internal/pkg/epubzip"
	"github.com/celogeek/go-comic-converter/v3/internal/pkg/fonts"
	"github.com/celogeek/go-comic-converter/v3/pkg/epuboptions"
)

func TestCopyRawDataToStorage(t *testing.T) {
	var buf bytes.Buffer
	if err := jpeg.Encode(&buf, image.NewGray(image.Rect(0, 0, 10, 20)), nil); err != nil {
		t.Fatal(err)
	}
	data := buf.Bytes()

	storage, err := epubzip.NewStorageImageWriter(filepath.Join(t.TempDir(), "img.zip"), "jpeg")
	if err != nil {
		t.Fatal(err)
	}
	defer storage.Close()

	e := ePUBImagePassthrough{epuboptions.EPUBOptions{}}

	for _, c := range []struct {
		name, dirname, filename string
		wantPath, wantName      string
	}{
		{"archive root", "", "001.jpg", "", "001.jpg"},
		{"archive subdir", "", "Chapter1/001.jpg", "Chapter1/", "001.jpg"},
		{"archive nested", "", "Vol1/Chapter1/001.jpg", "Vol1/Chapter1/", "001.jpg"},
		{"dir root", "books", "books/001.jpg", "", "001.jpg"},
		{"dir subdir", "books", "books/ch1/001.jpg", "ch1/", "001.jpg"},
		{"uppercase extension", "", "Chapter1/001.JPG", "Chapter1/", "001.JPG"},
	} {
		t.Run(c.name, func(t *testing.T) {
			img, err := e.copyRawDataToStorage(
				storage,
				func() ([]byte, error) { return data, nil },
				1, c.dirname, c.filename,
			)
			if err != nil {
				t.Fatal(err)
			}
			if img.Path != c.wantPath || img.Name != c.wantName {
				t.Errorf("got path %q name %q, want path %q name %q", img.Path, img.Name, c.wantPath, c.wantName)
			}
			if img.Format != "jpeg" || img.Width != 10 || img.Height != 20 {
				t.Errorf("got %s %dx%d", img.Format, img.Width, img.Height)
			}
		})
	}

	t.Run("jxl", func(t *testing.T) {
		var jbuf bytes.Buffer
		if err := jxl.Encode(&jbuf, image.NewGray(image.Rect(0, 0, 30, 20))); err != nil {
			t.Fatal(err)
		}
		// the cover (id 0) is also decoded
		img, err := e.copyRawDataToStorage(storage, func() ([]byte, error) { return jbuf.Bytes(), nil }, 0, "", "ch1/000.JXL")
		if err != nil {
			t.Fatal(err)
		}
		if img.Format != "jxl" || img.Width != 30 || img.Height != 20 || !img.DoublePage || img.Raw == nil {
			t.Errorf("got %s %dx%d double %v raw %v", img.Format, img.Width, img.Height, img.DoublePage, img.Raw != nil)
		}
		if img.MediaType() != "image/jxl" || img.EPUBImgPath() != "OEBPS/Images/img_0_p0.jxl" {
			t.Errorf("got %s %s", img.MediaType(), img.EPUBImgPath())
		}
	})

	t.Run("unsupported extension", func(t *testing.T) {
		if _, err := e.copyRawDataToStorage(storage, func() ([]byte, error) { return data, nil }, 1, "", "a.gif"); err == nil {
			t.Error("expected an error, got nil")
		}
	})
}

func TestCopyRawDataToStorageUnreadableImage(t *testing.T) {
	var buf bytes.Buffer
	if err := jpeg.Encode(&buf, image.NewGray(image.Rect(0, 0, 10, 20)), nil); err != nil {
		t.Fatal(err)
	}
	truncated := buf.Bytes()[:buf.Len()/2]

	e := ePUBImagePassthrough{epuboptions.EPUBOptions{
		TrueTypeFont: fonts.Default,
		Image: epuboptions.Image{
			Quality: 85,
			View:    epuboptions.View{Color: epuboptions.Color{Foreground: "000000", Background: "FFFFFF"}},
		},
	}}

	for _, c := range []struct {
		name     string
		data     []byte
		readErr  error
		id       int
		wantCov  bool
		filename string
	}{
		{"bad header", []byte("bad"), nil, 1, false, "ch1/001.jpg"},
		{"truncated cover", truncated, nil, 0, true, "ch1/000.jpg"},
		{"read error", nil, errors.New("read failed"), 2, false, "ch1/002.png"},
	} {
		t.Run(c.name, func(t *testing.T) {
			storage, err := epubzip.NewStorageImageWriter(filepath.Join(t.TempDir(), "img.zip"), "jpeg")
			if err != nil {
				t.Fatal(err)
			}
			defer storage.Close()

			img, err := e.copyRawDataToStorage(
				storage,
				func() ([]byte, error) { return c.data, c.readErr },
				c.id, "", c.filename,
			)
			if err != nil {
				t.Fatalf("unreadable image must not fail the conversion: %v", err)
			}
			if img.Error == nil {
				t.Error("the cause must be kept to be reported")
			}
			if img.Format != "jpeg" || img.Width != 1200 || img.Height != 1920 {
				t.Errorf("got %s %dx%d", img.Format, img.Width, img.Height)
			}
			if img.Path != "ch1/" {
				t.Errorf("got path %q", img.Path)
			}
			if (img.Raw != nil) != c.wantCov {
				t.Errorf("raw data kept = %v, want %v", img.Raw != nil, c.wantCov)
			}
		})
	}
}
