// Package epub Tools to create EPUB from images.
package epub

import (
	"archive/zip"
	"image"
	"image/draw"
	"math"
	"os"
	"os/signal"
	"path/filepath"
	"regexp"
	"sort"
	"strings"
	"sync"
	"syscall"
	"text/template"
	"time"

	"github.com/gofrs/uuid"

	"github.com/celogeek/go-comic-converter/v3/internal/pkg/epubimage"
	"github.com/celogeek/go-comic-converter/v3/internal/pkg/epubimagefilters"
	"github.com/celogeek/go-comic-converter/v3/internal/pkg/epubimagepassthrough"
	"github.com/celogeek/go-comic-converter/v3/internal/pkg/epubimageprocessor"
	"github.com/celogeek/go-comic-converter/v3/internal/pkg/epubprogress"
	"github.com/celogeek/go-comic-converter/v3/internal/pkg/epubtemplates"
	"github.com/celogeek/go-comic-converter/v3/internal/pkg/epubtree"
	"github.com/celogeek/go-comic-converter/v3/internal/pkg/epubzip"
	"github.com/celogeek/go-comic-converter/v3/internal/pkg/fonts"
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

	imageProcessor epubimageprocessor.EPUBImageProcessor
}

type epubPart struct {
	Cover  epubimage.EPUBImage
	Images []epubimage.EPUBImage
}

// New initialize EPUB
func New(options epuboptions.EPUBOptions) EPUB {
	uid := uuid.Must(uuid.NewV4())
	if options.TrueTypeFont == nil {
		options.TrueTypeFont = fonts.Default
	}

	var imageProcessor epubimageprocessor.EPUBImageProcessor
	if options.Image.Format == "copy" {
		imageProcessor = epubimagepassthrough.New(options)
	} else {
		imageProcessor = epubimageprocessor.New(options)
	}

	return epub{
		EPUBOptions:    options,
		UID:            uid.String(),
		Publisher:      "GO Comic Converter",
		UpdatedAt:      time.Now().UTC().Format("2006-01-02T15:04:05Z"),
		imageProcessor: imageProcessor,
	}
}

var (
	textTemplate  = template.Must(template.New("text").Parse(epubtemplates.Text))
	blankTemplate = template.Must(template.New("blank").Parse(epubtemplates.Blank))
	styleTemplate = template.Must(template.New("style").Parse(epubtemplates.Style))

	multipleNewLines = regexp.MustCompile("\n+")
)

// render templates
func (e epub) render(tmpl *template.Template, data map[string]any) string {
	var result strings.Builder
	if err := tmpl.Execute(&result, data); err != nil {
		panic(err)
	}
	return multipleNewLines.ReplaceAllString(result.String(), "\n")
}

// write image to the zip
func (e epub) writeImage(wz epubzip.EPUBZip, img epubimage.EPUBImage, imgStorage epubzip.StorageImageReader) error {
	styles := img.ImgStyles(e.Image.View.Width, e.Image.View.Height)
	images := make([]epubtemplates.Image, len(styles))
	zipImgs := make([]*zip.File, len(styles))
	for idx, style := range styles {
		images[idx] = epubtemplates.Image{Path: img.TileImgPath(idx), Style: style}
		zipImgs[idx] = imgStorage.Get(img.EPUBTileImgPath(idx))
	}

	var panel *epubtemplates.PanelView
	if e.Image.AddPanelView {
		panel = e.panelView(img, images)
	}

	err := wz.WriteContent(
		img.EPUBPagePath(),
		[]byte(e.render(textTemplate, map[string]any{
			"Title":    "Image " + utils.IntToString(img.Id) + " Part " + utils.IntToString(img.Part),
			"ViewPort": e.Image.View.Port(),
			"Images":   images,
			"Panel":    panel,
		})),
	)
	for i := 0; err == nil && i < len(zipImgs); i++ {
		err = wz.Copy(zipImgs[i])
	}

	return err
}

// panelViewZoom zoom of the panel view, relative to the size of the image in the page.
const panelViewZoom = 1.5

