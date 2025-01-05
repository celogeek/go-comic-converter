// Package epub Tools to create EPUB from images.
package epub

import (
	"archive/zip"
	"fmt"
	"math"
	"path/filepath"
	"regexp"
	"sort"
	"strings"
	"text/template"
	"time"

	"github.com/gofrs/uuid"

	"github.com/celogeek/go-comic-converter/v3/internal/pkg/epubimage"
	"github.com/celogeek/go-comic-converter/v3/internal/pkg/epubimageprocessor"
	"github.com/celogeek/go-comic-converter/v3/internal/pkg/epubprogress"
	"github.com/celogeek/go-comic-converter/v3/internal/pkg/epubtemplates"
	"github.com/celogeek/go-comic-converter/v3/internal/pkg/epubtree"
	"github.com/celogeek/go-comic-converter/v3/internal/pkg/epubzip"
	"github.com/celogeek/go-comic-converter/v3/internal/pkg/utils"
	"github.com/celogeek/go-comic-converter/v3/pkg/epuboptions"
)

type EPUB interface {
	Write() error
}

type epub struct {
	epuboptions.EPUBOptions
	UID       string
	Publisher string
	UpdatedAt string

	templateProcessor *template.Template
	imageProcessor    epubimageprocessor.EPUBImageProcessor
}

type epubPart struct {
	Cover  epubimage.EPUBImage
	Images []epubimage.EPUBImage
}

// New initialize EPUB
func New(options epuboptions.EPUBOptions) EPUB {
	uid := uuid.Must(uuid.NewV4())
	tmpl := template.New("parser")
	tmpl.Funcs(template.FuncMap{
		"mod":  func(i, j int) bool { return i%j == 0 },
		"zoom": func(s int, z float32) int { return int(float32(s) * z) },
	})

	return epub{
		EPUBOptions:       options,
		UID:               uid.String(),
		Publisher:         "GO Comic Converter",
		UpdatedAt:         time.Now().UTC().Format("2006-01-02T15:04:05Z"),
		templateProcessor: tmpl,
		imageProcessor:    epubimageprocessor.New(options),
	}
}

// render templates
func (e epub) render(templateString string, data map[string]any) string {
	var result strings.Builder
	tmpl := template.Must(e.templateProcessor.Parse(templateString))
	if err := tmpl.Execute(&result, data); err != nil {
		panic(err)
	}
	return regexp.MustCompile("\n+").ReplaceAllString(result.String(), "\n")
}

// write image to the zip
func (e epub) writeImage(wz epubzip.EPUBZip, img epubimage.EPUBImage, zipImg *zip.File) error {
	err := wz.WriteContent(
		img.EPUBPagePath(),
		[]byte(e.render(epubtemplates.Text, map[string]any{
			"Title":      "Image " + utils.IntToString(img.Id) + " Part " + utils.IntToString(img.Part),
			"ViewPort":   e.Image.View.Port(),
			"ImagePath":  img.ImgPath(),
			"ImageStyle": img.ImgStyle(e.Image.View.Width, e.Image.View.Height, ""),
		})),
	)
	if err == nil {
		err = wz.Copy(zipImg)
	}

	return err
}

// write blank page
func (e epub) writeBlank(wz epubzip.EPUBZip, img epubimage.EPUBImage) error {
	return wz.WriteContent(
		img.EPUBSpacePath(),
		[]byte(e.render(epubtemplates.Blank, map[string]any{
			"Title":    "Blank Page " + utils.IntToString(img.Id),
			"ViewPort": e.Image.View.Port(),
		})),
	)
}

// write title image
func (e epub) writeCoverImage(wz epubzip.EPUBZip, img epubimage.EPUBImage, part, totalParts int) error {
	title := "Cover"
	text := ""
	if totalParts > 1 {
		text = utils.IntToString(part) + " / " + utils.IntToString(totalParts)
		title = title + " " + text
	}

	if err := wz.WriteContent(
		"OEBPS/Text/cover.xhtml",
		[]byte(e.render(epubtemplates.Text, map[string]any{
			"Title":      title,
			"ViewPort":   e.Image.View.Port(),
			"ImagePath":  "Images/cover." + e.Image.Format,
			"ImageStyle": img.ImgStyle(e.Image.View.Width, e.Image.View.Height, ""),
		})),
	); err != nil {
		return err
	}

	coverTitle, err := e.imageProcessor.CoverTitleData(epubimageprocessor.CoverTitleDataOptions{
		Src:         img.Raw,
		Name:        "cover",
		Text:        text,
		Align:       "bottom",
		PctWidth:    50,
		PctMargin:   50,
		MaxFontSize: 96,
		BorderSize:  8,
	})

	if err != nil {
		return err
	}

	if err := wz.WriteRaw(coverTitle); err != nil {
		return err
	}

	return nil
}

