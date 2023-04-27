/*
Tools to create epub from images.
*/
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
	epubimageprocessing "github.com/celogeek/go-comic-converter/v2/internal/epub/imageprocessing"
	epuboptions "github.com/celogeek/go-comic-converter/v2/internal/epub/options"
	epubprogress "github.com/celogeek/go-comic-converter/v2/internal/epub/progress"
	epubtemplates "github.com/celogeek/go-comic-converter/v2/internal/epub/templates"
	epubtree "github.com/celogeek/go-comic-converter/v2/internal/epub/tree"
	epubzip "github.com/celogeek/go-comic-converter/v2/internal/epub/zip"
	"github.com/gofrs/uuid"
)

type ePub struct {
	*epuboptions.Options
	UID       string
	Publisher string
	UpdatedAt string

	templateProcessor *template.Template
}

type epubPart struct {
	Cover        *epubimageprocessing.LoadedImage
	LoadedImages epubimageprocessing.LoadedImages
}

// initialize epub
func New(options *epuboptions.Options) *ePub {
	uid := uuid.Must(uuid.NewV4())
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

// render templates
func (e *ePub) render(templateString string, data map[string]any) string {
	var result strings.Builder
	tmpl := template.Must(e.templateProcessor.Parse(templateString))
	if err := tmpl.Execute(&result, data); err != nil {
		panic(err)
	}
	return regexp.MustCompile("\n+").ReplaceAllString(result.String(), "\n")
}

// write image to the zip
func (e *ePub) writeImage(wz *epubzip.EpubZip, img *epubimageprocessing.LoadedImage) error {
	err := wz.WriteContent(
		fmt.Sprintf("OEBPS/%s", img.Image.TextPath()),
		[]byte(e.render(epubtemplates.Text, map[string]any{
			"Title":      fmt.Sprintf("Image %d Part %d", img.Image.Id, img.Image.Part),
			"ViewPort":   fmt.Sprintf("width=%d,height=%d", e.Image.View.Width, e.Image.View.Height),
			"ImagePath":  img.Image.ImgPath(),
			"ImageStyle": img.Image.ImgStyle(e.Image.View.Width, e.Image.View.Height, e.Image.Manga),
		})),
	)

	if err == nil {
		err = wz.WriteRaw(img.ZipImage)
	}

	return err
}

// write blank page
func (e *ePub) writeBlank(wz *epubzip.EpubZip, img *epubimage.Image) error {
	return wz.WriteContent(
		fmt.Sprintf("OEBPS/Text/%d_sp.xhtml", img.Id),
		[]byte(e.render(epubtemplates.Blank, map[string]any{
			"Title":    fmt.Sprintf("Blank Page %d", img.Id),
			"ViewPort": fmt.Sprintf("width=%d,height=%d", e.Image.View.Width, e.Image.View.Height),
		})),
	)
}

// extract image and split it into part
func (e *ePub) getParts() ([]*epubPart, error) {
	loadedImages, err := epubimageprocessing.LoadImages(e.Options)

	if err != nil {
		return nil, err
	}

	// sort result by id and part
	sort.Slice(loadedImages, func(i, j int) bool {
		if loadedImages[i].Image.Id == loadedImages[j].Image.Id {
			return loadedImages[i].Image.Part < loadedImages[j].Image.Part
		}
		return loadedImages[i].Image.Id < loadedImages[j].Image.Id
	})

	parts := make([]*epubPart, 0)
	cover := loadedImages[0]
	if e.Image.HasCover {
		loadedImages = loadedImages[1:]
	}

	if e.Dry {
		parts = append(parts, &epubPart{
			Cover:        cover,
			LoadedImages: loadedImages,
		})
		return parts, nil
	}

	// compute size of the epub part and try to be as close as possible of the target
	maxSize := uint64(e.LimitMb * 1024 * 1024)
	xhtmlSize := uint64(1024)
	// descriptor files + title
	baseSize := uint64(16*1024) + cover.ZipImage.CompressedSize()
	if e.Image.HasCover {
		baseSize += cover.ZipImage.CompressedSize()
	}

	currentSize := baseSize
	currentImages := make([]*epubimageprocessing.LoadedImage, 0)
	part := 1

	for _, img := range loadedImages {
		imgSize := img.ZipImage.CompressedSize() + xhtmlSize
		if maxSize > 0 && len(currentImages) > 0 && currentSize+imgSize > maxSize {
			parts = append(parts, &epubPart{
				Cover:        cover,
				LoadedImages: currentImages,
			})
			part += 1
			currentSize = baseSize
			if !e.Image.HasCover {
				currentSize += cover.ZipImage.CompressedSize()
			}
			currentImages = make([]*epubimageprocessing.LoadedImage, 0)
		}
		currentSize += imgSize
		currentImages = append(currentImages, img)
	}
	if len(currentImages) > 0 {
		parts = append(parts, &epubPart{
			Cover:        cover,
			LoadedImages: currentImages,
		})
	}

	return parts, nil
}

// create a tree from the directories.
//
// this is used to simulate the toc.
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

	return c.WriteString("")
}