// panelView build the panel view of an image page
func (e epub) panelView(img epubimage.EPUBImage, images []epubtemplates.Image) *epubtemplates.PanelView {
	zoomStyle, zoomStyles := img.PanelZoom(e.Image.View.Width, e.Image.View.Height, panelViewZoom)
	zoomImages := make([]epubtemplates.Image, len(images))
	for i, image := range images {
		zoomImages[i] = epubtemplates.Image{Path: image.Path, Style: zoomStyles[i]}
	}

	// Reading order of the corners: left to right, or right to left for manga.
	corners := []epubtemplates.PanelCorner{{Key: "TL", Ordinal: 1}, {Key: "TR", Ordinal: 2}, {Key: "BL", Ordinal: 3}, {Key: "BR", Ordinal: 4}}
	if e.Image.Manga {
		corners = []epubtemplates.PanelCorner{{Key: "TL", Ordinal: 2}, {Key: "TR", Ordinal: 1}, {Key: "BL", Ordinal: 4}, {Key: "BR", Ordinal: 3}}
	}

	return &epubtemplates.PanelView{
		Style:      img.ImgStyle(e.Image.View.Width, e.Image.View.Height),
		ZoomStyle:  zoomStyle,
		ZoomImages: zoomImages,
		Corners:    corners,
	}
}

// write blank page
func (e epub) writeBlank(wz epubzip.EPUBZip, img epubimage.EPUBImage) error {
	return wz.WriteContent(
		img.EPUBSpacePath(),
		[]byte(e.render(blankTemplate, map[string]any{
			"Title":    "Blank Page " + utils.IntToString(img.Id),
			"ViewPort": e.Image.View.Port(),
		})),
	)
}

// write cover image, used as thumbnail only (it has no page)
func (e epub) writeCoverImage(wz epubzip.EPUBZip, img epubimage.EPUBImage, coverText string) error {
	// the cover is reduced to the max size, the font follows
	coverWidth, _ := e.Image.Tile().FitSize(img.Raw.Bounds().Dx(), img.Raw.Bounds().Dy())
	fontSize := epubimagefilters.TitleFontSize(coverWidth)
	coverTitle, err := e.imageProcessor.CoverTitleData(epubimageprocessor.CoverTitleDataOptions{
		Src:  img.Raw,
		Name: "cover",
		Title: epubimagefilters.CoverTitle{
			Title:       coverText,
			Align:       "bottom",
			FontSize:    fontSize,
			BorderWidth: epubimagefilters.TitleBorderWidth(fontSize),
			Foreground:  utils.StyleColor(e.Image.GrayScale, e.Image.GrayScaleMode, e.Image.View.Color.Foreground),
			Background:  utils.StyleColor(e.Image.GrayScale, e.Image.GrayScaleMode, e.Image.View.Color.Background),
			Font:        e.TrueTypeFont,
		},
	})

	if err != nil {
		return err
	}

	for _, c := range coverTitle {
		if err := wz.WriteRaw(c); err != nil {
			return err
		}
	}

	return nil
}

// tileImages images of a tiled image displayed centered into the view
func (e epub) tileImages(name string, img epubimage.EPUBImage) []epubtemplates.Image {
	styles := img.ImgStyles(e.Image.View.Width, e.Image.View.Height)
	images := make([]epubtemplates.Image, len(styles))
	for idx, style := range styles {
		images[idx] = epubtemplates.Image{Path: "Images/" + epubimage.TileName(name, idx) + ".jpeg", Style: style}
	}

	return images
}

