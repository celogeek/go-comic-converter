package epub

import (
	"encoding/xml"
	"fmt"
	"os"
	"path/filepath"
	"regexp"
	"sort"
	"strings"
	"text/template"
	"time"

	"github.com/gofrs/uuid"
)

type ImageOptions struct {
	Crop                bool
	ViewWidth           int
	ViewHeight          int
	Quality             int
	Algo                string
	Brightness          int
	Contrast            int
	AutoRotate          bool
	AutoSplitDoublePage bool
	NoBlankPage         bool
	Manga               bool
	HasCover            bool
	AddPanelView        bool
	Workers             int
}

type EpubOptions struct {
	Input                      string
	Output                     string
	Title                      string
	Author                     string
	LimitMb                    int
	StripFirstDirectoryFromToc bool
	Dry                        bool
	DryVerbose                 bool
	SortPathMode               int

	*ImageOptions
}

type ePub struct {
	*EpubOptions
	UID       string
	Publisher string
	UpdatedAt string

	templateProcessor *template.Template
}

type epubPart struct {
	Cover  *Image
	Images []*Image
}

func NewEpub(options *EpubOptions) *ePub {
	uid, err := uuid.NewV4()
	if err != nil {
		panic(err)
	}

	tmpl := template.New("parser")
	tmpl.Funcs(template.FuncMap{
		"mod":  func(i, j int) bool { return i%j == 0 },
		"zoom": func(s int, z float32) int { return int(float32(s) * z) },
	})

	return &ePub{
		EpubOptions:       options,
		UID:               uid.String(),
		Publisher:         "GO Comic Converter",
		UpdatedAt:         time.Now().UTC().Format("2006-01-02T15:04:05Z"),
		templateProcessor: tmpl,
	}
}

func (e *ePub) render(templateString string, data any) string {
	tmpl, err := e.templateProcessor.Parse(templateString)
	if err != nil {
		panic(err)
	}
	result := &strings.Builder{}
	if err := tmpl.Execute(result, data); err != nil {
		panic(err)
	}

	stripBlank := regexp.MustCompile("\n+")

	return stripBlank.ReplaceAllString(result.String(), "\n")
}

func (e *ePub) getParts() ([]*epubPart, error) {
	images, err := e.LoadImages()

	if err != nil {
		return nil, err
	}

	sort.Slice(images, func(i, j int) bool {
		if images[i].Id < images[j].Id {
			return true
		} else if images[i].Id == images[j].Id {
			return images[i].Part < images[j].Part
		} else {
			return false
		}
	})

	parts := make([]*epubPart, 0)
	cover := images[0]
	if e.HasCover {
		images = images[1:]
	}

	if e.Dry {
		parts = append(parts, &epubPart{
			Cover:  cover,
			Images: images,
		})
		return parts, nil
	}

	maxSize := uint64(e.LimitMb * 1024 * 1024)

	xhtmlSize := uint64(1024)
	// descriptor files + image
	baseSize := uint64(16*1024) + cover.Data.CompressedSize()

	currentSize := baseSize
	currentImages := make([]*Image, 0)
	part := 1
	imgIsOnRightSide := false

	for _, img := range images {
		imgSize := img.Data.CompressedSize() + xhtmlSize
		if maxSize > 0 && len(currentImages) > 0 && currentSize+imgSize > maxSize {
			parts = append(parts, &epubPart{
				Cover:  cover,
				Images: currentImages,
			})
			part += 1
			imgIsOnRightSide = false
			currentSize = baseSize
			currentImages = make([]*Image, 0)
		}
		currentSize += imgSize
		img.NeedSpace = img.Part == 1 && imgIsOnRightSide
		currentImages = append(currentImages, img)
		imgIsOnRightSide = !imgIsOnRightSide
	}
	if len(currentImages) > 0 {
		parts = append(parts, &epubPart{
			Cover:  cover,
			Images: currentImages,
		})
	}

	return parts, nil
}

func (e *ePub) getToc(images []*Image) *TocChildren {
	paths := map[string]*TocPart{
		".": {},
	}
	for _, img := range images {
		currentPath := "."
		for _, path := range strings.Split(img.Path, string(filepath.Separator)) {
			parentPath := currentPath
			currentPath = filepath.Join(currentPath, path)
			if _, ok := paths[currentPath]; ok {
				continue
			}
			part := &TocPart{
				Title: TocTitle{
					Value: path,
					Link:  fmt.Sprintf("Text/%d_p%d.xhtml", img.Id, img.Part),
				},
			}
			paths[currentPath] = part
			if paths[parentPath].Children == nil {
				paths[parentPath].Children = &TocChildren{}
			}
			paths[parentPath].Children.Tags = append(paths[parentPath].Children.Tags, part)
		}
	}

	children := paths["."].Children

	if children != nil && e.StripFirstDirectoryFromToc && len(children.Tags) == 1 {
		children = children.Tags[0].Children
	}

	return children

}

