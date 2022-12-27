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
	"github.com/yosssi/gohtml"

	imageconverter "go-comic-converter/internal/image-converter"
)

type Image struct {
	Id     int
	Data   []byte
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

func (e *EPub) WriteFile(wz *zip.Writer, file string, data any) error {
	var content []byte
	switch b := data.(type) {
	case string:
		content = []byte(b)
	case []byte:
		content = b
	default:
		return fmt.Errorf("support string of []byte")
	}

	m, err := wz.CreateHeader(&zip.FileHeader{
		Name:     file,
		Modified: time.Now(),
	})
	if err != nil {
		return err
	}
	_, err = m.Write(content)
	return err
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

	return gohtml.Format(result.String())
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
					data, w, h := imageconverter.Convert(task.Path, e.Crop, e.ViewWidth, e.ViewHeight, e.Quality)
					results <- &Image{
						task.Id,
						data,
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
	totalSize := 0
	bar := progressbar.Default(int64(e.ImagesCount), "Processing")
	for img := range e.ProcessingImages() {
		images[img.Id] = img
		totalSize += len(img.Data)
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

	maxSize := e.LimitMb * 1024 * 1024
	currentSize := 512*1024 + len(cover.Data)
	currentImages := make([]*Image, 0)
	part := 1

	for _, img := range images {
		if len(currentImages) > 0 && currentSize+len(img.Data) > maxSize {
			epubPart = append(epubPart, &EpubPart{
				Cover:  cover,
				Images: currentImages,
			})
			part += 1
			currentSize = 512*1024 + len(cover.Data)
			currentImages = make([]*Image, 0)
		}
		currentSize += len(img.Data)
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

	for i, part := range epubParts {
		fmt.Printf("Writing part %d...\n", i+1)
		ext := filepath.Ext(e.Path)
		suffix := ""
		if totalParts > 1 {
			suffix = fmt.Sprintf(" PART_%02d", i+1)
		}
		path := fmt.Sprintf("%s%s%s", e.Path[0:len(e.Path)-len(ext)], suffix, ext)
		w, err := os.Create(path)
		if err != nil {
			return err
		}

		zipContent := []ZipContent{
			{"mimetype", TEMPLATE_MIME_TYPE},
			{"META-INF/container.xml", gohtml.Format(TEMPLATE_CONTAINER)},
			{"OEBPS/content.opf", e.Render(TEMPLATE_CONTENT, map[string]any{"Info": e, "Images": part.Images})},
			{"OEBPS/toc.ncx", e.Render(TEMPLATE_TOC, map[string]any{"Info": e, "Images": part.Images})},
			{"OEBPS/nav.xhtml", e.Render(TEMPLATE_NAV, map[string]any{"Info": e, "Images": part.Images})},
			{"OEBPS/Text/style.css", TEMPLATE_STYLE},
			{"OEBPS/Text/part.xhtml", e.Render(TEMPLATE_PART, map[string]any{
				"Info":  e,
				"Part":  i + 1,
				"Total": totalParts,
			})},
			{"OEBPS/Text/cover.xhtml", e.Render(TEMPLATE_TEXT, map[string]any{
				"Id":     "cover",
				"Width":  part.Cover.Width,
				"Height": part.Cover.Height,
			})},
			{"OEBPS/Images/cover.jpg", part.Cover.Data},
		}

		wz := zip.NewWriter(w)
		for _, content := range zipContent {
			if err := e.WriteFile(wz, content.Name, content.Content); err != nil {
				return err
			}
		}

		for _, img := range part.Images {
			text := fmt.Sprintf("OEBPS/Text/%d.xhtml", img.Id)
			image := fmt.Sprintf("OEBPS/Images/%d.jpg", img.Id)
			if err := e.WriteFile(wz, text, e.Render(TEMPLATE_TEXT, img)); err != nil {
				return err
			}
			if err := e.WriteFile(wz, image, img.Data); err != nil {
				return err
			}
		}
		wz.Close()
	}

	return nil
}
