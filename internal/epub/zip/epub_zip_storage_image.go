package epubzip

import (
	"archive/zip"
	"image"
	"os"
	"sync"
)

type EPUBZipStorageImageWriter struct {
	fh *os.File
	fz *zip.Writer

	mut *sync.Mutex
}

func NewEPUBZipStorageImageWriter(filename string) (*EPUBZipStorageImageWriter, error) {
	fh, err := os.Create(filename)
	if err != nil {
		return nil, err
	}
	fz := zip.NewWriter(fh)
	return &EPUBZipStorageImageWriter{fh, fz, &sync.Mutex{}}, nil
}

func (e *EPUBZipStorageImageWriter) Close() error {
	if err := e.fz.Close(); err != nil {
		e.fh.Close()
		return err
	}
	return e.fh.Close()
}

func (e *EPUBZipStorageImageWriter) Add(filename string, img image.Image, quality int) error {
	zipImage, err := CompressImage(filename, img, quality)
	if err != nil {
		return err
	}

	e.mut.Lock()
	defer e.mut.Unlock()
	fh, err := e.fz.CreateRaw(zipImage.Header)
	if err != nil {
		return err
	}
	_, err = fh.Write(zipImage.Data)
	if err != nil {
		return err
	}

	return nil
}

type EPUBZipStorageImageReader struct {
	filename string
	fh       *os.File
	fz       *zip.Reader

	files map[string]*zip.File
}

func NewEPUBZipStorageImageReader(filename string) (*EPUBZipStorageImageReader, error) {
	fh, err := os.Open(filename)
	if err != nil {
		return nil, err
	}
	s, err := fh.Stat()
	if err != nil {
		return nil, err
	}
	fz, err := zip.NewReader(fh, s.Size())
	if err != nil {
		return nil, err
	}
	files := map[string]*zip.File{}
	for _, z := range fz.File {
		files[z.Name] = z
	}
	return &EPUBZipStorageImageReader{filename, fh, fz, files}, nil
}

func (e *EPUBZipStorageImageReader) Get(filename string) *zip.File {
	return e.files[filename]
}

func (e *EPUBZipStorageImageReader) Size(filename string) uint64 {
	img := e.Get(filename)
	if img != nil {
		return img.CompressedSize64 + 30 + uint64(len(img.Name))
	}
	return 0
}

func (e *EPUBZipStorageImageReader) Close() error {
	return e.fh.Close()
}

func (e *EPUBZipStorageImageReader) Remove() error {
	return os.Remove(e.filename)
}
