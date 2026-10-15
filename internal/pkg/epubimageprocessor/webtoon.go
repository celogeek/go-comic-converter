package epubimageprocessor

import (
	"fmt"
	"image"
	"image/color"
	"image/draw"
	"slices"

	"github.com/disintegration/gift"

	"github.com/celogeek/go-comic-converter/v3/internal/pkg/utils"
)

// Thresholds used to cut a webtoon into pages.
const (
	// the cut is looked for in the last part of the page (in percent of its height)
	webtoonCutWindow = 25
	// without a gap there, the page can be shorter than the view (down to this ratio, in percent of its height),
	// the end of the page is left blank
	webtoonMinFill = 35
	// or longer than the view (up to this ratio): the page is then shrunk to fit the view,
	// better than a panel cut in 2 pages
	webtoonMaxStretch = 300
	// rows that differ by less than this (from 0 to 255) from end to end are a gap between 2 panels:
	// the compression of the jpeg adds some noise to a flat background
	webtoonGapTolerance = 20
	// min rows of a gap, shorter ones are inside a drawing (a flat sky, a speech bubble)
	webtoonGapMinRows = 6
	// without any gap, the cut goes through the drawing: the pages overlap (in percent of the page height)
	// so no line of the text is lost
	webtoonOverlap = 4
	// a page doesn't start with more blank than this (in percent of the page height):
	// the webtoons use a lot of blank to pace the story, that is only a waste of pages on a reader
	webtoonMargin = 3
	// the edge of a page has a main color when this ratio of its pixels has it (in percent)
	webtoonEdgeMainColor = 75
	// gaps that differ by more than this (from 0 to 255) are 2 different backgrounds: a black scene and a white one
	webtoonBackgroundChange = 64
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

// strip the strip of the row y of the chapter, and the row in it.
func (p *webtoonPager) strip(y int) (image.Image, int, bool) {
	for i := len(p.strips) - 1; i >= 0; i-- {
		s := p.strips[i]
		if y >= s.y0 {
			return s.img, s.img.Bounds().Min.Y + y - s.y0, true
		}
	}
	return nil, 0, false
}

// rowSpread difference between the lightest and the darkest pixel of a row of the chapter.
func (p *webtoonPager) rowSpread(y int) int {
	img, sy, ok := p.strip(y)
	if !ok {
		return 255
	}
	b := img.Bounds()
	return rowSpread(img, b.Min.X, b.Max.X, sy)
}

// rowGray gray of the middle of a row of the chapter: the color of a gap.
func (p *webtoonPager) rowGray(y int) int {
	img, sy, ok := p.strip(y)
	if !ok {
		return 0
	}
	b := img.Bounds()
	return int(color.GrayModel.Convert(img.At(b.Min.X+b.Dx()/2, sy)).(color.Gray).Y)
}

// webtoonGap a gap between 2 panels: where to cut it, and its number of rows.
//
// A page ends where the gap starts, the gap is left to the next page where it is trimmed:
// a black background is a gap too, it would be shown at the end of the previous page.
type webtoonGap struct {
	at, rows int
	// gray is the color of the gap where it starts
	gray int
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
			start := r - run
			// it started before the rows looked at: the page ends where the gap really starts
			if start == from {
				for start > p.top && p.rowSpread(start-1) <= webtoonGapTolerance {
					start--
					run++
				}
			}
			gaps = append(gaps, webtoonGap{at: start, rows: run, gray: p.rowGray(start)})
		}
		run = 0
	}
	return
}

// cut position of the end of the page that starts at p.top.
//
// A panel is never cut when there is a gap to cut in, in this order:
//   - the first change of background on the page (see background), the page is shorter than the view,
//   - the biggest gap of the window at the end of the page,
//   - the last gap before the window, the page is shorter than the view,
//   - the first gap after the end of the page, the page is longer than the view and will be shrunk,
//   - otherwise the cut goes through the drawing, at the end of the page (hard).
//
// The last one needs the rows that follow the end of the page: ok is false when they are not all known yet,
// and the cut must wait for more rows (it never waits at the end of the chapter, last).
func (p *webtoonPager) cut(last bool) (y int, hard, ok bool) {
	limit := p.top + p.pageH
	from := limit - p.pageH*webtoonCutWindow/100
	minEnd := p.top + p.pageH*webtoonMinFill/100

	if bg, ok := p.background(limit); ok {
		for _, g := range p.gaps(minEnd, limit) {
			if abs(g.gray-bg) > webtoonBackgroundChange {
				return g.at, false, true
			}
		}
	}

	if gaps := p.gaps(from, limit); len(gaps) > 0 {
		best := gaps[0]
		for _, g := range gaps[1:] {
			if g.rows >= best.rows {
				best = g
			}
		}
		return best.at, false, true
	}

	if gaps := p.gaps(minEnd, from); len(gaps) > 0 {
		return gaps[len(gaps)-1].at, false, true
	}

	maxEnd := p.top + p.pageH*webtoonMaxStretch/100
	if gaps := p.gaps(limit, maxEnd); len(gaps) > 0 {
		return gaps[0].at, false, true
	}
	if !last && p.end < maxEnd {
		return 0, false, false
	}

	return limit, true, true
}

