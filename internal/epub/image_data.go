package epub

import (
	"archive/zip"
	"bytes"
	"compress/flate"
	"hash/crc32"
	"time"
)

type imageData struct {
	Header *zip.FileHeader
	Data   []byte
}

func (img *imageData) CompressedSize() uint64 {
	return img.Header.CompressedSize64 + 30 + uint64(len(img.Header.Name))
}

func newImageData(name string, data []byte) *imageData {
	cdata := bytes.NewBuffer([]byte{})
	wcdata, err := flate.NewWriter(cdata, flate.BestCompression)
	if err != nil {
		panic(err)
	}
	wcdata.Write(data)
	wcdata.Close()
	if err != nil {
		panic(err)
	}
	t := time.Now()
	return &imageData{
		&zip.FileHeader{
			Name:               name,
			CompressedSize64:   uint64(cdata.Len()),
			UncompressedSize64: uint64(len(data)),
			CRC32:              crc32.Checksum(data, crc32.IEEETable),
			Method:             zip.Deflate,
			ModifiedTime:       uint16(t.Second()/2 + t.Minute()<<5 + t.Hour()<<11),
			ModifiedDate:       uint16(t.Day() + int(t.Month())<<5 + (t.Year()-1980)<<9),
		},
		cdata.Bytes(),
	}
}
