package epub

import (
	"fmt"
	"image/color"
	"path/filepath"
	"regexp"
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
	Palette             color.Palette
	Brightness          int
	Contrast            int
	AutoRotate          bool
	AutoSplitDoublePage bool
	NoBlankPage         bool
	Manga               bool
	HasCover            bool
	Workers             int
}

type EpubOptions struct {
	Input   string
	Output  string
	Title   string
	Author  string
	LimitMb int

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

	var spreadRight = options.Manga

	tmpl := template.New("parser")
	tmpl.Funcs(template.FuncMap{
		"mod":  func(i, j int) bool { return i%j == 0 },
		"zoom": func(s int, z float32) int { return int(float32(s) * z) },
		"spread": func() (spread string) {
			if spreadRight {
				spread = "right"
			} else {
				spread = "left"
			}
			spreadRight = !spreadRight
			return
		},
		"spread_blank": func(part int) bool {
			if !options.NoBlankPage && part == 1 && spreadRight == options.Manga {
				return true
			}
			return false
		},
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
	images, err := LoadImages(e.Input, e.ImageOptions)
	if err != nil {
		return nil, err
	}

	parts := make([]*epubPart, 0)
	cover := images[0]
	if e.HasCover {
		images = images[1:]
	}
	if e.LimitMb == 0 {
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

	for _, img := range images {
		imgSize := img.Data.CompressedSize() + xhtmlSize
		if len(currentImages) > 0 && currentSize+imgSize > maxSize {
			parts = append(parts, &epubPart{
				Cover:  cover,
				Images: currentImages,
			})
			part += 1
			currentSize = baseSize
			currentImages = make([]*Image, 0)
		}
		currentSize += imgSize
		currentImages = append(currentImages, img)
	}
	if len(currentImages) > 0 {
		parts = append(parts, &epubPart{
			Cover:  cover,
			Images: currentImages,
		})
	}

	return parts, nil
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
	totalParts := len(epubParts)

	bar := NewBar(totalParts, "Writing Part", 2, 2)
	for i, part := range epubParts {
		ext := filepath.Ext(e.Output)
		suffix := ""
		if totalParts > 1 {
			suffix = fmt.Sprintf(" PART_%02d", i+1)
		}

		path := fmt.Sprintf("%s%s%s", e.Output[0:len(e.Output)-len(ext)], suffix, ext)
		wz, err := newEpubZip(path)
		if err != nil {
			return err
		}
		defer wz.Close()

		content := []zipContent{
			{"META-INF/container.xml", containerTmpl},
			{"OEBPS/content.opf", e.render(contentTmpl, map[string]any{"Info": e, "Cover": part.Cover, "Images": part.Images})},
			{"OEBPS/toc.ncx", e.render(tocTmpl, map[string]any{"Info": e})},
			{"OEBPS/nav.xhtml", e.render(navTmpl, map[string]any{"Info": e})},
			{"OEBPS/Text/style.css", styleTmpl},
			{"OEBPS/Text/part.xhtml", e.render(partTmpl, map[string]any{
				"Info":  e,
				"Part":  i + 1,
				"Total": totalParts,
			})},
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
			if err := wz.WriteFile(
				fmt.Sprintf("OEBPS/Text/%d_p%d.xhtml", img.Id, img.Part),
				e.render(textTmpl, map[string]any{
					"Image": img,
					"Manga": e.Manga,
				}),
			); err != nil {
				return err
			}

			if !e.NoBlankPage && img.Part == 1 {
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
