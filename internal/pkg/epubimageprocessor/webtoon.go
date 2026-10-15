package epubimageprocessor

import (
	"fmt"
	"image"
	"image/color"
	"image/draw"

	"github.com/disintegration/gift"
)

// Thresholds used to cut a webtoon into pages.
const (
	// the cut is looked for in the last part of the page (in percent of its height)
	webtoonCutWindow = 25
	// without a gap there, the page can be shorter than the view (down to this ratio, in percent of its height),
	// the end of the page is left blank
	webtoonMinFill = 60
	// or longer than the view (up to this ratio): the page is then shrunk to fit the view,
	// better than a panel cut in 2 pages
	webtoonMaxStretch = 150
	// rows that differ by less than this (from 0 to 255) from end to end are a gap between 2 panels:
	// the compression of the jpeg adds some noise to a flat background
	webtoonGapTolerance = 12
	// min rows of a gap, shorter ones are inside a drawing (a flat sky, a speech bubble)
	webtoonGapMinRows = 6
	// without any gap, the cut goes through the drawing: the pages overlap (in percent of the page height)
	// so no line of the text is lost
	webtoonOverlap = 4
	// a page doesn't start with more blank than this (in percent of the page height):
	// the webtoons use a lot of blank to pace the story, that is only a waste of pages on a reader
	webtoonMargin = 3
)

// webtoonStrip a piece of the webtoon, with its position in the chapter.
type webtoonStrip struct {
	img image.Image
	y0  int
}

// webtoonPager cuts the strips of a chapter into pages of the size of the view.
//
// The strips are seen as 1 long image, only the strips that are still needed are kept.
type webtoonPager struct {
	width, pageH int
	strips       []webtoonStrip
	// position of the next page, and the end of the strips
	top, end int
}

func (p *webtoonPager) add(img image.Image) {
	b := img.Bounds()
	if b.Dx() != p.width {
		g := gift.New(gift.Resize(p.width, 0, gift.LanczosResampling))
		dst := image.NewRGBA(g.Bounds(b))
		g.Draw(dst, img)
		img = dst
		b = img.Bounds()
	}
	p.strips = append(p.strips, webtoonStrip{img: img, y0: p.end})
	p.end += b.Dy()
}

// rowSpread difference between the lightest and the darkest pixel of a row of the chapter.
func (p *webtoonPager) rowSpread(y int) int {
	for i := len(p.strips) - 1; i >= 0; i-- {
		s := p.strips[i]
		if y < s.y0 {
			continue
		}
		b := s.img.Bounds()
		return rowSpread(s.img, b.Min.X, b.Max.X, b.Min.Y+y-s.y0)
	}
	return 255
}

// webtoonGap a gap between 2 panels: its middle, and its number of rows.
type webtoonGap struct {
	mid, rows int
}

// gaps the gaps in the rows [from, to) of the chapter, from the top to the bottom.
func (p *webtoonPager) gaps(from, to int) (gaps []webtoonGap) {
	to = min(to, p.end)
	run := 0
	for r := from; r <= to; r++ {
		if r < to && p.rowSpread(r) <= webtoonGapTolerance {
			run++
			continue
		}
		// a run of gap ends at r (excluded)
		if run >= webtoonGapMinRows {
			gaps = append(gaps, webtoonGap{mid: r - run/2, rows: run})
		}
		run = 0
	}
	return
}

// cut position of the end of the page that starts at p.top.
//
// A panel is never cut when there is a gap to cut in, in this order:
//   - the biggest gap of the window at the end of the page,
//   - the last gap before the window, the page is shorter than the view,
//   - the first gap after the end of the page, the page is longer than the view and will be shrunk,
//   - otherwise the cut goes through the drawing, at the end of the page (hard).
func (p *webtoonPager) cut() (y int, hard bool) {
	limit := p.top + p.pageH
	from := limit - p.pageH*webtoonCutWindow/100

	if gaps := p.gaps(from, limit); len(gaps) > 0 {
		best := gaps[0]
		for _, g := range gaps[1:] {
			if g.rows >= best.rows {
				best = g
			}
		}
		return best.mid, false
	}

	if gaps := p.gaps(p.top+p.pageH*webtoonMinFill/100, from); len(gaps) > 0 {
		return gaps[len(gaps)-1].mid, false
	}

	if gaps := p.gaps(limit, p.top+p.pageH*webtoonMaxStretch/100); len(gaps) > 0 {
		return gaps[0].mid, false
	}

	return limit, true
}

