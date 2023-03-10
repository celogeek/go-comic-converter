package epub

import (
	"archive/zip"
	"fmt"
	"os"
	"time"
)

type epubZip struct {
	w  *os.File
	wz *zip.Writer
}

func newEpubZip(path string) (*epubZip, error) {
	w, err := os.Create(path)
	if err != nil {
		return nil, err
	}
	wz := zip.NewWriter(w)
	return &epubZip{w, wz}, nil
}

func (e *epubZip) Close() error {
	if err := e.wz.Close(); err != nil {
		return err
	}
	return e.w.Close()
}

func (e *epubZip) WriteMagic() error {
	t := time.Now()
	fh := &zip.FileHeader{
		Name:               "mimetype",
		Method:             zip.Store,
		Modified:           t,
		ModifiedTime:       uint16(t.Second()/2 + t.Minute()<<5 + t.Hour()<<11),
		ModifiedDate:       uint16(t.Day() + int(t.Month())<<5 + (t.Year()-1980)<<9),
		CompressedSize64:   20,
		UncompressedSize64: 20,
		CRC32:              0x2cab616f,
	}
	fh.SetMode(0600)
	m, err := e.wz.CreateRaw(fh)

	if err != nil {
		return err
	}
	_, err = m.Write([]byte("application/epub+zip"))
	return err
}

func (e *epubZip) WriteImage(image *ImageData) error {
	m, err := e.wz.CreateRaw(image.Header)
	if err != nil {
		return err
	}
	_, err = m.Write(image.Data)
	return err
}

func (e *epubZip) WriteFile(file string, data any) error {
	var content []byte
	switch b := data.(type) {
	case string:
		content = []byte(b)
	case []byte:
		content = b
	default:
		return fmt.Errorf("support string of []byte")
	}

	m, err := e.wz.CreateHeader(&zip.FileHeader{
		Name:     file,
		Modified: time.Now(),
		Method:   zip.Deflate,
	})
	if err != nil {
		return err
	}
	_, err = m.Write(content)
	return err
}
