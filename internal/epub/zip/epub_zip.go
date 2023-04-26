/*
Helper to write epub files.

We create a zip with the magic epub mimetype.
*/
package epubzip

import (
	"archive/zip"
	"os"
	"time"

	epubimagedata "github.com/celogeek/go-comic-converter/v2/internal/epub/imagedata"
)

type EpubZip struct {
	w  *os.File
	wz *zip.Writer
}

// create a new epub
func New(path string) (*EpubZip, error) {
	w, err := os.Create(path)
	if err != nil {
		return nil, err
	}
	wz := zip.NewWriter(w)
	return &EpubZip{w, wz}, nil
}

// close compress pipe and file.
func (e *EpubZip) Close() error {
	if err := e.wz.Close(); err != nil {
		return err
	}
	return e.w.Close()
}

// Write mimetype, in a very specific way.
// This will be valid with epubcheck tools.
func (e *EpubZip) WriteMagic() error {
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

// Write image. They are already compressed, so we write them down directly.
func (e *EpubZip) WriteImage(image *epubimagedata.ImageData) error {
	m, err := e.wz.CreateRaw(image.Header)
	if err != nil {
		return err
	}
	_, err = m.Write(image.Data)
	return err
}

// Write file. Compressed it using deflate.
func (e *EpubZip) WriteFile(file string, content []byte) error {
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