// background color of the page that starts at p.top: the one of its top when it is flat, or of its first gap before limit.
//
// A page doesn't mix 2 backgrounds: a white panel after a black scene starts a new page.
func (p *webtoonPager) background(limit int) (int, bool) {
	if p.rowSpread(p.top) <= webtoonGapTolerance {
		return p.rowGray(p.top), true
	}
	if gaps := p.gaps(p.top, limit); len(gaps) > 0 {
		return gaps[0].gray, true
	}
	return 0, false
}

// page copy the rows [y0, y1) of the chapter.
func (p *webtoonPager) page(y0, y1 int) *image.RGBA {
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
func (p *webtoonPager) next(last bool) (*image.RGBA, bool) {
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
	case rest >= p.pageH:
		y, hard, ok := p.cut(last)
		if !ok {
			return nil, false
		}
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

// padWebtoonPage fills a page shorter than the view (pageH) up to it, with the page centered.
//
// The rows added above and below take the color of the edge of the page, when most of it has this color:
// a dark scene stays dark on the whole page, instead of a black strip in the middle of the background of the view,
// and a cut through a text on a white background stays white.
// Otherwise the edge takes the color of the other one, or the background when none has a main color.
func padWebtoonPage(page *image.RGBA, pageH int, background color.Color) image.Image {
	b := page.Bounds()
	if b.Dy() >= pageH || b.Dy() == 0 {
		return page
	}

	edge := func(y int) (color.Color, bool) {
		grays := make([]int, b.Dx())
		for x := range grays {
			grays[x] = int(color.GrayModel.Convert(page.At(b.Min.X+x, y)).(color.Gray).Y)
		}
		sorted := slices.Clone(grays)
		slices.Sort(sorted)
		median := sorted[len(sorted)/2]

		near := 0
		for _, v := range grays {
			if abs(v-median) <= webtoonGapTolerance {
				near++
			}
		}
		if near*100 < len(grays)*webtoonEdgeMainColor {
			return nil, false
		}
		return page.At(b.Min.X+slices.Index(grays, median), y), true
	}
	top, topFlat := edge(b.Min.Y)
	bottom, bottomFlat := edge(b.Max.Y - 1)
	switch {
	case !topFlat && !bottomFlat:
		top, bottom = background, background
	case !topFlat:
		top = bottom
	case !bottomFlat:
		bottom = top
	}

	dst := image.NewRGBA(image.Rect(0, 0, b.Dx(), pageH))
	offset := (pageH - b.Dy()) / 2
	draw.Draw(dst, image.Rect(0, 0, b.Dx(), offset), image.NewUniform(top), image.Point{}, draw.Src)
	draw.Draw(dst, image.Rect(0, offset, b.Dx(), offset+b.Dy()), page, b.Min, draw.Src)
	draw.Draw(dst, image.Rect(0, offset+b.Dy(), b.Dx(), pageH), image.NewUniform(bottom), image.Point{}, draw.Src)
	return dst
}

// webtoonPages cuts the webtoon into pages of the size of the view.
//
// The images of a directory are the pieces of a long vertical strip: they are put back in order and
// joined, then cut into pages with the aspect ratio of the view (see WebtoonAspectRatio), in the gaps between the panels when possible.
// Each directory is a chapter, the pages keep the path of the directory, and are named after the first image
// they come from. A page shorter than the view is filled up to it (see padWebtoonPage).
// progress is called for each source image read.
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

		background := utils.StyleColor(e.Image.IsGrayScale(), e.Image.GrayScaleMode, e.Image.View.Color.Background)
		send := func(img *image.RGBA) {
			count++
			out <- task{
				Id:    id,
				Image: padWebtoonPage(img, pager.pageH, background),
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

			// can't be part of the strip: a placeholder is a page of its own, that keeps the error and the name of the image
			if t.Image == nil || t.Error != nil {
				flush()
				t.Id = id
				out <- t
				id++
				return
			}

			if pager == nil || t.Path != path {
				flush()
				w := t.Image.Bounds().Dx()
				pager = &webtoonPager{
					width: w,
					pageH: max(1, int(float64(w)*e.Image.WebtoonAspectRatio()+0.5)),
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
