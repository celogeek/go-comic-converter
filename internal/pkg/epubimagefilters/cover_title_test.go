package epubimagefilters

import (
	"image"
	"image/color"
	"image/draw"
	"testing"

	"github.com/celogeek/go-comic-converter/v3/internal/pkg/fonts"
	"github.com/stretchr/testify/assert"
)

func newCoverTitle(title, align string) CoverTitle {
	return CoverTitle{
		Title:       title,
		Align:       align,
		Font:        fonts.Default,
		FontSize:    40,
		BorderWidth: 4,
		Foreground:  color.Black,
		Background:  color.White,
	}
}

func drawCover(p CoverTitle, w, h int) *image.RGBA {
	src := image.NewRGBA(image.Rect(0, 0, w, h))
	draw.Draw(src, src.Bounds(), image.NewUniform(color.RGBA{R: 255, A: 255}), image.Point{}, draw.Src)
	dst := image.NewRGBA(src.Bounds())
	p.Draw(dst, src, nil)
	return dst
}

func TestCoverTitleTextBox(t *testing.T) {
	p := newCoverTitle("A rather long title that must be wrapped on several lines", "bottom")

	box := p.textBox(600, 800, p.FontSize)
	assert.Equal(t, 600, box.Bounds().Dx(), "box spans the max width")
	assert.Greater(t, box.Bounds().Dy(), 0)
	assert.LessOrEqual(t, box.Bounds().Dy(), 800)

	capped := p.textBox(600, 50, p.FontSize)
	assert.Equal(t, 50, capped.Bounds().Dy(), "height is capped at maxHeight")
}

func TestCoverTitleDraw(t *testing.T) {
	red := color.RGBA{R: 255, A: 255}

	t.Run("no title leaves image untouched", func(t *testing.T) {
		for _, title := range []string{"", " ", "\n", " \n\t "} {
			for _, align := range []string{"bottom", "center"} {
				dst := drawCover(newCoverTitle(title, align), 600, 800)
				assert.Equal(t, red, dst.RGBAAt(300, 400), "%q %s", title, align)
				assert.Equal(t, red, dst.RGBAAt(300, 799), "%q %s", title, align)
			}
		}
	})

	t.Run("bottom touches the bottom edge only", func(t *testing.T) {
		dst := drawCover(newCoverTitle("Title", "bottom"), 600, 800)
		assert.Equal(t, red, dst.RGBAAt(300, 0))
		assert.NotEqual(t, red, dst.RGBAAt(300, 798))
	})

	t.Run("center and unknown align draw the box in the middle", func(t *testing.T) {
		for _, align := range []string{"center", "whatever"} {
			dst := drawCover(newCoverTitle("Title", align), 600, 800)
			assert.Equal(t, red, dst.RGBAAt(300, 0), align)
			assert.Equal(t, red, dst.RGBAAt(300, 799), align)
			assert.NotEqual(t, red, dst.RGBAAt(300, 400), align)
			assert.NotEqual(t, red, dst.RGBAAt(2, 400), "%s: box spans the full width", align)
		}
	})
}
