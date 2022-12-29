package epub

import (
	"archive/zip"
	"fmt"
	"io/fs"
	"os"
	"path/filepath"
	"runtime"
	"sort"
	"strings"
	"sync"
	"text/template"
	"time"

	"github.com/gofrs/uuid"
	"github.com/schollz/progressbar/v3"

	imageconverter "go-comic-converter/internal/image-converter"
)

type image struct {
	Id     int
	Data   *imageData
	Width  int
	Height int
}

type EpubOptions struct {
	Input      string
	Output     string
	Title      string
	Author     string
	ViewWidth  int
	ViewHeight int
	Quality    int
	Crop       bool
	LimitMb    int
}

type ePub struct {
	*EpubOptions
	UID       string
	Publisher string
	UpdatedAt string

	imagesCount       int
	processingImages  func() chan *image
	templateProcessor *template.Template
}

type epubPart struct {
	Cover  *image
	Images []*image
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

	return result.String()
}

func (e *ePub) load() error {
	fi, err := os.Stat(e.Input)
	if err != nil {
		return err
	}

	if fi.IsDir() {
		return e.loadDir()
	}

	switch ext := strings.ToLower(filepath.Ext(e.Input)); ext {
	case ".cbz":
		return e.loadCBZ()
	case ".cbr":
		return e.loadCBR()
	case ".pdf":
		return e.loadPDF()
	default:
		return fmt.Errorf("unknown file format (%s): support .cbz, .cbr, .pdf", ext)
	}
}

func (e *ePub) loadCBZ() error {
	r, err := zip.OpenReader(e.Input)
	if err != nil {
		return err
	}

	images := make([]*zip.File, 0)
	for _, f := range r.File {
		if f.FileInfo().IsDir() {
			continue
		}
		if strings.ToLower(filepath.Ext(f.Name)) != ".jpg" {
			continue
		}
		images = append(images, f)
	}
	if len(images) == 0 {
		r.Close()
		return fmt.Errorf("no images found")
	}

	sort.SliceStable(images, func(i, j int) bool {
		return strings.Compare(images[i].Name, images[j].Name) < 0
	})

	e.imagesCount = len(images)

	type Todo struct {
		Id int
		FZ *zip.File
	}

	todo := make(chan *Todo)

	e.processingImages = func() chan *image {
		// defer r.Close()
		wg := &sync.WaitGroup{}
		results := make(chan *image)
		for i := 0; i < runtime.NumCPU(); i++ {
			wg.Add(1)
			go func() {
				defer wg.Done()
				for task := range todo {
					reader, err := task.FZ.Open()
					if err != nil {
						panic(err)
					}
					data, w, h := imageconverter.Convert(
						reader,
						e.Crop,
						e.ViewWidth,
						e.ViewHeight,
						e.Quality,
					)
					name := fmt.Sprintf("OEBPS/Images/%d.jpg", task.Id)
					if task.Id == 0 {
						name = "OEBPS/Images/cover.jpg"
					}
					results <- &image{
						task.Id,
						newImageData(name, data),
						w,
						h,
					}
				}
			}()
		}
		go func() {
			for i, fz := range images {
				todo <- &Todo{i, fz}
			}
			close(todo)
			wg.Wait()
			r.Close()
			close(results)
		}()

		return results
	}

	return nil
}

func (e *ePub) loadCBR() error {
	return fmt.Errorf("no implemented")
}

func (e *ePub) loadPDF() error {
	return fmt.Errorf("no implemented")
}

func (e *ePub) loadDir() error {
	images := make([]string, 0)
	err := filepath.WalkDir(e.Input, func(path string, d fs.DirEntry, err error) error {
		if err != nil {
			return err
		}
		if d.IsDir() {
			return nil
		}
		ext := filepath.Ext(path)
		if strings.ToLower(ext) != ".jpg" {
			return nil
		}

		images = append(images, path)
		return nil
	})
	if err != nil {
		return err
	}
	if len(images) == 0 {
		return fmt.Errorf("no images found")
	}
	sort.Strings(images)

	e.imagesCount = len(images)

	type Todo struct {
		Id   int
		Path string
	}

	todo := make(chan *Todo)

	e.processingImages = func() chan *image {
		wg := &sync.WaitGroup{}
		results := make(chan *image)
		for i := 0; i < runtime.NumCPU(); i++ {
			wg.Add(1)
			go func() {
				defer wg.Done()
				for task := range todo {
					reader, err := os.Open(task.Path)
					if err != nil {
						panic(err)
					}
					data, w, h := imageconverter.Convert(
						reader,
						e.Crop,
						e.ViewWidth,
						e.ViewHeight,
						e.Quality,
					)
					name := fmt.Sprintf("OEBPS/Images/%d.jpg", task.Id)
					if task.Id == 0 {
						name = "OEBPS/Images/cover.jpg"
					}
					results <- &image{
						task.Id,
						newImageData(name, data),
						w,
						h,
					}
				}
			}()
		}
		go func() {
			for i, path := range images {
				todo <- &Todo{i, path}
			}
			close(todo)
			wg.Wait()
			close(results)
		}()
		return results
	}

	return nil
}

func (e *ePub) getParts() []*epubPart {
	images := make([]*image, e.imagesCount)
	bar := progressbar.Default(int64(e.imagesCount), "Processing")
	for img := range e.processingImages() {
		images[img.Id] = img
		bar.Add(1)
	}
	bar.Close()

	parts := make([]*epubPart, 0)
	cover := images[0]
	images = images[1:]
	if e.LimitMb == 0 {
		parts = append(parts, &epubPart{
			Cover:  cover,
			Images: images,
		})
		return parts
	}

	maxSize := uint64(e.LimitMb * 1024 * 1024)

	xhtmlSize := uint64(1024)
	// descriptor files + image
	baseSize := uint64(16*1024) + cover.Data.CompressedSize()

	currentSize := baseSize
	currentImages := make([]*image, 0)
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
			currentImages = make([]*image, 0)
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

	return parts
}

func (e *ePub) Write() error {
	if err := e.load(); err != nil {
		return err
	}

	type ZipContent struct {
		Name    string
		Content any
	}

	epubParts := e.getParts()
	totalParts := len(epubParts)

	bar := progressbar.Default(int64(totalParts), "Writing Part")
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

		zipContent := []ZipContent{
			{"META-INF/container.xml", containerTmpl},
			{"OEBPS/content.opf", e.render(contentTmpl, map[string]any{"Info": e, "Images": part.Images})},
			{"OEBPS/toc.ncx", e.render(tocTmpl, map[string]any{"Info": e, "Image": part.Images[0]})},
			{"OEBPS/nav.xhtml", e.render(navTmpl, map[string]any{"Info": e, "Image": part.Images[0]})},
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
		for _, content := range zipContent {
			if err := wz.WriteFile(content.Name, content.Content); err != nil {
				return err
			}
		}
		wz.WriteImage(part.Cover.Data)

		for _, img := range part.Images {
			text := fmt.Sprintf("OEBPS/Text/%d.xhtml", img.Id)
			if err := wz.WriteFile(text, e.render(textTmpl, img)); err != nil {
				return err
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
