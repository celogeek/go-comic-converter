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

type Image struct {
	Header *zip.FileHeader
	Data   []byte
}

// CompressImage create gzip encoded jpeg
func CompressImage(filename string, format string, img image.Image, quality int) (Image, error) {
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
		return Image{}, err
	}

	wcdata, err := flate.NewWriter(&cdata, flate.BestCompression)
	if err != nil {
		return Image{}, err
	}

	_, err = wcdata.Write(data.Bytes())
	if err != nil {
		return Image{}, err
	}

	err = wcdata.Close()
	if err != nil {
		return Image{}, err
	}

	t := time.Now()
	//goland:noinspection GoDeprecation
	return Image{
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

func CompressRaw(filename string, uncompressedData []byte) (Image, error) {
	var (
		cdata bytes.Buffer
		err   error
	)
	wcdata, err := flate.NewWriter(&cdata, flate.BestCompression)
	if err != nil {
		return Image{}, err
	}

	_, err = wcdata.Write(uncompressedData)
	if err != nil {
		return Image{}, err
	}

	err = wcdata.Close()
	if err != nil {
		return Image{}, err
	}

	t := time.Now()
	//goland:noinspection GoDeprecation
	return Image{
		&zip.FileHeader{
			Name:               filename,
			CompressedSize64:   uint64(cdata.Len()),
			UncompressedSize64: uint64(len(uncompressedData)),
			CRC32:              crc32.Checksum(uncompressedData, crc32.IEEETable),
			Method:             zip.Deflate,
			ModifiedTime:       uint16(t.Second()/2 + t.Minute()<<5 + t.Hour()<<11),
			ModifiedDate:       uint16(t.Day() + int(t.Month())<<5 + (t.Year()-1980)<<9),
		},
		cdata.Bytes(),
	}, nil
}