// page copy the rows [y0, y1) of the chapter.
func (p *webtoonPager) page(y0, y1 int) image.Image {
	dst := image.NewRGBA(image.Rect(0, 0, p.width, y1-y0))
	for _, s := range p.strips {
		b := s.img.Bounds()
		s0, s1 := max(y0, s.y0), min(y1, s.y0+b.Dy())
		if s0 >= s1 {
			continue
		}
		draw.Draw(dst,
			image.Rect(0, s0-y0, p.width, s1-y0),
			s.img,
			image.Pt(b.Min.X, b.Min.Y+s0-s.y0),
			draw.Src,
		)
	}
	return dst
}

// drop forgets the strips above the next page.
func (p *webtoonPager) drop() {
	n := 0
	for n < len(p.strips) {
		s := p.strips[n]
		if s.y0+s.img.Bounds().Dy() > p.top {
			break
		}
		n++
	}
	p.strips = p.strips[n:]
}

// trim skips the blank rows at the top of the next page, but a small margin.
// It tells if everything up to the end is blank.
func (p *webtoonPager) trim() (blank bool) {
	y := p.top
	for y < p.end && p.rowSpread(y) <= webtoonGapTolerance {
		y++
	}
	p.top = max(p.top, y-p.pageH*webtoonMargin/100)
	p.drop()
	return y >= p.end
}

// next the next page of the chapter, when there is enough to fill it, or when it is the end (last).
func (p *webtoonPager) next(last bool) (image.Image, bool) {
	// a blank end is not a page
	if blank := p.trim(); blank && last {
		p.top = p.end
		p.drop()
		return nil, false
	}

	rest := p.end - p.top
	switch {
	case rest <= 0:
		return nil, false
	// the gap after the end of the page needs the rows that follow, until the end of the chapter
	case rest >= p.pageH*webtoonMaxStretch/100 || (last && rest >= p.pageH):
		y, hard := p.cut()
		page := p.page(p.top, y)
		p.top = y
		// nothing to overlap at the end of the chapter
		if hard && y < p.end {
			p.top -= p.pageH * webtoonOverlap / 100
		}
		p.drop()
		return page, true
	case !last:
		return nil, false
	}

	// the end of the chapter
	page := p.page(p.top, p.end)
	p.top = p.end
	p.drop()
	return page, true
}

// webtoonPages cuts the webtoon into pages of the size of the view.
//
// The images of a directory are the pieces of a long vertical strip: they are put back in order and
// joined, then cut into pages with the aspect ratio of the view, in the gaps between the panels when possible.
// Each directory is a chapter, the pages keep the path of the directory, and are named after the first image
// they come from. progress is called for each source image read.
func (e ePUBImageProcessor) webtoonPages(in <-chan task, progress func()) <-chan task {
	out := make(chan task, e.Workers)

	go func() {
		defer close(out)

		var (
			id    int
			path  string
			pager *webtoonPager
			// name of the first image of the chapter
			name string
			// pages of the chapter already sent
			count int
		)

		send := func(img image.Image) {
			count++
			out <- task{
				Id:    id,
				Image: img,
				Path:  path,
				Name:  fmt.Sprintf("%s-%04d", name, count),
			}
			id++
		}

		flush := func() {
			if pager == nil {
				return
			}
			for {
				img, ok := pager.next(true)
				if !ok {
					break
				}
				send(img)
			}
			pager = nil
		}

		forEachInOrder(in, out, func(t task) {
			defer progress()

			// can't be part of the strip: a placeholder is a page of its own
			if t.Image == nil || t.Error != nil {
				flush()
				count = 0
				path, name = t.Path, t.Name
				send(t.Image)
				return
			}

			if pager == nil || t.Path != path {
				flush()
				b := t.Image.Bounds()
				w, h := e.Image.View.Width, e.Image.View.Height
				if w <= 0 || h <= 0 {
					w, h = 3, 4
				}
				pager = &webtoonPager{
					width: b.Dx(),
					pageH: max(1, b.Dx()*h/w),
				}
				path, name, count = t.Path, t.Name, 0
			}

			pager.add(t.Image)
			for {
				img, ok := pager.next(false)
				if !ok {
					break
				}
				send(img)
			}
		})
		flush()
	}()

	return out
}

// rowSpread difference between the lightest and the darkest pixel of the row y, from x0 to x1 (excluded).
func rowSpread(img image.Image, x0, x1, y int) int {
	lo, hi := 255, 0
	update := func(v int) {
		lo, hi = min(lo, v), max(hi, v)
	}

	switch m := img.(type) {
	case *image.YCbCr:
		off := m.YOffset(x0, y)
		for _, v := range m.Y[off : off+x1-x0] {
			update(int(v))
		}
	case *image.Gray:
		off := m.PixOffset(x0, y)
		for _, v := range m.Pix[off : off+x1-x0] {
			update(int(v))
		}
	default:
		for x := x0; x < x1; x++ {
			update(int(color.GrayModel.Convert(img.At(x, y)).(color.Gray).Y))
		}
	}
	return hi - lo
}