// write title image
func (e epub) writeTitleImage(wz epubzip.EPUBZip, title, titlePage string) error {
	if err := wz.WriteContent(
		"OEBPS/Text/space_title.xhtml",
		[]byte(e.render(blankTemplate, map[string]any{
			"Title":    "Blank Page Title",
			"ViewPort": e.Image.View.Port(),
		})),
	); err != nil {
		return err
	}

	// The title image has the view size, so it fills the page.
	if err := wz.WriteContent(
		"OEBPS/Text/title.xhtml",
		[]byte(e.render(textTemplate, map[string]any{
			"Title":    title,
			"ViewPort": e.Image.View.Port(),
			"Images":   e.tileImages("title", e.Image.Tile().Tiled(e.Image.View.Width, e.Image.View.Height)),
		})),
	); err != nil {
		return err
	}

	foreground := utils.StyleColor(e.Image.GrayScale, e.Image.GrayScaleMode, e.Image.View.Color.Foreground)
	background := utils.StyleColor(e.Image.GrayScale, e.Image.GrayScaleMode, e.Image.View.Color.Background)

	var img draw.Image
	if e.Image.GrayScale {
		img = image.NewGray(image.Rect(0, 0, e.Image.View.Width, e.Image.View.Height))
	} else {
		img = image.NewRGBA(image.Rect(0, 0, e.Image.View.Width, e.Image.View.Height))
	}
	draw.Draw(img, img.Bounds(), image.NewUniform(background), image.Point{}, draw.Src)

	coverTitle, err := e.imageProcessor.CoverTitleData(epubimageprocessor.CoverTitleDataOptions{
		Src:  img,
		Name: "title",
		Title: epubimagefilters.CoverTitle{
			Title:      titlePage,
			Align:      "center",
			FontSize:   epubimagefilters.TitleFontSize(e.Image.View.Width),
			Foreground: foreground,
			Background: background,
			Font:       e.TrueTypeFont,
		},
	})
	if err != nil {
		return err
	}

	for _, c := range coverTitle {
		if err := wz.WriteRaw(c); err != nil {
			return err
		}
	}

	return nil
}

