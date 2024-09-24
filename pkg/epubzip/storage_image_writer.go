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

func NewStorageImageWriter(filename string, format string) (StorageImageWriter, error) {
	fh, err := os.Create(filename)
	if err != nil {
		return StorageImageWriter{}, err
	}
	fz := zip.NewWriter(fh)
	return StorageImageWriter{fh, fz, format, &sync.Mutex{}}, nil
}

func (e StorageImageWriter) Close() error {
	if err := e.fz.Close(); err != nil {
		_ = e.fh.Close()
		return err
	}
	return e.fh.Close()
}

func (e StorageImageWriter) Add(filename string, img image.Image, quality int) error {
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
