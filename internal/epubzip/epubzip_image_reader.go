package epubzip

import (
	"archive/zip"
	"os"
)

type EPUBZipImageReader struct {
	filename string
	fh       *os.File
	fz       *zip.Reader

	files map[string]*zip.File
}

func NewImageReader(filename string) (*EPUBZipImageReader, error) {
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
	return &EPUBZipImageReader{filename, fh, fz, files}, nil
}

func (e *EPUBZipImageReader) Get(filename string) *zip.File {
	return e.files[filename]
}

func (e *EPUBZipImageReader) Size(filename string) uint64 {
	img := e.Get(filename)
	if img != nil {
		return img.CompressedSize64 + 30 + uint64(len(img.Name))
	}
	return 0
}

func (e *EPUBZipImageReader) Close() error {
	return e.fh.Close()
}

func (e *EPUBZipImageReader) Remove() error {
	return os.Remove(e.filename)
}