// write title image
func (e epub) writeTitleImage(wz epubzip.EPUBZip, img epubimage.EPUBImage, title string) error {
	titleAlign := ""
	if !e.Image.View.PortraitOnly {
		if e.Image.Manga {
			titleAlign = "right:0"
		} else {
			titleAlign = "left:0"
		}
	}

	if !e.Image.View.PortraitOnly {
		if err := wz.WriteContent(
			"OEBPS/Text/space_title.xhtml",
			[]byte(e.render(epubtemplates.Blank, map[string]any{
				"Title":    "Blank Page Title",
				"ViewPort": e.Image.View.Port(),
			})),
		); err != nil {
			return err
		}
	}

	if err := wz.WriteContent(
		"OEBPS/Text/title.xhtml",
		[]byte(e.render(epubtemplates.Text, map[string]any{
			"Title":      title,
			"ViewPort":   e.Image.View.Port(),
			"ImagePath":  "Images/title." + e.Image.Format,
			"ImageStyle": img.ImgStyle(e.Image.View.Width, e.Image.View.Height, titleAlign),
		})),
	); err != nil {
		return err
	}

	coverTitle, err := e.imageProcessor.CoverTitleData(epubimageprocessor.CoverTitleDataOptions{
		Src:         img.Raw,
		Name:        "title",
		Text:        title,
		Align:       "center",
		PctWidth:    100,
		PctMargin:   100,
		MaxFontSize: 64,
		BorderSize:  4,
	})
	if err != nil {
		return err
	}

	if err := wz.WriteRaw(coverTitle); err != nil {
		return err
	}

	return nil
}

// extract image and split it into part
func (e epub) getParts() (parts []epubPart, imgStorage epubzip.StorageImageReader, err error) {
	images, err := e.imageProcessor.Load()

	if err != nil {
		return
	}

	// sort result by id and part
	sort.Slice(images, func(i, j int) bool {
		if images[i].Id == images[j].Id {
			return images[i].Part < images[j].Part
		}
		return images[i].Id < images[j].Id
	})

	parts = make([]epubPart, 0)
	cover := images[0]
	if e.Image.HasCover || (cover.DoublePage && !e.Image.KeepDoublePageIfSplit) {
		images = images[1:]
	}

	if e.Dry {
		parts = append(parts, epubPart{
			Cover:  cover,
			Images: images,
		})
		return
	}

	imgStorage, err = epubzip.NewStorageImageReader(e.ImgStorage())
	if err != nil {
		return
	}

	// compute size of the EPUB part and try to be as close as possible of the target
	maxSize := uint64(e.LimitMb * 1024 * 1024)
	xhtmlSize := uint64(1024)
	// descriptor files + title + cover
	baseSize := uint64(128*1024) + imgStorage.Size(cover.EPUBImgPath())*2

	currentSize := baseSize
	currentImages := make([]epubimage.EPUBImage, 0)
	part := 1

	for _, img := range images {
		imgSize := imgStorage.Size(img.EPUBImgPath()) + xhtmlSize
		if maxSize > 0 && len(currentImages) > 0 && currentSize+imgSize > maxSize {
			parts = append(parts, epubPart{
				Cover:  cover,
				Images: currentImages,
			})
			part += 1
			currentSize = baseSize
			currentImages = make([]epubimage.EPUBImage, 0)
		}
		currentSize += imgSize
		currentImages = append(currentImages, img)
	}
	if len(currentImages) > 0 {
		parts = append(parts, epubPart{
			Cover:  cover,
			Images: currentImages,
		})
	}

	return
}

// create a tree from the directories.
//
// this is used to simulate the toc.
func (e epub) getTree(images []epubimage.EPUBImage, skipFiles bool) string {
	t := epubtree.New()
	for _, img := range images {
		if skipFiles {
			t.Add(img.Path)
		} else {
			t.Add(filepath.Join(img.Path, img.Name))
		}
	}
	c := t.Root()
	if skipFiles && e.StripFirstDirectoryFromToc && c.ChildCount() == 1 {
		c = c.FirstChild()
	}

	return c.WriteString("")
}

func (e epub) computeAspectRatio(epubParts []epubPart) float64 {
	var (
		bestAspectRatio      float64
		bestAspectRatioCount int
		aspectRatio          = map[float64]int{}
	)

	trunc := func(v float64) float64 {
		return float64(math.Round(v*10000)) / 10000
	}

	for _, p := range epubParts {
		aspectRatio[trunc(p.Cover.OriginalAspectRatio)]++
		for _, i := range p.Images {
			aspectRatio[trunc(i.OriginalAspectRatio)]++
		}
	}

	for k, v := range aspectRatio {
		if v > bestAspectRatioCount {
			bestAspectRatio, bestAspectRatioCount = k, v
		}
	}

	return bestAspectRatio
}

func (e epub) computeViewPort(epubParts []epubPart) (int, int) {
	if e.Image.View.AspectRatio == -1 {
		//keep device size
		return e.Image.View.Width, e.Image.View.Height
	}

	// readjusting view port
	bestAspectRatio := e.Image.View.AspectRatio
	if bestAspectRatio == 0 {
		bestAspectRatio = e.computeAspectRatio(epubParts)
	}

	viewWidth, viewHeight := int(float64(e.Image.View.Height)/bestAspectRatio), int(float64(e.Image.View.Width)*bestAspectRatio)
	if viewWidth > e.Image.View.Width {
		return e.Image.View.Width, viewHeight
	} else {
		return viewWidth, e.Image.View.Height
	}
}

