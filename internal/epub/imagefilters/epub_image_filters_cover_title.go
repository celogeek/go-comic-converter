package epubimagefilters

import (
	"image"
	"image/draw"

	"github.com/disintegration/gift"
	"github.com/golang/freetype"
	"github.com/golang/freetype/truetype"
	"golang.org/x/image/font"
	"golang.org/x/image/font/gofont/gomonobold"
)

// Create a title with the cover image
func CoverTitle(title string) gift.Filter {
	return &coverTitle{title}
}

type coverTitle struct {
	title string
}

// size is the same as source
func (p *coverTitle) Bounds(srcBounds image.Rectangle) (dstBounds image.Rectangle) {
	return srcBounds
}

// blur the src image, and create a box with the title in the middle
func (p *coverTitle) Draw(dst draw.Image, src image.Image, options *gift.Options) {
	draw.Draw(dst, dst.Bounds(), src, src.Bounds().Min, draw.Src)

	srcWidth, srcHeight := src.Bounds().Dx(), src.Bounds().Dy()

	// Calculate size of title
	f, _ := truetype.Parse(gomonobold.TTF)
	borderSize := 4
	var fontSize, textWidth, textHeight int
	for fontSize = 64; fontSize >= 12; fontSize -= 1 {
		face := truetype.NewFace(f, &truetype.Options{Size: float64(fontSize), DPI: 72})
		textWidth = font.MeasureString(face, p.title).Ceil()
		textHeight = face.Metrics().Ascent.Ceil() + face.Metrics().Descent.Ceil()
		if textWidth+2*borderSize < srcWidth && 3*textHeight+2*borderSize < srcHeight {
			break
		}
	}

	// Draw rectangle in the middle of the image
	textPosStart := srcHeight/2 - textHeight/2
	textPosEnd := srcHeight/2 + textHeight/2
	marginSize := fontSize
	borderArea := image.Rect(0, textPosStart-borderSize-marginSize, srcWidth, textPosEnd+borderSize+marginSize)
	textArea := image.Rect(borderSize, textPosStart-marginSize, srcWidth-borderSize, textPosEnd+marginSize)

	draw.Draw(
		dst,
		borderArea,
		image.Black,
		borderArea.Min,
		draw.Src,
	)

	draw.Draw(
		dst,
		textArea,
		image.White,
		textArea.Min,
		draw.Src,
	)

	// Draw text
	c := freetype.NewContext()
	c.SetDPI(72)
	c.SetFontSize(float64(fontSize))
	c.SetFont(f)
	c.SetClip(textArea)
	c.SetDst(dst)
	c.SetSrc(image.Black)

	textLeft := srcWidth/2 - textWidth/2
	if textLeft < borderSize {
		textLeft = borderSize
	}
	c.DrawString(p.title, freetype.Pt(textLeft, srcHeight/2+textHeight/4))
}