// create the zip
func (e *ePub) Write() error {
	type zipContent struct {
		Name    string
		Content string
	}

	epubParts, err := e.getParts()
	if err != nil {
		return err
	}

	if e.Dry {
		p := epubParts[0]
		fmt.Fprintf(os.Stderr, "TOC:\n  - %s\n%s\n", e.Title, e.getTree(p.LoadedImages.Images(), true))
		if e.DryVerbose {
			if e.Image.HasCover {
				fmt.Fprintf(os.Stderr, "Cover:\n%s\n", e.getTree([]*epubimage.Image{p.Cover.Image}, false))
			}
			fmt.Fprintf(os.Stderr, "Files:\n%s\n", e.getTree(p.LoadedImages.Images(), false))
		}
		return nil
	}

	totalParts := len(epubParts)

	bar := epubprogress.New(epubprogress.Options{
		Max:         totalParts,
		Description: "Writing Part",
		CurrentJob:  2,
		TotalJob:    2,
		Quiet:       e.Quiet,
	})
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
			{"OEBPS/content.opf", epubtemplates.Content(&epubtemplates.ContentOptions{
				Title:        title,
				UID:          e.UID,
				Author:       e.Author,
				Publisher:    e.Publisher,
				UpdatedAt:    e.UpdatedAt,
				ImageOptions: e.Image,
				Cover:        part.Cover.Image,
				Images:       part.LoadedImages.Images(),
				Current:      i + 1,
				Total:        totalParts,
			})},
			{"OEBPS/toc.xhtml", epubtemplates.Toc(title, e.StripFirstDirectoryFromToc, part.LoadedImages.Images())},
			{"OEBPS/Text/style.css", e.render(epubtemplates.Style, map[string]any{
				"PageWidth":  e.Image.View.Width,
				"PageHeight": e.Image.View.Height,
			})},
			{"OEBPS/Text/title.xhtml", e.render(epubtemplates.Text, map[string]any{
				"Title":      title,
				"ViewPort":   fmt.Sprintf("width=%d,height=%d", e.Image.View.Width, e.Image.View.Height),
				"ImagePath":  "Images/title.jpg",
				"ImageStyle": part.Cover.Image.ImgStyle(e.Image.View.Width, e.Image.View.Height, e.Image.Manga),
			})},
		}

		if err = wz.WriteMagic(); err != nil {
			return err
		}
		for _, c := range content {
			if err := wz.WriteContent(c.Name, []byte(c.Content)); err != nil {
				return err
			}
		}
		if err := wz.WriteRaw(epubimageprocessing.CoverTitleData(part.Cover.Image.Raw, title, e.Image.Quality)); err != nil {
			return err
		}

		// Cover exist or part > 1
		// If no cover, part 2 and more will include the image as a cover
		if e.Image.HasCover || i > 0 {
			if err := e.writeImage(wz, part.Cover); err != nil {
				return err
			}
		}

		for i, img := range part.LoadedImages {
			if err := e.writeImage(wz, img); err != nil {
				return err
			}

			// Double Page or Last Image
			if img.Image.DoublePage || (i+1 == len(part.LoadedImages)) {
				if err := e.writeBlank(wz, img.Image); err != nil {
					return err
				}
			}
		}
		bar.Add(1)
	}
	bar.Close()

	return nil
}
