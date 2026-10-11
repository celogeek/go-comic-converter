package epubimageprocessor

import (
	"archive/zip"
	"os"
	"path/filepath"
	"testing"

	"github.com/celogeek/go-comic-converter/v3/internal/pkg/fonts"
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
