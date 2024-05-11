package epubzip

import (
	"archive/zip"
	"bytes"
	"compress/flate"
	"fmt"
	"hash/crc32"
	"image"
	"image/jpeg"
	"image/png"
	"time"
)

type ZipImage struct {
	Header *zip.FileHeader
	Data   []byte
}

// CompressImage create gzip encoded jpeg
func CompressImage(filename string, format string, img image.Image, quality int) (ZipImage, error) {
	var (
		data, cdata bytes.Buffer
		err         error
	)

	switch format {
	case "png":
		err = png.Encode(&data, img)
	case "jpeg":
		err = jpeg.Encode(&data, img, &jpeg.Options{Quality: quality})
	default:
		err = fmt.Errorf("unknown format %q", format)
	}
	if err != nil {
		return ZipImage{}, err
	}

	wcdata, err := flate.NewWriter(&cdata, flate.BestCompression)
	if err != nil {
		return ZipImage{}, err
	}

	_, err = wcdata.Write(data.Bytes())
	if err != nil {
		return ZipImage{}, err
	}

	err = wcdata.Close()
	if err != nil {
		return ZipImage{}, err
	}

	t := time.Now()
	//goland:noinspection GoDeprecation
	return ZipImage{
		&zip.FileHeader{
			Name:               filename,
			CompressedSize64:   uint64(cdata.Len()),
			UncompressedSize64: uint64(data.Len()),
			CRC32:              crc32.Checksum(data.Bytes(), crc32.IEEETable),
			Method:             zip.Deflate,
			ModifiedTime:       uint16(t.Second()/2 + t.Minute()<<5 + t.Hour()<<11),
			ModifiedDate:       uint16(t.Day() + int(t.Month())<<5 + (t.Year()-1980)<<9),
		},
		cdata.Bytes(),
	}, nil
}