func (e *ePub) getTree(images []*Image, skip_files bool) string {
	t := NewTree()
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

	return c.toString("")
}

func (e *ePub) Write() error {
	type zipContent struct {
		Name    string
		Content any
	}

	epubParts, err := e.getParts()
	if err != nil {
		return err
	}

	if e.Dry {
		p := epubParts[0]
		fmt.Fprintf(os.Stderr, "TOC:\n  - %s\n%s\n", e.Title, e.getTree(p.Images, true))
		if e.DryVerbose {
			if e.HasCover {
				fmt.Fprintf(os.Stderr, "Cover:\n%s\n", e.getTree([]*Image{p.Cover}, false))
			}
			fmt.Fprintf(os.Stderr, "Files:\n%s\n", e.getTree(p.Images, false))
		}
		return nil
	}

	totalParts := len(epubParts)

	bar := NewBar(totalParts, "Writing Part", 2, 2)
	for i, part := range epubParts {
		ext := filepath.Ext(e.Output)
		suffix := ""
		if totalParts > 1 {
			suffix = fmt.Sprintf(" [%d_%d]", i+1, totalParts)
		}

		path := fmt.Sprintf("%s%s%s", e.Output[0:len(e.Output)-len(ext)], suffix, ext)
		wz, err := newEpubZip(path)
		if err != nil {
			return err
		}
		defer wz.Close()

		title := e.Title
		if totalParts > 1 {
			title = fmt.Sprintf("%s [%d/%d]", title, i+1, totalParts)
		}

		tocChildren := e.getToc(part.Images)
		toc := []byte{}
		if tocChildren != nil {
			toc, err = xml.MarshalIndent(tocChildren.Tags, "        ", "  ")
			if err != nil {
				return err
			}
		}

		content := []zipContent{
			{"META-INF/container.xml", containerTmpl},
			{"OEBPS/content.opf", e.render(contentTmpl, map[string]any{
				"Info":   e,
				"Cover":  part.Cover,
				"Images": part.Images,
				"Title":  title,
				"Part":   i + 1,
				"Total":  totalParts,
			})},
			{"OEBPS/toc.ncx", e.render(tocTmpl, map[string]any{
				"Info":  e,
				"Title": title,
			})},
			{"OEBPS/nav.xhtml", e.render(navTmpl, map[string]any{
				"Title": title,
				"TOC":   string(toc),
			})},
			{"OEBPS/Text/style.css", styleTmpl},
			{"OEBPS/Text/part.xhtml", e.render(partTmpl, map[string]any{
				"Info":  e,
				"Part":  i + 1,
				"Total": totalParts,
			})},
		}
		if e.AddPanelView {
			content = append(content, zipContent{"OEBPS/Text/panelview.css", panelViewTmpl})
		}

		if err = wz.WriteMagic(); err != nil {
			return err
		}
		for _, content := range content {
			if err := wz.WriteFile(content.Name, content.Content); err != nil {
				return err
			}
		}

		// Cover exist or part > 1
		// If no cover, part 2 and more will include the image as a cover
		if e.HasCover || i > 0 {
			wz.WriteImage(part.Cover.Data)
		}

		for _, img := range part.Images {
			var content string
			if e.AddPanelView {
				content = e.render(textTmpl, map[string]any{
					"Image": img,
					"Manga": e.Manga,
				})
			} else {
				content = e.render(textNoPanelTmpl, map[string]any{
					"Image": img,
				})
			}

			if err := wz.WriteFile(fmt.Sprintf("OEBPS/Text/%d_p%d.xhtml", img.Id, img.Part), content); err != nil {
				return err
			}

			if img.NeedSpace {
				if err := wz.WriteFile(
					fmt.Sprintf("OEBPS/Text/%d_sp.xhtml", img.Id),
					e.render(blankTmpl, map[string]any{
						"Info":  e,
						"Image": img,
					}),
				); err != nil {
					return err
				}
			}

			if err := wz.WriteImage(img.Data); err != nil {
				return err
			}
		}
		bar.Add(1)
	}
	bar.Close()

	return nil
}