func (e epub) writePart(path string, currentPart, totalParts int, part epubPart, imgStorage epubzip.StorageImageReader) error {
	hasTitlePage := e.TitlePage == 1 || (e.TitlePage == 2 && totalParts > 1)

	wz, err := epubzip.New(path)
	if err != nil {
		return err
	}
	defer func(wz epubzip.EPUBZip) {
		_ = wz.Close()
	}(wz)

	title := e.Title
	if totalParts > 1 {
		title = title + " [" + utils.IntToString(currentPart) + "/" + utils.IntToString(totalParts) + "]"
	}

	type zipContent struct {
		Name    string
		Content string
	}
	content := []zipContent{
		{"META-INF/container.xml", epubtemplates.Container},
		{"META-INF/com.apple.ibooks.display-options.xml", epubtemplates.AppleBooks},
		{"OEBPS/content.opf", epubtemplates.Content{
			Title:        title,
			HasTitlePage: hasTitlePage,
			UID:          e.UID,
			Author:       e.Author,
			Publisher:    e.Publisher,
			UpdatedAt:    e.UpdatedAt,
			ImageOptions: e.Image,
			Cover:        part.Cover,
			Images:       part.Images,
			Current:      currentPart,
			Total:        totalParts,
		}.String()},
		{"OEBPS/toc.xhtml", epubtemplates.Toc(title, hasTitlePage, e.StripFirstDirectoryFromToc, part.Images)},
		{"OEBPS/Text/style.css", e.render(epubtemplates.Style, map[string]any{
			"View": e.Image.View,
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

	if err = e.writeCoverImage(wz, part.Cover, currentPart, totalParts); err != nil {
		return err
	}

	if hasTitlePage {
		if err = e.writeTitleImage(wz, part.Cover, title); err != nil {
			return err
		}
	}

	lastImage := part.Images[len(part.Images)-1]
	for _, img := range part.Images {
		if err := e.writeImage(wz, img, imgStorage.Get(img.EPUBImgPath())); err != nil {
			return err
		}

		// Double Page or Last Image that is not a double page
		if !e.Image.View.PortraitOnly &&
			(img.DoublePage ||
				(!e.Image.KeepDoublePageIfSplit && img.Part == 1) ||
				(img.Part == 0 && img == lastImage)) {
			if err := e.writeBlank(wz, img); err != nil {
				return err
			}
		}
	}
	return nil
}

// create the zip
func (e epub) Write() error {
	epubParts, imgStorage, err := e.getParts()
	if err != nil {
		return err
	}

	if e.Dry {
		p := epubParts[0]
		utils.Printf("TOC:\n  - %s\n%s\n", e.Title, e.getTree(p.Images, true))
		if e.DryVerbose {
			if e.Image.HasCover {
				utils.Printf("Cover:\n%s\n", e.getTree([]epubimage.EPUBImage{p.Cover}, false))
			}
			utils.Printf("Files:\n%s\n", e.getTree(p.Images, false))
		}
		return nil
	}
	defer func() {
		_ = imgStorage.Close()
		_ = imgStorage.Remove()
	}()

	totalParts := len(epubParts)

	bar := epubprogress.New(epubprogress.Options{
		Max:         totalParts,
		Description: "Writing Part",
		CurrentJob:  2,
		TotalJob:    2,
		Quiet:       e.Quiet,
		Json:        e.Json,
	})

	e.Image.View.Width, e.Image.View.Height = e.computeViewPort(epubParts)
	for i, part := range epubParts {
		ext := filepath.Ext(e.Output)
		suffix := ""
		if totalParts > 1 {
			fmtLen := utils.FormatNumberOfDigits(totalParts)
			fmtPart := "Part " + fmtLen + " of " + fmtLen
			suffix = fmt.Sprintf(fmtPart, i+1, totalParts)
		}

		path := e.Output[0:len(e.Output)-len(ext)] + suffix + ext

		if err := e.writePart(
			path,
			i+1,
			totalParts,
			part,
			imgStorage,
		); err != nil {
			return err
		}

		_ = bar.Add(1)
	}
	_ = bar.Close()
	if !e.Json {
		utils.Println()
	}

	// display corrupted images
	hasError := false
	for pId, part := range epubParts {
		if pId == 0 && e.Image.HasCover && part.Cover.Error != nil {
			hasError = true
			utils.Printf("Error on image %s: %v\n", filepath.Join(part.Cover.Path, part.Cover.Name), part.Cover.Error)
		}
		for _, img := range part.Images {
			if img.Part == 0 && img.Error != nil {
				hasError = true
				utils.Printf("Error on image %s: %v\n", filepath.Join(img.Path, img.Name), img.Error)
			}
		}
	}
	if hasError {
		utils.Println()
	}

	return nil
}
