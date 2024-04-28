package epubzip

import (
	"archive/zip"
	"image"
	"os"
	"sync"
)

type StorageImageWriter struct {
	fh     *os.File
	fz     *zip.Writer
	format string
	mut    *sync.Mutex
}

func NewStorageImageWriter(filename string, format string) (*StorageImageWriter, error) {
	fh, err := os.Create(filename)
	if err != nil {
		return nil, err
	}
	fz := zip.NewWriter(fh)
	return &StorageImageWriter{fh, fz, format, &sync.Mutex{}}, nil
}

func (e *StorageImageWriter) Close() error {
	if err := e.fz.Close(); err != nil {
		_ = e.fh.Close()
		return err
	}
	return e.fh.Close()
}

func (e *StorageImageWriter) Add(filename string, img image.Image, quality int) error {
	zipImage, err := CompressImage(filename, e.format, img, quality)
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

type StorageImageReader struct {
	filename string
	fh       *os.File
	fz       *zip.Reader

	files map[string]*zip.File
}

func NewStorageImageReader(filename string) (*StorageImageReader, error) {
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
	return &StorageImageReader{filename, fh, fz, files}, nil
}

func (e *StorageImageReader) Get(filename string) *zip.File {
	return e.files[filename]
}

func (e *StorageImageReader) Size(filename string) uint64 {
	img := e.Get(filename)
	if img != nil {
		return img.CompressedSize64 + 30 + uint64(len(img.Name))
	}
	return 0
}

func (e *StorageImageReader) Close() error {
	return e.fh.Close()
}

func (e *StorageImageReader) Remove() error {
	return os.Remove(e.filename)
}
