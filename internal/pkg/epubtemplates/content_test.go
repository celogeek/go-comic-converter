package epubtemplates

import (
	"fmt"
	"testing"

	"github.com/stretchr/testify/assert"

	"github.com/celogeek/go-comic-converter/v3/internal/pkg/epubimage"
	"github.com/celogeek/go-comic-converter/v3/pkg/epuboptions"
)

// a part can start or end anywhere, even in the middle of a split double page:
// the spine must only refer to pages of the manifest.
func TestContentSpineRefersToManifest(t *testing.T) {
	// 0: page, 1: double page with its 2 halves, 2: page, 3: double page, 4: page
	var book []epubimage.EPUBImage
	for id := range 5 {
		book = append(book, epubimage.EPUBImage{Id: id, Format: "jpeg", DoublePage: id%2 == 1})
		if id%2 == 1 {
			book = append(book,
				epubimage.EPUBImage{Id: id, Part: 1, Format: "jpeg"},
				epubimage.EPUBImage{Id: id, Part: 2, Format: "jpeg"},
			)
		}
	}

	for mode := range 3 {
		for _, manga := range []bool{false, true} {
			for _, device := range []epuboptions.DeviceType{epuboptions.Kindle, epuboptions.Apple} {
				for _, title := range []bool{false, true} {
					for from := range book {
						for to := from + 1; to <= len(book); to++ {
							name := fmt.Sprintf("mode=%d manga=%v device=%v title=%v [%d:%d]", mode, manga, device, title, from, to)
							c := Content{
								HasTitlePage: title,
								ImageOptions: epuboptions.Image{DoublePageMode: mode, Manga: manga, Device: device},
								Images:       book[from:to],
							}

							ids := map[string]int{}
							for _, it := range c.getManifest() {
								ids[it.attrs["id"]]++
							}
							for id, n := range ids {
								assert.Equal(t, 1, n, "%s: duplicate manifest id %s", name, id)
							}
							for _, ref := range c.getSpineAuto() {
								assert.Contains(t, ids, ref.attrs["idref"], "%s: spine refers to a missing page", name)
							}
						}
					}
				}
			}
		}
	}
}

// Only the double pages have a side: a side on every page shows the "Preview double page" on every page on Kindle.
// A blank page fills the spread when the double page would start on the wrong side.
func TestContentSideOnlyDoublePage(t *testing.T) {
	page := func(id int) epubimage.EPUBImage { return epubimage.EPUBImage{Id: id, Format: "jpeg"} }
	split := func(id int) []epubimage.EPUBImage {
		return []epubimage.EPUBImage{{Id: id, Part: 1, Format: "jpeg"}, {Id: id, Part: 2, Format: "jpeg"}}
	}
	double := func(id int) epubimage.EPUBImage { return epubimage.EPUBImage{Id: id, Format: "jpeg", DoublePage: true} }

	for _, c := range []struct {
		name  string
		manga bool
		book  []epubimage.EPUBImage
		want  []string
	}{
		{"split on the first side", false, append([]epubimage.EPUBImage{page(0), page(1)}, append(split(2), page(3))...),
			[]string{"", "", "page-spread-left", "page-spread-right", ""}},
		{"split on the first side, manga", true, append([]epubimage.EPUBImage{page(0), page(1)}, append(split(2), page(3))...),
			[]string{"", "", "page-spread-right", "page-spread-left", ""}},
		{"split on the second side", false, append([]epubimage.EPUBImage{page(0)}, append(split(1), page(2))...),
			[]string{"", "page-spread-right layout-blank", "page-spread-left", "page-spread-right", ""}},
		{"split on the second side, manga", true, append([]epubimage.EPUBImage{page(0)}, append(split(1), page(2))...),
			[]string{"", "page-spread-left layout-blank", "page-spread-right", "page-spread-left", ""}},
		{"double page on the second side", true, []epubimage.EPUBImage{page(0), double(1), page(2)},
			[]string{"", "page-spread-left layout-blank", "page-spread-center", ""}},
		{"double page on the first side", true, []epubimage.EPUBImage{page(0), page(1), double(2), page(3)},
			[]string{"", "", "page-spread-center", ""}},
		// the double page starts a new spread: the next page is on the first side again
		{"split after a double page", true, append([]epubimage.EPUBImage{page(0), double(1)}, append(split(2), page(3))...),
			[]string{"", "page-spread-left layout-blank", "page-spread-center", "page-spread-right", "page-spread-left", ""}},
	} {
		content := Content{ImageOptions: epuboptions.Image{Manga: c.manga}, Images: c.book}

		var got []string
		ids := map[string]bool{}
		for _, it := range content.getManifest() {
			ids[it.attrs["id"]] = true
		}
		for _, ref := range content.getSpineAuto() {
			got = append(got, ref.attrs["properties"])
			assert.Contains(t, ids, ref.attrs["idref"], "%s: spine refers to a missing page", c.name)
		}
		assert.Equal(t, c.want, got, c.name)
	}
}
