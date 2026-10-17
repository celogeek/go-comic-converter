package epubzip

import (
	"image"
	"os/exec"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestCompressImageJXL(t *testing.T) {
	if _, err := exec.LookPath("cjxl"); err != nil {
		t.Skip("cjxl not installed")
	}
	img := image.NewGray(image.Rect(0, 0, 32, 32))
	for i := range img.Pix {
		img.Pix[i] = uint8(i)
	}

	for _, q := range []int{80, 100} {
		z, err := CompressImage("a.jxl", "jxl", img, q)
		assert.NoError(t, err)
		assert.NotEmpty(t, z.Data)
	}
}
