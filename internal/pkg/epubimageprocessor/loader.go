package epubimageprocessor

import (
	"archive/zip"
	"bytes"
	"errors"
	"fmt"
	"image"
	"image/draw"
	_ "image/jpeg"
	_ "image/png"
	"io"
	"io/fs"
	"os"
	"path/filepath"
	"sort"
	"strings"
	"sync"

	"github.com/gen2brain/go-fitz"
	_ "golang.org/x/image/tiff"
	_ "golang.org/x/image/webp"

	"github.com/celogeek/go-comic-converter/v3/internal/pkg/epubimagefilters"
	"github.com/celogeek/go-comic-converter/v3/internal/pkg/sortpath"
	"github.com/celogeek/go-comic-converter/v3/internal/pkg/utils"
	"github.com/celogeek/go-comic-converter/v3/pkg/epuboptions"
	"github.com/nwaples/rardecode/v2"
)

// size of the page shown in place of an unreadable image.
const (
	placeholderWidth  = 1200
	placeholderHeight = 1920
)

type task struct {
	Id    int
	Image image.Image
	Path  string
	Name  string
	// Error is the cause of a corrupted image, the page shown in its place is the image.
	Error error
	// Fatal is an error that stops the conversion, the other fields are not set.
	Fatal error
}

var errNoImagesFound = errors.New("no images found")

// only accept jpg, png and webp as source file
func (e ePUBImageProcessor) isSupportedImage(path string) bool {
	switch strings.ToLower(filepath.Ext(path)) {
	case ".jpg", ".jpeg", ".png", ".webp", ".tiff":
		{
			return !strings.HasPrefix(filepath.Base(path), ".")
		}
	}
	return false
}

// load images from input
func (e ePUBImageProcessor) load() (totalImages int, output chan task, err error) {
	fi, err := os.Stat(e.Input)
	if err != nil {
		return
	}

	// get all images though a channel of bytes
	if fi.IsDir() {
		return e.loadDir()
	} else {
		switch ext := strings.ToLower(filepath.Ext(e.Input)); ext {
		case ".cbz", ".zip":
			return e.loadCbz()
		case ".cbr", ".rar":
			return e.loadCbr()
		case ".pdf":
			return e.loadPdf()
		default:
			err = fmt.Errorf("unknown file format (%s): support .cbz, .zip, .cbr, .rar, .pdf", ext)
			return
		}
	}
}

// CorruptedImage page shown in place of an image that cannot be read, with the options of the view.
func CorruptedImage(o epuboptions.EPUBOptions, path, name string) image.Image {
	return ePUBImageProcessor{o}.corruptedImage(path, name)
}

// corruptedImage page shown in place of an image that cannot be read, with the title box of the cover,
// in the foreground and background colors of the view.
func (e ePUBImageProcessor) corruptedImage(path, name string) image.Image {
	txt := name
	if path != "" {
		txt += "\nin " + filepath.Clean(path)
	}
	txt += "\nis corrupted!"

	colors := e.Image.View.Color
	foreground := utils.StyleColor(e.Image.GrayScale, e.Image.GrayScaleMode, colors.Foreground)
	background := utils.StyleColor(e.Image.GrayScale, e.Image.GrayScaleMode, colors.Background)

	src := image.NewRGBA(image.Rect(0, 0, placeholderWidth, placeholderHeight))
	draw.Draw(src, src.Bounds(), image.NewUniform(background), image.Point{}, draw.Src)

	fontSize := epubimagefilters.TitleFontSize(src.Bounds().Dx())

	dst := image.NewRGBA(src.Bounds())
	epubimagefilters.CoverTitle{
		Title:       txt,
		Align:       "center",
		Font:        e.TrueTypeFont,
		FontSize:    fontSize,
		BorderWidth: epubimagefilters.TitleBorderWidth(fontSize),
		Foreground:  foreground,
		Background:  background,
	}.Draw(dst, src, nil)

	return dst
}

// load a directory of images
func (e ePUBImageProcessor) loadDir() (totalImages int, output chan task, err error) {
	images := make([]string, 0)

	input := filepath.Clean(e.Input)
	err = filepath.WalkDir(input, func(path string, d fs.DirEntry, err error) error {
		if err != nil {
			return err
		}
		if !d.IsDir() && e.isSupportedImage(path) {
			images = append(images, path)
		}
		return nil
	})

	if err != nil {
		return
	}

	totalImages = len(images)

	if totalImages == 0 {
		err = errNoImagesFound
		return
	}

	sort.Sort(sortpath.By(images, e.SortPathMode))

	// Queue all file with id
	type job struct {
		Id   int
		Path string
	}
	jobs := make(chan job)
	go func() {
		defer close(jobs)
		for i, path := range images {
			jobs <- job{i, path}
		}
	}()

	// read in parallel and get an image
	output = make(chan task, e.Workers)
	wg := &sync.WaitGroup{}
	for range e.WorkersRatio(50) {
		wg.Add(1)
		go func() {
			defer wg.Done()
			for job := range jobs {
				var img image.Image
				var err error
				if !e.Dry {
					var f *os.File
					f, err = os.Open(job.Path)
					if err == nil {
						img, _, err = image.Decode(f)
						_ = f.Close()
					}
				}

				p, fn := utils.SplitPath(input, job.Path)
				if err != nil {
					img = e.corruptedImage(p, fn)
				}
				output <- task{
					Id:    job.Id,
					Image: img,
					Path:  p,
					Name:  fn,
					Error: err,
				}
			}
		}()
	}

	// wait all done and close
	go func() {
		wg.Wait()
		close(output)
	}()

	return
}

