package epubzip

import (
	"archive/zip"
	"os"
)

type StorageImageReader struct {
	filename string
	fh       *os.File
	fz       *zip.Reader

	files map[string]*zip.File
}

func NewStorageImageReader(filename string) (StorageImageReader, error) {
	fh, err := os.Open(filename)
	if err != nil {
		return StorageImageReader{}, err
	}
	s, err := fh.Stat()
	if err != nil {
		return StorageImageReader{}, err
	}
	fz, err := zip.NewReader(fh, s.Size())
	if err != nil {
		return StorageImageReader{}, err
	}
	files := map[string]*zip.File{}
	for _, z := range fz.File {
		files[z.Name] = z
	}
	return StorageImageReader{filename, fh, fz, files}, nil
}

func (e StorageImageReader) Get(filename string) *zip.File {
	return e.files[filename]
}

func (e StorageImageReader) Size(filename string) uint64 {
	if img, ok := e.files[filename]; ok {
		return img.CompressedSize64 + 30 + uint64(len(img.Name))
	}
	return 0
}

func (e StorageImageReader) Close() error {
	return e.fh.Close()
}

func (e StorageImageReader) Remove() error {
	return os.Remove(e.filename)
}
