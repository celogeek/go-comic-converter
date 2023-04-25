package epubimagedata

import (
	"archive/zip"
	"bytes"
	"compress/flate"
	"fmt"
	"hash/crc32"
	"image"
	"image/jpeg"
	"os"
	"time"
)

type ImageData struct {
	Header *zip.FileHeader
	Data   []byte
}

func (img *ImageData) CompressedSize() uint64 {
	return img.Header.CompressedSize64 + 30 + uint64(len(img.Header.Name))
}

func exitWithError(err error) {
	fmt.Fprintln(os.Stderr, err)
	os.Exit(1)
}

func New(id int, part int, img image.Image, quality int) *ImageData {
	name := fmt.Sprintf("OEBPS/Images/%d_p%d.jpg", id, part)
	return NewRaw(name, img, quality)
}

func NewRaw(name string, img image.Image, quality int) *ImageData {
	var (
		data, cdata bytes.Buffer
		err         error
	)

	err = jpeg.Encode(&data, img, &jpeg.Options{Quality: quality})
	if err != nil {
		exitWithError(err)
	}

	wcdata, err := flate.NewWriter(&cdata, flate.BestCompression)
	if err != nil {
		exitWithError(err)
	}

	_, err = wcdata.Write(data.Bytes())
	if err != nil {
		exitWithError(err)
	}

	err = wcdata.Close()
	if err != nil {
		exitWithError(err)
	}

	t := time.Now()
	return &ImageData{
		&zip.FileHeader{
			Name:               name,
			CompressedSize64:   uint64(cdata.Len()),
			UncompressedSize64: uint64(data.Len()),
			CRC32:              crc32.Checksum(data.Bytes(), crc32.IEEETable),
			Method:             zip.Deflate,
			ModifiedTime:       uint16(t.Second()/2 + t.Minute()<<5 + t.Hour()<<11),
			ModifiedDate:       uint16(t.Day() + int(t.Month())<<5 + (t.Year()-1980)<<9),
		},
		cdata.Bytes(),
	}
}