// load a zip file that include images
func (e ePUBImageProcessor) loadCbz() (totalImages int, output chan task, err error) {
	r, err := zip.OpenReader(e.Input)
	if err != nil {
		return
	}

	images := make([]*zip.File, 0)
	for _, f := range r.File {
		if !f.FileInfo().IsDir() && e.isSupportedImage(f.Name) {
			images = append(images, f)
		}
	}

	totalImages = len(images)

	if totalImages == 0 {
		_ = r.Close()
		err = errNoImagesFound
		return
	}

	var names []string
	for _, img := range images {
		names = append(names, img.Name)
	}
	sort.Sort(sortpath.By(names, e.SortPathMode))

	indexedNames := make(map[string]int)
	for i, name := range names {
		indexedNames[name] = i
	}

	type job struct {
		Id int
		F  *zip.File
	}
	jobs := make(chan job)
	go func() {
		defer close(jobs)
		for _, img := range images {
			jobs <- job{indexedNames[img.Name], img}
		}
	}()

	output = make(chan task, e.Workers)
	wg := &sync.WaitGroup{}
	for range e.WorkersRatio(50) {
		wg.Add(1)
		go func() {
			defer wg.Done()
			for job := range jobs {
				var img image.Image
				var err error
				if !e.Dry {
					var f io.ReadCloser
					f, err = job.F.Open()
					if err == nil {
						img, _, err = image.Decode(f)
						_ = f.Close()
					}
				}

				p, fn := utils.SplitPath("", job.F.Name)
				if err != nil {
					img = e.corruptedImage(p, fn)
				}
				output <- task{
					Id:    job.Id,
					Image: img,
					Path:  p,
					Name:  fn,
					Error: err,
				}
			}
		}()
	}

	go func() {
		wg.Wait()
		close(output)
		_ = r.Close()
	}()
	return
}

// load a rar file that include images
func (e ePUBImageProcessor) loadCbr() (totalImages int, output chan task, err error) {
	var isSolid bool
	files, err := rardecode.List(e.Input)
	if err != nil {
		return
	}

	names := make([]string, 0)
	for _, f := range files {
		if !f.IsDir && e.isSupportedImage(f.Name) {
			if f.Solid {
				isSolid = true
			}
			names = append(names, f.Name)
		}
	}

	totalImages = len(names)
	if totalImages == 0 {
		err = errNoImagesFound
		return
	}

	sort.Sort(sortpath.By(names, e.SortPathMode))

	indexedNames := make(map[string]int)
	for i, name := range names {
		indexedNames[name] = i
	}

	type job struct {
		Id   int
		Name string
		Open func() (io.ReadCloser, error)
		Err  error
	}

	jobs := make(chan job)
	go func() {
		defer close(jobs)
		if isSolid && !e.Dry {
			r, rerr := rardecode.OpenReader(e.Input)
			if rerr != nil {
				jobs <- job{Err: fmt.Errorf("error processing image %s: %w", e.Input, rerr)}
				return
			}
			defer func(r *rardecode.ReadCloser) {
				_ = r.Close()
			}(r)
			for {
				f, rerr := r.Next()
				if rerr != nil {
					if errors.Is(rerr, io.EOF) {
						break
					}
					// f is nil on error
					jobs <- job{Err: fmt.Errorf("error processing image %s: %w", e.Input, rerr)}
					return
				}
				if i, ok := indexedNames[f.Name]; ok {
					var b bytes.Buffer
					_, rerr = io.Copy(&b, r)
					if rerr != nil {
						jobs <- job{Err: fmt.Errorf("error processing image %s: %w", f.Name, rerr)}
						return
					}
					jobs <- job{Id: i, Name: f.Name, Open: func() (io.ReadCloser, error) {
						return io.NopCloser(bytes.NewReader(b.Bytes())), nil
					}}
				}
			}
		} else {
			for _, img := range files {
				if i, ok := indexedNames[img.Name]; ok {
					jobs <- job{Id: i, Name: img.Name, Open: img.Open}
				}
			}
		}
	}()

	// send file to the queue
	output = make(chan task, e.Workers)
	wg := &sync.WaitGroup{}
	for range e.WorkersRatio(50) {
		wg.Add(1)
		go func() {
			defer wg.Done()
			for job := range jobs {
				var img image.Image
				var err error
				if !e.Dry {
					var f io.ReadCloser
					f, err = job.Open()
					if err == nil {
						img, _, err = image.Decode(f)
						_ = f.Close()
					}
				}

				p, fn := utils.SplitPath("", job.Name)
				if err != nil {
					img = e.corruptedImage(p, fn)
				}
				output <- task{
					Id:    job.Id,
					Image: img,
					Path:  p,
					Name:  fn,
					Error: err,
				}
			}
		}()
	}
	go func() {
		wg.Wait()
		close(output)
	}()
	return
}

// extract image from a pdf
func (e ePUBImageProcessor) loadPdf() (totalImages int, output chan task, err error) {
	pdf, err := fitz.New(e.Input)
	if err != nil {
		return
	}

	totalImages = pdf.NumPage()
	pageFmt := "page " + utils.FormatNumberOfDigits(totalImages)
	output = make(chan task)
	go func() {
		defer close(output)
		defer pdf.Close()
		for i := range totalImages {
			var img image.Image
			var err error
			if !e.Dry {
				img, err = pdf.Image(i)
			}

			name := fmt.Sprintf(pageFmt, i+1)
			if err != nil {
				img = e.corruptedImage("", name)
			}
			output <- task{
				Id:    i,
				Image: img,
				Path:  "",
				Name:  name,
				Error: err,
			}
		}
	}()

	return
}