// imgSize size of the image in the storage, with all its tiles
func imgSize(imgStorage epubzip.StorageImageReader, img epubimage.EPUBImage) uint64 {
	var size uint64
	for idx := range img.Tiles() {
		size += imgStorage.Size(img.EPUBTileImgPath(idx))
	}
	return size
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
	// a double page that is split is only kept for the cover, its parts follow it as the pages.
	// the copy mode never splits, the double page is then a page of its own.
	if cover.DoublePage && !e.Image.KeepDoublePage() && len(images) > 1 && images[1].Id == cover.Id {
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
	baseSize := uint64(128*1024) + imgSize(imgStorage, cover)*2

	currentSize := baseSize
	currentImages := make([]epubimage.EPUBImage, 0)
	part := 1

	for _, img := range images {
		size := imgSize(imgStorage, img) + xhtmlSize
		if maxSize > 0 && len(currentImages) > 0 && currentSize+size > maxSize {
			parts = append(parts, epubPart{
				Cover:  cover,
				Images: currentImages,
			})
			part += 1
			currentSize = baseSize
			currentImages = make([]epubimage.EPUBImage, 0)
		}
		currentSize += size
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
	// the ratio of the source can be extreme, as for a 1 pixel wide image
	bestAspectRatio = min(max(bestAspectRatio, epuboptions.MinAspectRatio), epuboptions.MaxAspectRatio)

	viewWidth, viewHeight := int(float64(e.Image.View.Height)/bestAspectRatio), int(float64(e.Image.View.Width)*bestAspectRatio)
	if viewWidth > e.Image.View.Width {
		return e.Image.View.Width, viewHeight
	} else {
		return viewWidth, e.Image.View.Height
	}
}

func (e epub) writePart(path string, currentPart, totalParts int, part epubPart, imgStorage epubzip.StorageImageReader) (err error) {
	hasTitlePage := e.TitlePage == 1 || (e.TitlePage == 2 && totalParts > 1)

	wz, err := epubzip.New(path)
	if err != nil {
		return err
	}
	defer func() {
		// the end of the zip is written on close, a failure there makes the epub invalid
		if cerr := wz.Close(); err == nil {
			err = cerr
		}
		// do not leave a partial epub behind
		if err != nil {
			_ = os.Remove(path)
		}
	}()

	title := e.Title
	titlePage := e.Title
	coverText := ""
	if totalParts > 1 {
		partLabel := utils.PartLabel(currentPart, totalParts)
		title = title + " - " + partLabel
		titlePage = titlePage + "\n" + partLabel
		if e.CoverCaption {
			coverText = partLabel
		}
	}

	type zipContent struct {
		Name    string
		Content string
	}
	opf := epubtemplates.Content{
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
	}
	content := []zipContent{
		{"META-INF/container.xml", epubtemplates.Container},
		{"META-INF/com.apple.ibooks.display-options.xml", epubtemplates.AppleBooks},
		{"OEBPS/content.opf", opf.String()},
		{"OEBPS/toc.xhtml", epubtemplates.Toc(title, hasTitlePage, e.StripFirstDirectoryFromToc, part.Images)},
		{"OEBPS/Text/style.css", e.render(styleTemplate, map[string]any{
			"Width":      e.Image.View.Width,
			"Height":     e.Image.View.Height,
			"Foreground": utils.ColorToHex(utils.StyleColor(e.Image.GrayScale, e.Image.GrayScaleMode, e.Image.View.Color.Foreground)),
			"Background": utils.ColorToHex(utils.StyleColor(e.Image.GrayScale, e.Image.GrayScaleMode, e.Image.View.Color.Background)),
		})},
	}

	if e.Image.AddPanelView {
		content = append(content, zipContent{"OEBPS/Text/panelview.css", epubtemplates.PanelViewStyle})
	}

	if err = wz.WriteMagic(); err != nil {
		return err
	}
	for _, c := range content {
		if err := wz.WriteContent(c.Name, []byte(c.Content)); err != nil {
			return err
		}
	}

	if err = e.writeCoverImage(wz, part.Cover, coverText); err != nil {
		return err
	}

	if hasTitlePage {
		if err = e.writeTitleImage(wz, title, titlePage); err != nil {
			return err
		}
	}

	for _, img := range part.Images {
		if err := e.writeImage(wz, img, imgStorage); err != nil {
			return err
		}
	}

	// the blank pages listed in the content: the spine refers to them
	for _, img := range opf.Spaces() {
		if err := e.writeBlank(wz, img); err != nil {
			return err
		}
	}
	return nil
}

// create the zip
func (e epub) Write() error {
	// the epub being written, removed on an interruption
	var (
		writingMu sync.Mutex
		writing   string
	)
	setWriting := func(path string) {
		writingMu.Lock()
		defer writingMu.Unlock()
		writing = path
	}

	// the temporary storage must not outlive the conversion, whatever happens
	if !e.Dry {
		defer func() {
			_ = os.Remove(e.ImgStorage())
		}()

		// an interruption exits without running the deferred calls: remove the files by hand
		sigs := make(chan os.Signal, 1)
		signal.Notify(sigs, os.Interrupt, syscall.SIGTERM)
		done := make(chan struct{})
		defer func() {
			signal.Stop(sigs)
			close(done)
		}()
		go func() {
			select {
			case <-done:
			case sig := <-sigs:
				_ = os.Remove(e.ImgStorage())
				writingMu.Lock()
				if writing != "" {
					_ = os.Remove(writing)
				}
				writingMu.Unlock()
				utils.Printf("\nInterrupted\n")
				if sig == os.Interrupt {
					os.Exit(130)
				}
				os.Exit(143)
			}
		}()
	}

	epubParts, imgStorage, err := e.getParts()
	if err != nil {
		return err
	}

	if e.Dry {
		p := epubParts[0]
		utils.Printf("TOC:\n  - %s\n%s\n", e.Title, e.getTree(p.Images, true))
		utils.Printf("Cover:\n%s\n", e.getTree([]epubimage.EPUBImage{p.Cover}, false))
		utils.Printf("Files:\n%s\n", e.getTree(p.Images, false))
		return nil
	}
	defer func() {
		_ = imgStorage.Close()
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
			suffix = " - " + utils.PartLabel(i+1, totalParts)
		}

		path := e.Output[0:len(e.Output)-len(ext)] + suffix + ext

		setWriting(path)
		if err := e.writePart(
			path,
			i+1,
			totalParts,
			part,
			imgStorage,
		); err != nil {
			return err
		}

		setWriting("")
		_ = bar.Add(1)
	}
	_ = bar.Close()
	if !e.Json {
		utils.Println()
	}

	// display corrupted images
	hasError := false
	for _, part := range epubParts {
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
