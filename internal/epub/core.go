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

type ImageDetails struct {
	*Images
	Data   string
	Width  int
	Height int
}

type Images struct {
	Id    int
	Title string
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

	Images          []*Images
	FirstImageTitle string
	Error           error

	ProcessingImages func() chan *ImageDetails
}

func NewEpub(path string) *EPub {
	uid, err := uuid.NewV4()
	if err != nil {
		panic(err)
	}

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

func (e *EPub) WriteFile(wz *zip.Writer, file, content string) error {
	m, err := wz.CreateHeader(&zip.FileHeader{
		Name:     file,
		Modified: time.Now(),
	})
	if err != nil {
		return err
	}
	_, err = m.Write([]byte(content))
	return err
}

func (e *EPub) Render(templateString string, data any) string {
	tmpl := template.New("parser")
	tmpl.Funcs(template.FuncMap{"mod": func(i, j int) bool { return i%j == 0 }})
	tmpl.Funcs(template.FuncMap{"zoom": func(s int, z float32) int { return int(float32(s) * z) }})
	tmpl, err := tmpl.Parse(templateString)
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

	titleFormat := fmt.Sprintf("%%0%dd", len(fmt.Sprint(len(images)-1)))

	for i := range images {
		e.Images = append(e.Images, &Images{
			Id:    i,
			Title: fmt.Sprintf(titleFormat, i),
		})
	}

	type Todo struct {
		*Images
		Path string
	}

	todo := make(chan *Todo)

	e.ProcessingImages = func() chan *ImageDetails {
		wg := &sync.WaitGroup{}
		results := make(chan *ImageDetails)
		for i := 0; i < runtime.NumCPU(); i++ {
			wg.Add(1)
			go func() {
				defer wg.Done()
				for task := range todo {
					data, w, h := imageconverter.Convert(task.Path, true, e.ViewWidth, e.ViewHeight, e.Quality)
					results <- &ImageDetails{task.Images, data, w, h}
				}
			}()
		}
		go func() {
			for i, path := range images {
				todo <- &Todo{e.Images[i], path}
			}
			close(todo)
			wg.Wait()
			close(results)
		}()
		return results
	}

	e.FirstImageTitle = e.Images[0].Title

	return e
}

func (e *EPub) Write() error {
	if e.Error != nil {
		return e.Error
	}

	w, err := os.Create(e.Path)
	if err != nil {
		return err
	}

	zipContent := [][]string{
		{"mimetype", TEMPLATE_MIME_TYPE},
		{"META-INF/container.xml", gohtml.Format(TEMPLATE_CONTAINER)},
		{"OEBPS/content.opf", e.Render(TEMPLATE_CONTENT, e)},
		{"OEBPS/toc.ncx", e.Render(TEMPLATE_TOC, e)},
		{"OEBPS/nav.xhtml", e.Render(TEMPLATE_NAV, e)},
		{"OEBPS/Text/style.css", TEMPLATE_STYLE},
	}

	wz := zip.NewWriter(w)
	defer wz.Close()
	for _, content := range zipContent {
		if err := e.WriteFile(wz, content[0], content[1]); err != nil {
			return err
		}
	}

	bar := progressbar.Default(int64(len(e.Images)), "Processing")
	defer bar.Close()
	for img := range e.ProcessingImages() {
		text := fmt.Sprintf("OEBPS/Text/%s.xhtml", img.Title)
		image := fmt.Sprintf("OEBPS/Images/%s.jpg", img.Title)
		if err := e.WriteFile(wz, text, e.Render(TEMPLATE_TEXT, img)); err != nil {
			return err
		}
		if err := e.WriteFile(wz, image, img.Data); err != nil {
			return err
		}
		bar.Add(1)
	}

	return nil
}
