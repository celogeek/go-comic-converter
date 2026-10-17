package epubzip

import (
	"bytes"
	"compress/flate"
	"image"
	"io"
	"testing"

	"github.com/gen2brain/jxl"
	"github.com/stretchr/testify/assert"
)

func TestCompressImageJXL(t *testing.T) {
	img := image.NewGray(image.Rect(0, 0, 32, 32))
	for i := range img.Pix {
		img.Pix[i] = uint8(i)
	}

	for _, q := range []int{80, 100} {
		z, err := CompressImage("a.jxl", "jxl", img, q, 4)
		assert.NoError(t, err)
		assert.NotEmpty(t, z.Data)

		raw, err := flateDecode(z.Data)
		assert.NoError(t, err)
		dec, err := jxl.Decode(bytes.NewReader(raw))
		assert.NoError(t, err)
		assert.Equal(t, img.Bounds(), dec.Bounds())
	}
}

func flateDecode(data []byte) ([]byte, error) {
	return io.ReadAll(flate.NewReader(bytes.NewReader(data)))
}
