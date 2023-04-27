package epubimageprocessing

import (
	"archive/zip"
	"bytes"
	"errors"
	"fmt"
	"image"
	_ "image/jpeg"
	_ "image/png"
	"io"
	"io/fs"
	"os"
	"path/filepath"
	"sort"
	"strings"
	"sync"

	_ "golang.org/x/image/webp"

	epuboptions "github.com/celogeek/go-comic-converter/v2/internal/epub/options"
	"github.com/celogeek/go-comic-converter/v2/internal/sortpath"
	"github.com/nwaples/rardecode/v2"
	pdfimage "github.com/raff/pdfreader/image"
	"github.com/raff/pdfreader/pdfread"
)

type tasks struct {
	Id    int
	Image image.Image
	Path  string
	Name  string
}

var errNoImagesFound = errors.New("no images found")

// Load images from input
func Load(o *epuboptions.Options) (totalImages int, output chan *tasks, err error) {
	fi, err := os.Stat(o.Input)
	if err != nil {
		return
	}

	// get all images though a channel of bytes
	if fi.IsDir() {
		return loadDir(o)
	} else {
		switch ext := strings.ToLower(filepath.Ext(o.Input)); ext {
		case ".cbz", ".zip":
			return loadCbz(o)
		case ".cbr", ".rar":
			return loadCbr(o)
		case ".pdf":
			return loadPdf(o)
		default:
			err = fmt.Errorf("unknown file format (%s): support .cbz, .zip, .cbr, .rar, .pdf", ext)
			return
		}
	}
}

// load a directory of images
func loadDir(o *epuboptions.Options) (totalImages int, output chan *tasks, err error) {
	images := make([]string, 0)

	input := filepath.Clean(o.Input)
	err = filepath.WalkDir(input, func(path string, d fs.DirEntry, err error) error {
		if err != nil {
			return err
		}
		if !d.IsDir() && isSupportedImage(path) {
			images = append(images, path)
		}
		return nil
	})

	if err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}

	totalImages = len(images)

	if totalImages == 0 {
		err = errNoImagesFound
		return
	}

	sort.Sort(sortpath.By(images, o.SortPathMode))

	// Queue all file with id
	type job struct {
		Id   int
		Path string
	}
	jobs := make(chan *job)
	go func() {
		defer close(jobs)
		for i, path := range images {
			jobs <- &job{i, path}
		}
	}()

	// read in parallel and get an image
	output = make(chan *tasks, o.Workers)
	wg := &sync.WaitGroup{}
	for j := 0; j < o.WorkersRatio(50); j++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			for job := range jobs {
				var img image.Image
				if !o.Dry {
					f, err := os.Open(job.Path)
					if err != nil {
						fmt.Fprintf(os.Stderr, "\nerror processing image %s: %s\n", job.Path, err)
						os.Exit(1)
					}
					img, _, err = image.Decode(f)
					if err != nil {
						fmt.Fprintf(os.Stderr, "\nerror processing image %s: %s\n", job.Path, err)
						os.Exit(1)
					}
					f.Close()
				}

				p, fn := filepath.Split(job.Path)
				if p == input {
					p = ""
				} else {
					p = p[len(input)+1:]
				}
				output <- &tasks{
					Id:    job.Id,
					Image: img,
					Path:  p,
					Name:  fn,
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
func loadCbz(o *epuboptions.Options) (totalImages int, output chan *tasks, err error) {
	r, err := zip.OpenReader(o.Input)
	if err != nil {
		return
	}

	images := make([]*zip.File, 0)
	for _, f := range r.File {
		if !f.FileInfo().IsDir() && isSupportedImage(f.Name) {
			images = append(images, f)
		}
	}

	totalImages = len(images)

	if totalImages == 0 {
		r.Close()
		err = errNoImagesFound
		return
	}

	names := []string{}
	for _, img := range images {
		names = append(names, img.Name)
	}
	sort.Sort(sortpath.By(names, o.SortPathMode))

	indexedNames := make(map[string]int)
	for i, name := range names {
		indexedNames[name] = i
	}

	type job struct {
		Id int
		F  *zip.File
	}
	jobs := make(chan *job)
	go func() {
		defer close(jobs)
		for _, img := range images {
			jobs <- &job{indexedNames[img.Name], img}
		}
	}()

	output = make(chan *tasks, o.Workers)
	wg := &sync.WaitGroup{}
	for j := 0; j < o.WorkersRatio(50); j++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			for job := range jobs {
				var img image.Image
				if !o.Dry {
					f, err := job.F.Open()
					if err != nil {
						fmt.Fprintf(os.Stderr, "\nerror processing image %s: %s\n", job.F.Name, err)
						os.Exit(1)
					}
					img, _, err = image.Decode(f)
					if err != nil {
						fmt.Fprintf(os.Stderr, "\nerror processing image %s: %s\n", job.F.Name, err)
						os.Exit(1)
					}
					f.Close()
				}

				p, fn := filepath.Split(filepath.Clean(job.F.Name))
				output <- &tasks{
					Id:    job.Id,
					Image: img,
					Path:  p,
					Name:  fn,
				}
			}
		}()
	}

	go func() {
		wg.Wait()
		close(output)
		r.Close()
	}()
	return
}

