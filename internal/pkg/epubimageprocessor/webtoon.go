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
	// the new background is a gap of at least this height (in percent of the page height):
	// a shorter one is a flat part of a drawing (a sky)
	webtoonBackgroundMinRows = 3
	// the page can be longer than the view (up to this ratio, in percent of its height) to end with its scene:
	// the speed lines of a black scene that fade in the white of the next one are not shown on a white page
	webtoonSceneStretch = 150
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
	// end of the last page: the next one can't take the rows above it
	floor int
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

// rowColor main color of a row of the chapter: the color of a gap.
//
// A gap is flat, but not to the pixel: the middle of the row can be a light shade (the circle behind a title).
func (p *webtoonPager) rowColor(y int) color.Color {
	img, sy, ok := p.strip(y)
	if !ok {
		return color.Black
	}
	b := img.Bounds()
	x := b.Min.X + b.Dx()/2
	if _, mx, ok := mainGray(img, b.Min.X, b.Max.X, sy); ok {
		x = mx
	}
	return img.At(x, sy)
}

// rowGray main gray of a row of the chapter.
func (p *webtoonPager) rowGray(y int) int {
	return grayOf(p.rowColor(y))
}

// rowMain main gray of a row of the chapter, when most of it has it (see mainGray).
func (p *webtoonPager) rowMain(y int) (int, bool) {
	img, sy, ok := p.strip(y)
	if !ok {
		return 0, false
	}
	b := img.Bounds()
	g, _, ok := mainGray(img, b.Min.X, b.Max.X, sy)
	return g, ok
}

// gapColor color of the row y when it is in a gap, nil otherwise.
func (p *webtoonPager) gapColor(y int) color.Color {
	if y >= p.end || p.rowSpread(y) > webtoonGapTolerance {
		return nil
	}
	return p.rowColor(y)
}

