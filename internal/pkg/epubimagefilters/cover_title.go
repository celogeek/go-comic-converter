package epubimagefilters

import (
	"image"
	"image/color"
	"image/draw"
	"math"
	"strings"

	"github.com/disintegration/gift"
	"github.com/fogleman/gg"
	"github.com/golang/freetype/truetype"
	"golang.org/x/image/font"
)

type CoverTitle struct {
	Title       string
	Align       string
	Font        *truetype.Font
	FontSize    float64
	BorderWidth float64
	Foreground  color.Color
	Background  color.Color
}

// Bounds size is the same as source
func (p CoverTitle) Bounds(srcBounds image.Rectangle) (dstBounds image.Rectangle) {
	return srcBounds
}

// Draw copies the src image and draws the title box at the bottom or in the center, depending on Align.
func (p CoverTitle) Draw(dst draw.Image, src image.Image, _ *gift.Options) {
	draw.Draw(dst, dst.Bounds(), src, src.Bounds().Min, draw.Src)

	// If no title, do nothing. A title made only of newlines has no line to draw
	// (textBox would get an infinite height), and only spaces would draw an empty box.
	if strings.TrimSpace(p.Title) == "" {
		return
	}

	box := p.textBox(float64(dst.Bounds().Dx()), float64(dst.Bounds().Dy()), p.FontSize)
	bY, bMin := box.Bounds().Dy(), box.Bounds().Min
	dX, dY := dst.Bounds().Dx(), dst.Bounds().Dy()

	// The box always spans the full width, only its vertical position depends on Align.
	switch p.Align {
	case "bottom":
		draw.Draw(dst, image.Rect(0, dY-bY, dX, dY), box, bMin, draw.Over)
	default: // center
		draw.Draw(dst, image.Rect(0, (dY-bY)/2, dX, (dY+bY)/2), box, bMin, draw.Over)
	}

}

// textBox renders the title wrapped and centered in a rounded box of maxWidth,
// with an even margin between the box edges and the visible glyphs. The height
// fits the text, capped at maxHeight (the text is clipped beyond that).
func (p CoverTitle) textBox(maxWidth, maxHeight, margin float64) image.Image {
	const lineSpacing = 1.2
	// Scale the corners with the text: 24px for a 64px font.
	cornerRadius := p.FontSize * 3 / 8

	face := truetype.NewFace(p.Font, &truetype.Options{Size: p.FontSize})
	textW := maxWidth - 2*margin

	measure := gg.NewContext(1, 1)
	measure.SetFontFace(face)
	lines := measure.WordWrap(p.Title, textW)
	lineH := measure.FontHeight()

	// Ink extent of the lines relative to the first baseline (y=0).
	// BoundString returns 26.6 fixed-point values: divide by 64 for pixels.
	inkTop, inkBottom := math.Inf(1), math.Inf(-1)
	for i, line := range lines {
		b, _ := font.BoundString(face, line)
		baseline := float64(i) * lineH * lineSpacing
		inkTop = math.Min(inkTop, baseline+float64(b.Min.Y)/64)
		inkBottom = math.Max(inkBottom, baseline+float64(b.Max.Y)/64)
	}

	h := math.Min(math.Ceil(inkBottom-inkTop+2*margin), maxHeight)
	w := math.Ceil(maxWidth)
	firstBaseline := (h-(inkBottom-inkTop))/2 - inkTop

	dc := gg.NewContext(int(w), int(h))
	dc.SetFontFace(face)
	// Inset by half the border so the stroke isn't clipped at the image edges.
	dc.DrawRoundedRectangle(p.BorderWidth/2, p.BorderWidth/2, w-p.BorderWidth, h-p.BorderWidth, cornerRadius-p.BorderWidth/2)
	dc.SetColor(p.Background)
	if p.BorderWidth > 0 {
		dc.FillPreserve()
		dc.SetColor(p.Foreground)
		dc.SetLineWidth(p.BorderWidth)
		dc.Stroke()
	} else {
		dc.Fill()
	}
	// With ay=0, DrawStringWrapped puts the first baseline at y+FontHeight.
	dc.SetColor(p.Foreground)
	dc.DrawStringWrapped(p.Title, w/2, firstBaseline-lineH, 0.5, 0, textW, lineSpacing, gg.AlignCenter)
	return dc.Image()
}