// load a rar file that include images
func loadCbr(o *epuboptions.Options) (totalImages int, output chan *tasks, err error) {
	var isSolid bool
	files, err := rardecode.List(o.Input)
	if err != nil {
		return
	}

	names := make([]string, 0)
	for _, f := range files {
		if !f.IsDir && isSupportedImage(f.Name) {
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

	sort.Sort(sortpath.By(names, o.SortPathMode))

	indexedNames := make(map[string]int)
	for i, name := range names {
		indexedNames[name] = i
	}

	type job struct {
		Id   int
		Name string
		Open func() (io.ReadCloser, error)
	}

	jobs := make(chan *job)
	go func() {
		defer close(jobs)
		if isSolid && !o.Dry {
			r, rerr := rardecode.OpenReader(o.Input)
			if rerr != nil {
				fmt.Fprintf(os.Stderr, "\nerror processing image %s: %s\n", o.Input, rerr)
				os.Exit(1)
			}
			defer r.Close()
			for {
				f, rerr := r.Next()
				if rerr != nil {
					if rerr == io.EOF {
						break
					}
					fmt.Fprintf(os.Stderr, "\nerror processing image %s: %s\n", f.Name, rerr)
					os.Exit(1)
				}
				if i, ok := indexedNames[f.Name]; ok {
					var b bytes.Buffer
					_, rerr = io.Copy(&b, r)
					if rerr != nil {
						fmt.Fprintf(os.Stderr, "\nerror processing image %s: %s\n", f.Name, rerr)
						os.Exit(1)
					}
					jobs <- &job{i, f.Name, func() (io.ReadCloser, error) {
						return io.NopCloser(bytes.NewReader(b.Bytes())), nil
					}}
				}
			}
		} else {
			for _, img := range files {
				if i, ok := indexedNames[img.Name]; ok {
					jobs <- &job{i, img.Name, img.Open}
				}
			}
		}
	}()

	// send file to the queue
	output = make(chan *tasks, o.Workers)
	wg := &sync.WaitGroup{}
	for j := 0; j < o.WorkersRatio(50); j++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			for job := range jobs {
				var img image.Image
				if !o.Dry {
					f, err := job.Open()
					if err != nil {
						fmt.Fprintf(os.Stderr, "\nerror processing image %s: %s\n", job.Name, err)
						os.Exit(1)
					}
					img, _, err = image.Decode(f)
					if err != nil {
						fmt.Fprintf(os.Stderr, "\nerror processing image %s: %s\n", job.Name, err)
						os.Exit(1)
					}
					f.Close()
				}

				p, fn := filepath.Split(filepath.Clean(job.Name))
				output <- &tasks{
					Id:    job.Id,
					Image: img,
					Path:  p,
					Name:  fn,
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
func loadPdf(o *epuboptions.Options) (totalImages int, output chan *tasks, err error) {
	pdf := pdfread.Load(o.Input)
	if pdf == nil {
		err = fmt.Errorf("can't read pdf")
		return
	}

	totalImages = len(pdf.Pages())
	pageFmt := fmt.Sprintf("page %%0%dd", len(fmt.Sprintf("%d", totalImages)))
	output = make(chan *tasks)
	go func() {
		defer close(output)
		defer pdf.Close()
		for i := 0; i < totalImages; i++ {
			var img image.Image
			if !o.Dry {
				img, err = pdfimage.Extract(pdf, i+1)
				if err != nil {
					fmt.Fprintln(os.Stderr, err)
					os.Exit(1)
				}
			}

			output <- &tasks{
				Id:    i,
				Image: img,
				Path:  "",
				Name:  fmt.Sprintf(pageFmt, i+1),
			}
		}
	}()

	return
}
