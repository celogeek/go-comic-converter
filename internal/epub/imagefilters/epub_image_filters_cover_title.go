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

// CoverTitle Create a title with the cover image
func CoverTitle(title string, align string, pctWidth int, pctMargin int, maxFontSize int, borderSize int) gift.Filter {
	return &coverTitle{title, align, pctWidth, pctMargin, maxFontSize, borderSize}
}

type coverTitle struct {
	title       string
	align       string
	pctWidth    int
	pctMargin   int
	maxFontSize int
	borderSize  int
}

// Bounds size is the same as source
func (p *coverTitle) Bounds(srcBounds image.Rectangle) (dstBounds image.Rectangle) {
	return srcBounds
}

// Draw blur the src image, and create a box with the title in the middle
func (p *coverTitle) Draw(dst draw.Image, src image.Image, _ *gift.Options) {
	draw.Draw(dst, dst.Bounds(), src, src.Bounds().Min, draw.Src)
	if p.title == "" {
		return
	}

	srcWidth, srcHeight := src.Bounds().Dx(), src.Bounds().Dy()

	// Calculate size of title
	f, _ := truetype.Parse(gomonobold.TTF)
	var fontSize, textWidth, textHeight int
	for fontSize = p.maxFontSize; fontSize >= 12; fontSize -= 1 {
		face := truetype.NewFace(f, &truetype.Options{Size: float64(fontSize), DPI: 72})
		textWidth = font.MeasureString(face, p.title).Ceil()
		textHeight = face.Metrics().Ascent.Ceil() + face.Metrics().Descent.Ceil()
		if textWidth+2*p.borderSize < srcWidth*p.pctWidth/100 && 3*textHeight+2*p.borderSize < srcHeight {
			break
		}
	}

	// Draw rectangle in the middle of the image
	marginSize := fontSize * p.pctMargin / 100
	var textPosStart, textPosEnd int
	if p.align == "bottom" {
		textPosStart = srcHeight - textHeight - p.borderSize - marginSize
		textPosEnd = srcHeight - p.borderSize - marginSize
	} else {
		textPosStart = srcHeight/2 - textHeight/2
		textPosEnd = srcHeight/2 + textHeight/2
	}
	borderArea := image.Rect((srcWidth-(srcWidth*p.pctWidth/100))/2, textPosStart-p.borderSize-marginSize, (srcWidth+(srcWidth*p.pctWidth/100))/2, textPosEnd+p.borderSize+marginSize)
	textArea := image.Rect(borderArea.Bounds().Min.X+p.borderSize, textPosStart-marginSize, borderArea.Bounds().Max.X-p.borderSize, textPosEnd+marginSize)

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

	textLeft := textArea.Min.X + textArea.Dx()/2 - textWidth/2
	if textLeft < textArea.Min.X {
		textLeft = textArea.Min.X
	}
	textTop := textArea.Min.Y + textArea.Dy()/2 + textHeight/4
	_, _ = c.DrawString(p.title, freetype.Pt(textLeft, textTop))
}
