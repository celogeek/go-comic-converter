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
			for _, device := range []epuboptions.DeviceType{epuboptions.Kindle, epuboptions.Apple, epuboptions.Kobo} {
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
							for _, ref := range c.getSpine() {
								assert.Contains(t, ids, ref.attrs["idref"], "%s: spine refers to a missing page", name)
							}
						}
					}
				}
			}
		}
	}
}

// Kindle: only the double pages have a side, a side on every page shows the "Preview double page" on every page.
// Other devices: no hint, a blank page fills the spread when the double page would start on the wrong side.
func TestContentSideOnlyDoublePage(t *testing.T) {
	page := func(id int) epubimage.EPUBImage { return epubimage.EPUBImage{Id: id, Format: "jpeg"} }
	split := func(id int) []epubimage.EPUBImage {
		return []epubimage.EPUBImage{{Id: id, Part: 1, Format: "jpeg"}, {Id: id, Part: 2, Format: "jpeg"}}
	}
	double := func(id int) epubimage.EPUBImage { return epubimage.EPUBImage{Id: id, Format: "jpeg", DoublePage: true} }
	join := func(parts ...[]epubimage.EPUBImage) (book []epubimage.EPUBImage) {
		for _, p := range parts {
			book = append(book, p...)
		}
		return book
	}
	one := func(i epubimage.EPUBImage) []epubimage.EPUBImage { return []epubimage.EPUBImage{i} }

	for _, c := range []struct {
		name      string
		device    epuboptions.DeviceType
		manga     bool
		landscape bool
		book      []epubimage.EPUBImage
		want      []string // idref[:properties]
	}{
		{"kindle split", epuboptions.Kindle, false, false, join(one(page(0)), split(1), one(page(2))),
			[]string{"page_0_p0", "page_1_p1:page-spread-left", "page_1_p2:page-spread-right", "page_2_p0"}},
		{"kindle split manga", epuboptions.Kindle, true, false, join(one(page(0)), split(1), one(page(2))),
			[]string{"page_0_p0", "page_1_p1:page-spread-right", "page_1_p2:page-spread-left", "page_2_p0"}},
		{"kindle double page", epuboptions.Kindle, true, false, join(one(page(0)), one(double(1)), one(page(2))),
			[]string{"page_0_p0", "page_1_p0:page-spread-center", "page_2_p0"}},
		// Kobo: the first page is on the first side of the spread, Apple on the second side
		{"kobo split on the second side", epuboptions.Kobo, true, false, join(one(page(0)), split(1), one(page(2))),
			[]string{"page_0_p0", "space_1", "page_1_p1", "page_1_p2", "page_2_p0"}},
		{"apple split on the first side", epuboptions.Apple, true, false, join(one(page(0)), split(1), one(page(2))),
			[]string{"page_0_p0", "page_1_p1", "page_1_p2", "page_2_p0"}},
		{"kobo split on the first side", epuboptions.Kobo, false, false, join(one(page(0)), one(page(1)), split(2), one(page(3))),
			[]string{"page_0_p0", "page_1_p0", "page_2_p1", "page_2_p2", "page_3_p0"}},
		{"apple split on the second side", epuboptions.Apple, false, false, join(one(page(0)), one(page(1)), split(2), one(page(3))),
			[]string{"page_0_p0", "page_1_p0", "space_2", "page_2_p1", "page_2_p2", "page_3_p0"}},
		{"remarkable is like kobo", epuboptions.ReMarkable, false, false, join(one(page(0)), one(double(1)), one(page(2))),
			[]string{"page_0_p0", "space_1", "page_1_p0", "page_2_p0"}},
		// the landscape mode is a Kindle spine for now
		{"landscape is not for kobo", epuboptions.Kobo, true, true, join(one(page(0)), split(1), one(page(2))),
			[]string{"page_0_p0", "space_1", "page_1_p1", "page_1_p2", "page_2_p0"}},
		{"apple double page on the first side", epuboptions.Apple, false, false, join(one(page(0)), one(double(1)), one(page(2))),
			[]string{"page_0_p0", "page_1_p0", "page_2_p0"}},
		{"apple double page on the second side", epuboptions.Apple, true, false, join(one(page(0)), one(page(1)), one(double(2))),
			[]string{"page_0_p0", "page_1_p0", "space_2", "page_2_p0"}},
		{"kindle scribe landscape", epuboptions.Kindle, true, true, join(one(page(0)), split(1), one(page(2))),
			[]string{"page_0_p0:page-spread-right", "space_1:page-spread-left", "page_1_p1:page-spread-right", "page_1_p2:page-spread-left", "page_2_p0:page-spread-right", "space_2:page-spread-left"}},
	} {
		content := Content{ImageOptions: epuboptions.Image{Manga: c.manga, Landscape: c.landscape, ProfileLandscape: c.landscape, Device: c.device}, Images: c.book}

		ids := map[string]bool{}
		for _, it := range content.getManifest() {
			ids[it.attrs["id"]] = true
		}
		var got []string
		for _, ref := range content.getSpine() {
			item := ref.attrs["idref"]
			if p := ref.attrs["properties"]; p != "" {
				item += ":" + p
			}
			got = append(got, item)
			assert.Contains(t, ids, ref.attrs["idref"], "%s: spine refers to a missing page", c.name)
		}
		assert.Equal(t, c.want, got, c.name)
	}
}
