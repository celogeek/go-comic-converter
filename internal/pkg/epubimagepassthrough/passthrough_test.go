package epubimagepassthrough

import (
	"bytes"
	"image"
	"image/jpeg"
	"path/filepath"
	"testing"

	"github.com/celogeek/go-comic-converter/v3/internal/pkg/epubzip"
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

	t.Run("unsupported extension", func(t *testing.T) {
		if _, err := e.copyRawDataToStorage(storage, func() ([]byte, error) { return data, nil }, 1, "", "a.gif"); err == nil {
			t.Error("expected an error, got nil")
		}
	})
}
