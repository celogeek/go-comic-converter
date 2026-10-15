package epubimageprocessor

import (
	"image"
	"image/color"
	"image/draw"
	"testing"

	"github.com/stretchr/testify/assert"
)

// testStrip is a white image of the given size, with a band of black stripes for each [from, to) rows.
// A solid band would be a gap, like the blank.
func testStrip(w, h int, bands ...[2]int) image.Image {
	img := image.NewGray(image.Rect(0, 0, w, h))
	draw.Draw(img, img.Bounds(), image.NewUniform(color.White), image.Point{}, draw.Src)
	for _, b := range bands {
		for x := 0; x < w; x += 2 {
			draw.Draw(img, image.Rect(x, b[0], x+1, b[1]), image.NewUniform(color.Black), image.Point{}, draw.Src)
		}
	}
	return img
}

// testPages the height of all the pages of the pager, once the end of the chapter is reached.
func testPages(p *webtoonPager) (heights []int) {
	for last := range []bool{false, true} {
		for {
			img, ok := p.next(last == 1)
			if !ok {
				break
			}
			heights = append(heights, img.Bounds().Dy())
		}
	}
	return
}

func TestWebtoonPagerCutsInTheGap(t *testing.T) {
	p := &webtoonPager{width: 100, pageH: 100}
	// content, a gap of 15 rows [80, 95) at the end of the first page, then content up to the end
	p.add(testStrip(100, 100, [2]int{0, 80}, [2]int{95, 100}))
	p.add(testStrip(100, 100, [2]int{0, 100}))

	// the first page ends in the middle of the gap,
	// the second one has no gap to cut in: it goes through the drawing and the next one overlaps it (4 rows)
	assert.Equal(t, []int{88, 100, 12}, testPages(p))
}

func TestWebtoonPagerShortensThePageToKeepThePanel(t *testing.T) {
	p := &webtoonPager{width: 100, pageH: 100}
	// a gap of 10 rows at 60, no gap at the end of the page [75, 100)
	p.add(testStrip(100, 300, [2]int{0, 60}, [2]int{70, 300}))

	// the first page ends in the gap, before the end of the view
	assert.Equal(t, []int{65, 100, 100, 41}, testPages(p))
}

func TestWebtoonPagerStretchesThePageToKeepThePanel(t *testing.T) {
	p := &webtoonPager{width: 100, pageH: 100}
	// a panel of 120 rows, then a gap: the page is longer than the view
	p.add(testStrip(100, 300, [2]int{0, 120}, [2]int{130, 300}))

	assert.Equal(t, []int{125, 100, 77}, testPages(p))
}

func TestWebtoonPagerWaitsForTheGapOfALargePanel(t *testing.T) {
	p := &webtoonPager{width: 100, pageH: 100}
	// a panel of 250 rows, then a gap
	p.add(testStrip(100, 240, [2]int{0, 240}))

	// the gap is not known yet: no page
	if _, ok := p.next(false); ok {
		t.Fatal("the page must wait for the end of the panel")
	}

	p.add(testStrip(100, 160, [2]int{0, 10}, [2]int{20, 160}))
	// the page is the whole panel, up to the middle of the gap [250, 260)
	assert.Equal(t, []int{255, 100, 47}, testPages(p))
}

func TestWebtoonPagerSkipsBlankEnd(t *testing.T) {
	p := &webtoonPager{width: 100, pageH: 100}
	p.add(testStrip(100, 100, [2]int{0, 50}))
	p.add(testStrip(100, 40))

	// nothing is left but the blank after the cut
	assert.Equal(t, []int{88}, testPages(p))
}

func TestWebtoonPagerTrimsBlankStart(t *testing.T) {
	p := &webtoonPager{width: 100, pageH: 100}
	p.add(testStrip(100, 300, [2]int{250, 300}))

	// the blank is skipped but a margin of 3 rows: [247, 300)
	assert.Equal(t, []int{53}, testPages(p))
}

func TestWebtoonPagerJoinsStripsOfAnyWidth(t *testing.T) {
	p := &webtoonPager{width: 100, pageH: 100}
	p.add(testStrip(100, 60, [2]int{0, 60}))
	p.add(testStrip(50, 20, [2]int{0, 20}))

	// the second strip is stretched to the width of the chapter
	assert.Equal(t, []int{100}, testPages(p))
}
