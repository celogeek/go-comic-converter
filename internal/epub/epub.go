package epub

import (
	"fmt"
	"os"
	"path/filepath"
	"regexp"
	"sort"
	"strings"
	"text/template"
	"time"

	epubimage "github.com/celogeek/go-comic-converter/v2/internal/epub/image"
	epubtemplates "github.com/celogeek/go-comic-converter/v2/internal/epub/templates"
	epubzip "github.com/celogeek/go-comic-converter/v2/internal/epub/zip"
	"github.com/gofrs/uuid"
)

type Options struct {
	Input                      string
	Output                     string
	Title                      string
	Author                     string
	LimitMb                    int
	StripFirstDirectoryFromToc bool
	Dry                        bool
	DryVerbose                 bool
	SortPathMode               int
	Quiet                      bool
	Workers                    int
	Image                      *epubimage.Options
}

type ePub struct {
	*Options
	UID       string
	Publisher string
	UpdatedAt string

	templateProcessor *template.Template
}

type epubPart struct {
	Cover  *epubimage.Image
	Images []*epubimage.Image
}

func New(options *Options) *ePub {
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
		Options:           options,
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

func (e *ePub) writeImage(wz *epubzip.EpubZip, img *epubimage.Image) error {
	err := wz.WriteFile(
		fmt.Sprintf("OEBPS/%s", img.TextPath()),
		e.render(epubtemplates.Text, map[string]any{
			"Title":      fmt.Sprintf("Image %d Part %d", img.Id, img.Part),
			"ViewPort":   fmt.Sprintf("width=%d,height=%d", e.Image.ViewWidth, e.Image.ViewHeight),
			"ImagePath":  img.ImgPath(),
			"ImageStyle": img.ImgStyle(e.Image.ViewWidth, e.Image.ViewHeight, e.Image.Manga),
		}),
	)

	if err == nil {
		err = wz.WriteImage(img.Data)
	}

	return err
}

func (e *ePub) writeBlank(wz *epubzip.EpubZip, img *epubimage.Image) error {
	return wz.WriteFile(
		fmt.Sprintf("OEBPS/Text/%d_sp.xhtml", img.Id),
		e.render(epubtemplates.Blank, map[string]any{
			"Title":    fmt.Sprintf("Blank Page %d", img.Id),
			"ViewPort": fmt.Sprintf("width=%d,height=%d", e.Image.ViewWidth, e.Image.ViewHeight),
		}),
	)
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
	if e.Image.HasCover {
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
	// descriptor files + title
	baseSize := uint64(16*1024) + cover.Data.CompressedSize()
	if e.Image.HasCover {
		baseSize += cover.Data.CompressedSize()
	}

	currentSize := baseSize
	currentImages := make([]*epubimage.Image, 0)
	part := 1

	for _, img := range images {
		imgSize := img.Data.CompressedSize() + xhtmlSize
		if maxSize > 0 && len(currentImages) > 0 && currentSize+imgSize > maxSize {
			parts = append(parts, &epubPart{
				Cover:  cover,
				Images: currentImages,
			})
			part += 1
			currentSize = baseSize
			if !e.Image.HasCover {
				currentSize += cover.Data.CompressedSize()
			}
			currentImages = make([]*epubimage.Image, 0)
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

	if e.Dry {
		p := epubParts[0]
		fmt.Fprintf(os.Stderr, "TOC:\n  - %s\n%s\n", e.Title, e.getTree(p.Images, true))
		if e.DryVerbose {
			if e.Image.HasCover {
				fmt.Fprintf(os.Stderr, "Cover:\n%s\n", e.getTree([]*epubimage.Image{p.Cover}, false))
			}
			fmt.Fprintf(os.Stderr, "Files:\n%s\n", e.getTree(p.Images, false))
		}
		return nil
	}

	totalParts := len(epubParts)

	bar := e.NewBar(totalParts, "Writing Part", 2, 2)
	for i, part := range epubParts {
		ext := filepath.Ext(e.Output)
		suffix := ""
		if totalParts > 1 {
			fmtLen := len(fmt.Sprint(totalParts))
			fmtPart := fmt.Sprintf(" Part %%0%dd of %%0%dd", fmtLen, fmtLen)
			suffix = fmt.Sprintf(fmtPart, i+1, totalParts)
		}

		path := fmt.Sprintf("%s%s%s", e.Output[0:len(e.Output)-len(ext)], suffix, ext)
		wz, err := epubzip.New(path)
		if err != nil {
			return err
		}
		defer wz.Close()

		title := e.Title
		if totalParts > 1 {
			title = fmt.Sprintf("%s [%d/%d]", title, i+1, totalParts)
		}

		content := []zipContent{
			{"META-INF/container.xml", epubtemplates.Container},
			{"META-INF/com.apple.ibooks.display-options.xml", epubtemplates.AppleBooks},
			{"OEBPS/content.opf", e.getContent(title, part, i+1, totalParts).String()},
			{"OEBPS/toc.xhtml", e.getToc(title, part.Images)},
			{"OEBPS/Text/style.css", e.render(epubtemplates.Style, map[string]any{
				"PageWidth":  e.Image.ViewWidth,
				"PageHeight": e.Image.ViewHeight,
			})},
			{"OEBPS/Text/title.xhtml", e.render(epubtemplates.Text, map[string]any{
				"Title":      title,
				"ViewPort":   fmt.Sprintf("width=%d,height=%d", e.Image.ViewWidth, e.Image.ViewHeight),
				"ImagePath":  "Images/title.jpg",
				"ImageStyle": part.Cover.ImgStyle(e.Image.ViewWidth, e.Image.ViewHeight, e.Image.Manga),
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
		if err := wz.WriteImage(e.createTitleImageData(title, part.Cover, i+1, totalParts)); err != nil {
			return err
		}

		// Cover exist or part > 1
		// If no cover, part 2 and more will include the image as a cover
		if e.Image.HasCover || i > 0 {
			if err := e.writeImage(wz, part.Cover); err != nil {
				return err
			}
		}

		for i, img := range part.Images {
			if err := e.writeImage(wz, img); err != nil {
				return err
			}

			// Double Page or Last Image
			if img.DoublePage || (i+1 == len(part.Images)) {
				if err := e.writeBlank(wz, img); err != nil {
					return err
				}
			}
		}
		bar.Add(1)
	}
	bar.Close()

	return nil
}
