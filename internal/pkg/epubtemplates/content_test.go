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
func TestContentSideOnlyDoublePage(t *testing.T) {
	book := []epubimage.EPUBImage{
		{Id: 0, Format: "jpeg"},
		{Id: 1, Part: 1, Format: "jpeg"},
		{Id: 1, Part: 2, Format: "jpeg"},
		{Id: 2, Format: "jpeg", DoublePage: true},
		{Id: 3, Format: "jpeg"},
	}

	for _, c := range []struct {
		manga bool
		want  []string
	}{
		{false, []string{"", "page-spread-left", "page-spread-right", "page-spread-center", ""}},
		{true, []string{"", "page-spread-right", "page-spread-left", "page-spread-center", ""}},
	} {
		content := Content{
			ImageOptions: epuboptions.Image{Manga: c.manga},
			Images:       book,
		}
		var got []string
		for _, ref := range content.getSpineAuto() {
			got = append(got, ref.attrs["properties"])
		}
		assert.Equal(t, c.want, got, "manga=%v", c.manga)
	}
}
