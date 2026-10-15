package epubimageprocessor

import (
	"errors"
	"fmt"
	"image"
	"image/color"
	"image/draw"
	"testing"

	"github.com/stretchr/testify/assert"

	"github.com/celogeek/go-comic-converter/v3/pkg/epuboptions"
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
			page, ok := p.next(last == 1)
			if !ok {
				break
			}
			heights = append(heights, page.img.Bounds().Dy())
		}
	}
	return
}

func TestWebtoonPagerCutsInTheGap(t *testing.T) {
	p := &webtoonPager{width: 100, pageH: 100}
	// content, a gap of 15 rows [80, 95) at the end of the first page, then content up to the end
	p.add(testStrip(100, 100, [2]int{0, 80}, [2]int{95, 100}))
	p.add(testStrip(100, 100, [2]int{0, 100}))

	// the first page ends at the start of the gap, and takes it (the page has no background of its own),
	// the second one has no gap to cut in: it goes through the drawing and the next one overlaps it (4 rows)
	assert.Equal(t, []int{95, 100, 9}, testPages(p))
}

func TestWebtoonPagerShortensThePageToKeepThePanel(t *testing.T) {
	p := &webtoonPager{width: 100, pageH: 100}
	// a gap of 10 rows at 60, no gap at the end of the page [75, 100)
	p.add(testStrip(100, 300, [2]int{0, 60}, [2]int{70, 300}))

	// the first page ends in the gap, before the end of the view, and takes it
	assert.Equal(t, []int{70, 100, 100, 38}, testPages(p))
}

func TestWebtoonPagerStretchesThePageToKeepThePanel(t *testing.T) {
	p := &webtoonPager{width: 100, pageH: 100}
	// a panel of 120 rows, then a gap: the page is longer than the view
	p.add(testStrip(100, 300, [2]int{0, 120}, [2]int{130, 300}))

	assert.Equal(t, []int{120, 100, 77}, testPages(p))
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
	// the page is the whole panel, up to the start of the gap [250, 260)
	assert.Equal(t, []int{250, 100, 47}, testPages(p))
}

func TestWebtoonPagerCutsAtAChangeOfBackground(t *testing.T) {
	p := &webtoonPager{width: 100, pageH: 100}
	// a white scene with a white gap at 40, then a black one from 55, that has a white gap in the window at 85
	img := testStrip(100, 300, [2]int{0, 40}, [2]int{48, 55}, [2]int{70, 85}, [2]int{92, 300}).(*image.Gray)
	draw.Draw(img, image.Rect(0, 55, 100, 70), image.NewUniform(color.Black), image.Point{}, draw.Src)
	p.add(img)

	// the first page ends where the black starts, not in the white gap of the window
	assert.Equal(t, []int{55, 100, 100, 41}, testPages(p))
}

func TestWebtoonPagerTakesTheBackgroundAroundTheBox(t *testing.T) {
	p := &webtoonPager{width: 100, pageH: 100}
	// a box on white: its bottom row is gray, but the white of its rounded corners
	img := testStrip(100, 300, [2]int{10, 29}).(*image.Gray)
	draw.Draw(img, image.Rect(5, 29, 95, 30), image.NewUniform(color.Gray{Y: 100}), image.Point{}, draw.Src)
	p.add(img)

	page, ok := p.next(true)
	if !ok {
		t.Fatal("a page is expected")
	}
	// the box [10, 30) is shorter than the view: the page takes the white gap below it, the one above is kept by the trim
	assert.Equal(t, image.Rect(0, 0, 100, 100), page.img.Bounds())
	assert.Equal(t, uint8(255), color.GrayModel.Convert(page.img.At(50, 99)).(color.Gray).Y)
	assert.Equal(t, color.Gray{Y: 255}, page.top)
	assert.Equal(t, color.Gray{Y: 255}, page.bottom)
}

func TestWebtoonPagerDoesntTakeTheBackgroundOfTheNextScene(t *testing.T) {
	p := &webtoonPager{width: 100, pageH: 100}
	// a white gap, a panel, then a black scene
	img := testStrip(100, 300, [2]int{10, 50}, [2]int{80, 300}).(*image.Gray)
	draw.Draw(img, image.Rect(0, 50, 100, 80), image.NewUniform(color.Black), image.Point{}, draw.Src)
	p.add(img)

	page, ok := p.next(true)
	if !ok {
		t.Fatal("a page is expected")
	}
	// the page takes the white above it, ends at the black, and doesn't take it: the rest is filled with the white of the page
	assert.Equal(t, image.Rect(0, 0, 100, 50), page.img.Bounds())
	assert.Equal(t, color.Gray{Y: 255}, page.bottom)
}

func TestWebtoonPagerSkipsBlankEnd(t *testing.T) {
	p := &webtoonPager{width: 100, pageH: 100}
	p.add(testStrip(100, 100, [2]int{0, 50}))
	p.add(testStrip(100, 40))

	// the page ends at the start of the blank, that starts before the end of the page, and takes it up to the view:
	// nothing is left after it
	assert.Equal(t, []int{100}, testPages(p))
}

func TestWebtoonPagerTrimsBlankStart(t *testing.T) {
	p := &webtoonPager{width: 100, pageH: 100}
	p.add(testStrip(100, 300, [2]int{250, 300}))

	// the blank is skipped but a margin of 3 rows: [247, 300), then the page takes the blank above it up to the view
	assert.Equal(t, []int{100}, testPages(p))
}

