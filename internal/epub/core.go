package epub

import (
	"archive/zip"
	"fmt"
	"io/fs"
	"os"
	"path/filepath"
	"sort"
	"strings"
	"text/template"
	"time"

	"github.com/gofrs/uuid"
)

type Images struct {
	Id     int
	Path   string
	Name   string
	Title  string
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

	Images []Images
	Error  error
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

		Images: make([]Images, 0),
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
	m, err := wz.Create(file)
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
	return result.String()
}

func (e *EPub) LoadDir(dirname string) *EPub {
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
		name := filepath.Base(path)
		title := name[0 : len(name)-len(ext)]

		e.Images = append(e.Images, Images{
			Path:  path,
			Name:  name,
			Title: title,
		})
		return nil
	})
	if err != nil {
		e.Error = err
		return e
	}
	sort.SliceStable(e.Images, func(i, j int) bool {
		return strings.Compare(e.Images[i].Path, e.Images[j].Path) < 0
	})

	for i := range e.Images {
		e.Images[i].Id = i
	}

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
		{"META-INF/container.xml", TEMPLATE_CONTAINER},
		{"OEBPS/content.opf", e.Render(TEMPLATE_CONTENT, e)},
		{"OEBPS/toc.ncx", e.Render(TEMPLATE_TOC, e)},
		{"OEBPS/nav.xhtml", e.Render(TEMPLATE_NAV, e)},
		{"OEBPS/Text/style.css", TEMPLATE_STYLE},
	}
	for _, img := range e.Images {
		filename := fmt.Sprintf("OEBPS/Text/%d.xhtml", img.Id)
		zipContent = append(zipContent, []string{filename, e.Render(TEMPLATE_TEXT, img)})
	}

	wz := zip.NewWriter(w)
	defer wz.Close()
	for _, content := range zipContent {
		if err := e.WriteFile(wz, content[0], content[1]); err != nil {
			return err
		}
	}
	return nil
}