// gapAbove color of the gap the page that starts at y was cut in: the row above, or the first row of the page
// when the page starts where the gap ends. nil without a gap.
//
// The row above comes first: the first row can be the border of the panel, that is flat too.
func (p *webtoonPager) gapAbove(y int) color.Color {
	if y > 0 {
		if c := p.gapColor(y - 1); c != nil {
			return c
		}
	}
	return p.gapColor(y)
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
//   - the first change of background on the page (see background), even a bit after the end of the page,
//   - the biggest gap of the window at the end of the page,
//   - the last gap before the window, the page is shorter than the view,
//   - the first gap after the end of the page, the page is longer than the view and will be shrunk,
//   - otherwise the cut goes through the drawing, at the end of the page (hard).
//
// The first and the last ones need the rows that follow the end of the page: ok is false when they are not all known yet,
// and the cut must wait for more rows (it never waits at the end of the chapter, last).
func (p *webtoonPager) cut(last bool) (y int, hard, ok bool) {
	limit := p.top + p.pageH
	from := limit - p.pageH*webtoonCutWindow/100
	minEnd := p.top + p.pageH*webtoonMinFill/100

	sceneEnd := p.top + p.pageH*webtoonSceneStretch/100
	if !last && p.end < sceneEnd {
		return 0, false, false
	}
	if bg := p.background(limit); bg != nil {
		for _, g := range p.gaps(minEnd, sceneEnd) {
			if abs(g.gray-grayOf(bg)) > webtoonBackgroundChange && g.rows*100 >= p.pageH*webtoonBackgroundMinRows {
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

// background color of the page that starts at p.top: the one of the gap it starts in or after,
// or of its first gap before limit.
//
// A page doesn't mix 2 backgrounds: a white panel after a black scene starts a new page.
func (p *webtoonPager) background(limit int) color.Color {
	if c := p.gapAbove(p.top); c != nil {
		return c
	}
	if gaps := p.gaps(p.top, limit); len(gaps) > 0 {
		return p.rowColor(gaps[0].at)
	}
	return nil
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

// grow takes the rows of the gaps around the page [y0, y1), up to the height of the view, centered when possible:
// a short panel is shown with the real background around it. A side is only grown with the rows of its color (nil: not grown),
// the border of the next panel is flat too.
func (p *webtoonPager) grow(y0, y1 int, top, bottom color.Color) (int, int) {
	need := p.pageH - (y1 - y0)
	if need <= 0 {
		return y0, y1
	}

	same := func(y int, c color.Color) bool {
		return c != nil && p.rowSpread(y) <= webtoonGapTolerance && abs(p.rowGray(y)-grayOf(c)) <= webtoonGapTolerance
	}
	above, below := 0, 0
	for above < need && y0-above-1 >= p.floor && same(y0-above-1, top) {
		above++
	}
	for below < need && y1+below < p.end && same(y1+below, bottom) {
		below++
	}

	half := need / 2
	switch {
	case above >= half && below >= need-half:
		above, below = half, need-half
	case above < half:
		below = min(below, need-above)
	default:
		above = min(above, need-below)
	}
	return y0 - above, y1 + below
}

// webtoonPage a page of a webtoon, with the colors to fill it up to the view: nil when they are not known.
type webtoonPage struct {
	img         *image.RGBA
	top, bottom color.Color
}

// frame the page [y0, y1), with the background around it when the page is shorter than the view.
//
// A gap above or below the page is its background when it has the color of the page (see background),
// or of the edge of the page on this side (the speed lines of a black scene that end in a white gap),
// or when the page has no background.
// Otherwise it is the background of the next scene: a black scene after a white panel is not shown below it.
// The page takes the rows of its background (see grow), and the color of the side to fill the rest:
// its background, or the one of the page. It returns the end of the page with the rows taken below.
func (p *webtoonPager) frame(y0, y1 int) (webtoonPage, int) {
	bg := p.background(y1)
	side := func(gap color.Color, edge int) color.Color {
		if gap == nil {
			return nil
		}
		// without a background, the gap is the best guess
		if bg == nil {
			return gap
		}
		// the same color, as flat as a gap
		if abs(grayOf(gap)-grayOf(bg)) <= webtoonGapTolerance {
			return gap
		}
		if g, ok := p.rowMain(edge); ok && abs(grayOf(gap)-g) <= webtoonGapTolerance {
			return gap
		}
		return nil
	}

	top, bottom := side(p.gapAbove(y0), y0), side(p.gapColor(y1), y1-1)

	y0, y1 = p.grow(y0, y1, top, bottom)
	if top == nil {
		top = bg
	}
	if bottom == nil {
		bottom = bg
	}
	return webtoonPage{img: p.page(y0, y1), top: top, bottom: bottom}, y1
}

// next the next page of the chapter, when there is enough to fill it, or when it is the end (last).
func (p *webtoonPager) next(last bool) (webtoonPage, bool) {
	// a blank end is not a page
	if blank := p.trim(); blank && last {
		p.top = p.end
		p.drop()
		return webtoonPage{}, false
	}

	rest := p.end - p.top
	switch {
	case rest <= 0:
		return webtoonPage{}, false
	case rest >= p.pageH:
		y, hard, ok := p.cut(last)
		if !ok {
			return webtoonPage{}, false
		}
		// the rows taken below are not shown again
		page, end := p.frame(p.top, y)
		p.top, p.floor = end, end
		// nothing to overlap at the end of the chapter
		if hard && y < p.end {
			p.top -= p.pageH * webtoonOverlap / 100
		}
		p.drop()
		return page, true
	case !last:
		return webtoonPage{}, false
	}

	// the end of the chapter
	page, _ := p.frame(p.top, p.end)
	p.top, p.floor = p.end, p.end
	p.drop()
	return page, true
}

// padWebtoonPage fills a page shorter than the view (pageH) up to it, with the page centered.
//
// The rows added above and below take the color of the gap the page was cut in:
// a dark scene stays dark on the whole page, instead of a black strip in the middle of the background of the view.
// Without a gap (a cut through a drawing or a text), they take the main color of the edge of the page,
// when most of it has this color: a cut through a text on a white background stays white.
// Otherwise the edge takes the color of the other one, or the background when none has a color.
func padWebtoonPage(p webtoonPage, pageH int, background color.Color) image.Image {
	page := p.img
	b := page.Bounds()
	if b.Dy() >= pageH || b.Dy() == 0 {
		return page
	}

	edge := func(y int) (color.Color, bool) {
		_, x, ok := mainGray(page, b.Min.X, b.Max.X, y)
		if !ok {
			return nil, false
		}
		return page.At(x, y), true
	}
	top, topFlat := p.top, p.top != nil
	if !topFlat {
		top, topFlat = edge(b.Min.Y)
	}
	bottom, bottomFlat := p.bottom, p.bottom != nil
	if !bottomFlat {
		bottom, bottomFlat = edge(b.Max.Y - 1)
	}
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
		send := func(page webtoonPage) {
			count++
			out <- task{
				Id:    id,
				Image: padWebtoonPage(page, pager.pageH, background),
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
				page, ok := pager.next(true)
				if !ok {
					break
				}
				send(page)
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
				page, ok := pager.next(false)
				if !ok {
					break
				}
				send(page)
			}
		})
		flush()
	}()

	return out
}

// grayOf the gray of a color, from 0 to 255.
func grayOf(c color.Color) int {
	return int(color.GrayModel.Convert(c).(color.Gray).Y)
}

// mainGray the main gray of the row y, from x0 to x1 (excluded), when most of the row has it (webtoonEdgeMainColor),
// and the x of a pixel of this gray.
func mainGray(img image.Image, x0, x1, y int) (gray, x int, ok bool) {
	grays := make([]int, x1-x0)
	for i := range grays {
		grays[i] = grayOf(img.At(x0+i, y))
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
		return 0, 0, false
	}
	return median, x0 + slices.Index(grays, median), true
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