func TestWebtoonPagerJoinsStripsOfAnyWidth(t *testing.T) {
	p := &webtoonPager{width: 100, pageH: 100}
	p.add(testStrip(100, 60, [2]int{0, 60}))
	p.add(testStrip(50, 20, [2]int{0, 20}))

	// the second strip is stretched to the width of the chapter
	assert.Equal(t, []int{100}, testPages(p))
}

func TestWebtoonPages(t *testing.T) {
	e := ePUBImageProcessor{epuboptions.EPUBOptions{
		Workers: 1,
		Image: epuboptions.Image{
			Mode: epuboptions.ModeWebtoon,
			// the aspect ratio of the view wins over the one of the device
			View: epuboptions.View{Width: 100, Height: 100, AspectRatio: 2},
		},
	}}
	corrupted := errors.New("corrupted")

	in := make(chan task, 4)
	// given in any order
	in <- task{Id: 2, Image: testStrip(100, 300, [2]int{0, 300}), Path: "b", Name: "1.jpg"}
	in <- task{Id: 0, Image: testStrip(100, 300, [2]int{0, 300}), Path: "a", Name: "1.jpg"}
	in <- task{Id: 1, Image: testStrip(10, 10), Path: "a", Name: "2.jpg", Error: corrupted}
	close(in)

	var got []string
	for p := range e.webtoonPages(in, func() {}) {
		got = append(got, fmt.Sprintf("%d %s/%s %d %v", p.Id, p.Path, p.Name, p.Image.Bounds().Dy(), p.Error))
	}

	// pages of 200 rows, without gap: cut through the drawing with an overlap of 8 rows,
	// the last one (108 rows) is filled up to the height of the view.
	// The corrupted image is a page of its own, with its error.
	assert.Equal(t, []string{
		"0 a/1.jpg-0001 200 <nil>",
		"1 a/1.jpg-0002 200 <nil>",
		"2 a/2.jpg 10 corrupted",
		"3 b/1.jpg-0001 200 <nil>",
		"4 b/1.jpg-0002 200 <nil>",
	}, got)
}

func TestPadWebtoonPage(t *testing.T) {
	gray := func(img image.Image, y int) uint8 {
		return color.GrayModel.Convert(img.At(50, y)).(color.Gray).Y
	}
	// a page of 40 rows: [0, 10) top edge, [10, 30) content, [30, 40) bottom edge
	page := func(top, bottom color.Color) *image.RGBA {
		img := image.NewRGBA(image.Rect(0, 0, 100, 40))
		draw.Draw(img, image.Rect(0, 0, 100, 10), image.NewUniform(top), image.Point{}, draw.Src)
		draw.Draw(img, image.Rect(0, 30, 100, 40), image.NewUniform(bottom), image.Point{}, draw.Src)
		for x := 0; x < 100; x += 2 {
			draw.Draw(img, image.Rect(x, 10, x+1, 30), image.NewUniform(color.White), image.Point{}, draw.Src)
		}
		return img
	}
	gray128 := color.Gray{Y: 128}

	for _, c := range []struct {
		name        string
		page        *image.RGBA
		gapTop      color.Color
		gapBottom   color.Color
		top, bottom uint8
	}{
		{"black scene", page(color.Black, color.Black), nil, nil, 0, 0},
		// the bottom of a box: its edge is mostly brown, the page is cut in the white gap below it
		{"the gaps win over the edges", page(color.Black, gray128), color.Black, color.White, 0, 255},
		{"white gutter above a black scene", page(color.White, color.Black), nil, nil, 255, 0},
		{"the bottom is not flat", func() *image.RGBA {
			img := page(color.Black, color.Black)
			for x := 0; x < 100; x += 2 {
				draw.Draw(img, image.Rect(x, 39, x+1, 40), image.NewUniform(color.White), image.Point{}, draw.Src)
			}
			return img
		}(), nil, nil, 0, 0},
		{"a cut through a text on white", func() *image.RGBA {
			img := page(color.Black, color.White)
			// 10% of the bottom edge is the text
			draw.Draw(img, image.Rect(0, 39, 10, 40), image.NewUniform(color.Black), image.Point{}, draw.Src)
			return img
		}(), nil, nil, 0, 255},
		{"no flat edge: the background", func() *image.RGBA {
			img := page(color.Black, color.Black)
			for x := 0; x < 100; x += 2 {
				draw.Draw(img, image.Rect(x, 0, x+1, 1), image.NewUniform(gray128), image.Point{}, draw.Src)
				draw.Draw(img, image.Rect(x, 39, x+1, 40), image.NewUniform(color.White), image.Point{}, draw.Src)
			}
			return img
		}(), nil, nil, 128, 128},
	} {
		t.Run(c.name, func(t *testing.T) {
			got := padWebtoonPage(webtoonPage{img: c.page, top: c.gapTop, bottom: c.gapBottom}, 100, gray128)
			// centered: 30 rows added above, 30 below
			assert.Equal(t, image.Rect(0, 0, 100, 100), got.Bounds())
			assert.Equal(t, c.top, gray(got, 0), "top")
			assert.Equal(t, c.bottom, gray(got, 99), "bottom")
			assert.Equal(t, gray(c.page, 20), gray(got, 50), "content")
		})
	}

	// a page as high as the view is kept
	full := image.NewRGBA(image.Rect(0, 0, 100, 100))
	assert.Same(t, full, padWebtoonPage(webtoonPage{img: full}, 100, gray128))
}
