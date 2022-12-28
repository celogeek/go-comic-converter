package epub

import (
	"fmt"
	"io/fs"
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

type Image struct {
	Id     int
	Data   *ImageData
	Width  int
	Height int
}

type EPub struct {
	Path string

	UID        string
	Title      string
	Author     string
	Publisher  string
	UpdatedAt  string
	ViewWidth  int
	ViewHeight int
	Quality    int
	Crop       bool
	LimitMb    int

	Error error

	ImagesCount       int
	ProcessingImages  func() chan *Image
	TemplateProcessor *template.Template
}

type EpubPart struct {
	Cover  *Image
	Images []*Image
}

func NewEpub(path string) *EPub {
	uid, err := uuid.NewV4()
	if err != nil {
		panic(err)
	}

	tmpl := template.New("parser")
	tmpl.Funcs(template.FuncMap{
		"mod":  func(i, j int) bool { return i%j == 0 },
		"zoom": func(s int, z float32) int { return int(float32(s) * z) },
	})

	return &EPub{
		Path: path,

		UID:        uid.String(),
		Title:      "Unknown title",
		Author:     "Unknown author",
		Publisher:  "GO Comic Converter",
		UpdatedAt:  time.Now().UTC().Format("2006-01-02T15:04:05Z"),
		ViewWidth:  0,
		ViewHeight: 0,
		Quality:    75,

		TemplateProcessor: tmpl,
	}
}

func (e *EPub) SetTitle(title string) *EPub {
	e.Title = title
	return e
}

func (e *EPub) SetAuthor(author string) *EPub {
	e.Author = author
	return e
}

func (e *EPub) SetSize(w, h int) *EPub {
	e.ViewWidth = w
	e.ViewHeight = h
	return e
}

func (e *EPub) SetQuality(q int) *EPub {
	e.Quality = q
	return e
}

func (e *EPub) SetCrop(c bool) *EPub {
	e.Crop = c
	return e
}

func (e *EPub) SetLimitMb(l int) *EPub {
	e.LimitMb = l
	return e
}

func (e *EPub) Render(templateString string, data any) string {
	tmpl, err := e.TemplateProcessor.Parse(templateString)
	if err != nil {
		panic(err)
	}
	result := &strings.Builder{}
	if err := tmpl.Execute(result, data); err != nil {
		panic(err)
	}

	return result.String()
}

func (e *EPub) LoadDir(dirname string) *EPub {
	images := make([]string, 0)
	err := filepath.WalkDir(dirname, func(path string, d fs.DirEntry, err error) error {
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
		e.Error = err
		return e
	}
	if len(images) == 0 {
		e.Error = fmt.Errorf("no images found")
		return e
	}
	sort.Strings(images)

	e.ImagesCount = len(images)

	type Todo struct {
		Id   int
		Path string
	}

	todo := make(chan *Todo)

	e.ProcessingImages = func() chan *Image {
		wg := &sync.WaitGroup{}
		results := make(chan *Image)
		for i := 0; i < runtime.NumCPU(); i++ {
			wg.Add(1)
			go func() {
				defer wg.Done()
				for task := range todo {
					data, w, h := imageconverter.Convert(
						task.Path,
						e.Crop,
						e.ViewWidth,
						e.ViewHeight,
						e.Quality,
					)
					name := fmt.Sprintf("OEBPS/Images/%d.jpg", task.Id)
					if task.Id == 0 {
						name = "OEBPS/Images/cover.jpg"
					}
					results <- &Image{
						task.Id,
						NewImageData(name, data),
						w,
						h,
					}
				}
			}()
		}
		go func() {
			for i, path := range images {
				if i == 0 {
					todo <- &Todo{i, path}
				} else {
					todo <- &Todo{i, path}
				}
			}
			close(todo)
			wg.Wait()
			close(results)
		}()
		return results
	}

	return e
}

func (e *EPub) GetParts() []*EpubPart {
	images := make([]*Image, e.ImagesCount)
	bar := progressbar.Default(int64(e.ImagesCount), "Processing")
	for img := range e.ProcessingImages() {
		images[img.Id] = img
		bar.Add(1)
	}
	bar.Close()

	epubPart := make([]*EpubPart, 0)

	cover := images[0]
	images = images[1:]
	if e.LimitMb == 0 {
		epubPart = append(epubPart, &EpubPart{
			Cover:  cover,
			Images: images,
		})
		return epubPart
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
			epubPart = append(epubPart, &EpubPart{
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
		epubPart = append(epubPart, &EpubPart{
			Cover:  cover,
			Images: currentImages,
		})
	}

	return epubPart
}

func (e *EPub) Write() error {
	if e.Error != nil {
		return e.Error
	}

	type ZipContent struct {
		Name    string
		Content any
	}

	epubParts := e.GetParts()
	totalParts := len(epubParts)

	bar := progressbar.Default(int64(totalParts), "Writing Part")
	for i, part := range epubParts {
		ext := filepath.Ext(e.Path)
		suffix := ""
		if totalParts > 1 {
			suffix = fmt.Sprintf(" PART_%02d", i+1)
		}

		path := fmt.Sprintf("%s%s%s", e.Path[0:len(e.Path)-len(ext)], suffix, ext)
		wz, err := NewEpubZip(path)
		if err != nil {
			return err
		}
		defer wz.Close()

		zipContent := []ZipContent{
			{"META-INF/container.xml", TEMPLATE_CONTAINER},
			{"OEBPS/content.opf", e.Render(TEMPLATE_CONTENT, map[string]any{"Info": e, "Images": part.Images})},
			{"OEBPS/toc.ncx", e.Render(TEMPLATE_TOC, map[string]any{"Info": e, "Image": part.Images[0]})},
			{"OEBPS/nav.xhtml", e.Render(TEMPLATE_NAV, map[string]any{"Info": e, "Image": part.Images[0]})},
			{"OEBPS/Text/style.css", TEMPLATE_STYLE},
			{"OEBPS/Text/part.xhtml", e.Render(TEMPLATE_PART, map[string]any{
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
			if err := wz.WriteFile(text, e.Render(TEMPLATE_TEXT, img)); err != nil {
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
