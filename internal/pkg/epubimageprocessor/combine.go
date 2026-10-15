package epubimageprocessor

import (
	"image"
	"image/color"
	"image/draw"
	"math"
	"sort"
)

// Thresholds used to tell that 2 pages are the 2 halves of a double page.
const (
	// the 2 halves must have the same height, with a small tolerance (in percent)
	combineHeightTolerance = 1
	// and about the same width (in percent)
	combineWidthTolerance = 10
	// rows averaged together to compare the edges: the hatching and the screentone are not aligned to the pixel
	combineRows = 4
	// the seam can differ from the difference between 2 columns inside a page by this factor
	combineSeamFactor = 4
	// but not more than that (from 0 to 255), and it is always accepted under combineSeamMin
	combineSeamMax = 32
	combineSeamMin = 4
	// ratio of the edge (in percent) that must not be blank: 2 blank margins always match
	combineMinEdgeContent = 10
	// the edge must change along the page (standard deviation, from 0 to 255):
	// 2 solid edges, black or any color, always match
	combineMinEdgeVariation = 10
)

// combineSplitDoublePages merges the pages that are the 2 halves of a double page, scanned as 2 files.
//
// The tasks are put back in order, then 2 consecutive portrait pages are merged in a landscape page
// when their facing edges continue each other. The merged page keeps the id and the name of the first one,
// the id of the second one is not used. onMerge is called for each page that disappears in a merge.
//
// The input must be fully read, the pages are kept until the previous ones are known.
func (e ePUBImageProcessor) combineSplitDoublePages(in <-chan task, onMerge func()) <-chan task {
	out := make(chan task, e.Workers)

	go func() {
		defer close(out)

		var held *task

		// a page that can be the first half of a double page waits for the next one
		step := func(t task) {
			if held != nil {
				if merged, ok := e.combine(*held, t); ok {
					out <- merged
					held = nil
					onMerge()
					return
				}
				out <- *held
				held = nil
			}
			if e.isCombineCandidate(t) {
				held = &t
				return
			}
			out <- t
		}

		forEachInOrder(in, out, step)

		if held != nil {
			out <- *held
		}
	}()

	return out
}

// forEachInOrder calls step with the tasks in the order of their id, as soon as the previous ones are known.
//
// A fatal error can't be ordered, it is sent as is to out.
// The ids are expected to be continuous, whatever is left at the end is flushed in order.
func forEachInOrder(in <-chan task, out chan<- task, step func(task)) {
	pending := make(map[int]task)
	next := 0

	for t := range in {
		if t.Fatal != nil {
			out <- t
			continue
		}

		pending[t.Id] = t
		for {
			p, ok := pending[next]
			if !ok {
				break
			}
			delete(pending, next)
			next++
			step(p)
		}
	}

	ids := make([]int, 0, len(pending))
	for id := range pending {
		ids = append(ids, id)
	}
	sort.Ints(ids)
	for _, id := range ids {
		step(pending[id])
	}
}

// isCombineCandidate tells if the page can be one half of a double page: a readable portrait page.
func (e ePUBImageProcessor) isCombineCandidate(t task) bool {
	if t.Image == nil || t.Error != nil {
		return false
	}
	b := t.Image.Bounds()
	return b.Dy() > b.Dx()
}

// combine puts the 2 pages side by side, when they are the 2 halves of the same double page.
//
// The first page is on the left, or on the right with the manga mode.
func (e ePUBImageProcessor) combine(first, second task) (task, bool) {
	if !e.isCombineCandidate(second) {
		return task{}, false
	}

	left, right := first.Image, second.Image
	if e.Image.MangaMode() {
		left, right = right, left
	}

	if !isSplitDoublePage(left, right) {
		return task{}, false
	}

	lb, rb := left.Bounds(), right.Bounds()
	h := max(lb.Dy(), rb.Dy())
	dst := e.createImage(first.Image, image.Rect(0, 0, lb.Dx()+rb.Dx(), h))
	draw.Draw(dst, dst.Bounds(), image.NewUniform(color.White), image.Point{}, draw.Src)
	draw.Draw(dst, image.Rect(0, 0, lb.Dx(), lb.Dy()), left, lb.Min, draw.Src)
	draw.Draw(dst, image.Rect(lb.Dx(), 0, lb.Dx()+rb.Dx(), rb.Dy()), right, rb.Min, draw.Src)

	first.Image = dst
	return first, true
}

// isSplitDoublePage tells if the right edge of the left image continues on the left edge of the right image.
//
// The seam is compared to the difference between the 2 outer columns of each page:
// the content of a drawing changes from a column to the next one anyway, so a seam that is as smooth
// as the inside of the pages is a continuation, whatever the drawing.
func isSplitDoublePage(left, right image.Image) bool {
	lb, rb := left.Bounds(), right.Bounds()

	if abs(lb.Dy()-rb.Dy())*100 > max(lb.Dy(), rb.Dy())*combineHeightTolerance {
		return false
	}
	if abs(lb.Dx()-rb.Dx())*100 > max(lb.Dx(), rb.Dx())*combineWidthTolerance {
		return false
	}

	var seam, inner, content, blocks int
	var leftEdge, rightEdge edgeStats
	for y := 0; y+combineRows <= min(lb.Dy(), rb.Dy()); y += combineRows {
		var l0, l1, r0, r1 int
		for i := range combineRows {
			l0 += gray(left.At(lb.Max.X-1, lb.Min.Y+y+i))
			l1 += gray(left.At(lb.Max.X-2, lb.Min.Y+y+i))
			r0 += gray(right.At(rb.Min.X, rb.Min.Y+y+i))
			r1 += gray(right.At(rb.Min.X+1, rb.Min.Y+y+i))
		}
		leftEdge.add(float64(l0) / combineRows)
		rightEdge.add(float64(r0) / combineRows)
		seam += abs(l0 - r0)
		inner += abs(l0-l1) + abs(r0-r1)
		if !isBlankGray(l0/combineRows) || !isBlankGray(r0/combineRows) {
			content++
		}
		blocks++
	}
	if blocks == 0 || content*100 < blocks*combineMinEdgeContent {
		return false
	}
	if max(leftEdge.stdDev(), rightEdge.stdDev()) < combineMinEdgeVariation {
		return false
	}

	// mean per pixel: inner counts 2 pages
	n := blocks * combineRows
	seamMean := float64(seam) / float64(n)
	innerMean := float64(inner) / float64(n) / 2

	return seamMean <= combineSeamMin ||
		(seamMean <= combineSeamMax && seamMean <= innerMean*combineSeamFactor)
}

// edgeStats standard deviation of the values along an edge.
type edgeStats struct {
	n, sum, sumSq float64
}

func (s *edgeStats) add(v float64) {
	s.n++
	s.sum += v
	s.sumSq += v * v
}

func (s *edgeStats) stdDev() float64 {
	if s.n == 0 {
		return 0
	}
	mean := s.sum / s.n
	return math.Sqrt(max(0, s.sumSq/s.n-mean*mean))
}

func gray(c color.Color) int {
	return int(color.GrayModel.Convert(c).(color.Gray).Y)
}

// same threshold as the crop
func isBlankGray(y int) bool {
	return y >= 0xe0
}

func abs(v int) int {
	if v < 0 {
		return -v
	}
	return v
}
