package epub

import (
	"path/filepath"

	epubimage "github.com/celogeek/go-comic-converter/v2/internal/epub/image"
	epubtree "github.com/celogeek/go-comic-converter/v2/internal/epub/tree"
)

func (e *ePub) getTree(images []*epubimage.Image, skip_files bool) string {
	t := epubtree.New()
	for _, img := range images {
		if skip_files {
			t.Add(img.Path)
		} else {
			t.Add(filepath.Join(img.Path, img.Name))
		}
	}
	c := t.Root()
	if skip_files && e.StripFirstDirectoryFromToc && len(c.Children) == 1 {
		c = c.Children[0]
	}

	return c.ToString("")
}
