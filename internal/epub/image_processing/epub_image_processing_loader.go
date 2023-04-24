package epubimageprocessing

import (
	"archive/zip"
	"bytes"
	"fmt"
	"io"
	"io/fs"
	"os"
	"path/filepath"
	"sort"

	epubimage "github.com/celogeek/go-comic-converter/v2/internal/epub/image"
	"github.com/celogeek/go-comic-converter/v2/internal/sortpath"
	"github.com/nwaples/rardecode"
	pdfimage "github.com/raff/pdfreader/image"
	"github.com/raff/pdfreader/pdfread"
	"golang.org/x/image/tiff"
)

type Options struct {
	Input        string
	SortPathMode int
	Quiet        bool
	Dry          bool
	Workers      int
	Image        *epubimage.Options
}

func (o *Options) mustExtractImage(imageOpener func() (io.ReadCloser, error)) *bytes.Buffer {
	if o.Dry {
		return &bytes.Buffer{}
	}
	f, err := imageOpener()
	if err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
	defer f.Close()
	var b bytes.Buffer
	_, err = io.Copy(&b, f)
	if err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
	return &b
}

func (o *Options) loadDir() (totalImages int, output chan *tasks, err error) {
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
		err = fmt.Errorf("image not found")
		return
	}

	sort.Sort(sortpath.By(images, o.SortPathMode))

	output = make(chan *tasks, o.Workers*2)
	go func() {
		defer close(output)
		for i, img := range images {
			p, fn := filepath.Split(img)
			if p == input {
				p = ""
			} else {
				p = p[len(input)+1:]
			}
			output <- &tasks{
				Id:     i,
				Reader: o.mustExtractImage(func() (io.ReadCloser, error) { return os.Open(img) }),
				Path:   p,
				Name:   fn,
			}
		}
	}()

	return
}

func (o *Options) loadCbz() (totalImages int, output chan *tasks, err error) {
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
		err = fmt.Errorf("no images found")
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

	output = make(chan *tasks, o.Workers*2)
	go func() {
		defer close(output)
		defer r.Close()
		for _, img := range images {
			p, fn := filepath.Split(filepath.Clean(img.Name))
			output <- &tasks{
				Id:     indexedNames[img.Name],
				Reader: o.mustExtractImage(img.Open),
				Path:   p,
				Name:   fn,
			}
		}
	}()
	return
}

func (o *Options) loadCbr() (totalImages int, output chan *tasks, err error) {
	// listing and indexing
	rl, err := rardecode.OpenReader(o.Input, "")
	if err != nil {
		return
	}

	names := make([]string, 0)
	for {
		f, ferr := rl.Next()

		if ferr != nil && ferr != io.EOF {
			rl.Close()
			err = ferr
			return
		}

		if f == nil {
			break
		}

		if !f.IsDir && isSupportedImage(f.Name) {
			names = append(names, f.Name)
		}
	}
	rl.Close()

	totalImages = len(names)
	if totalImages == 0 {
		err = fmt.Errorf("no images found")
		return
	}

	sort.Sort(sortpath.By(names, o.SortPathMode))

	indexedNames := make(map[string]int)
	for i, name := range names {
		indexedNames[name] = i
	}

	// send file to the queue
	output = make(chan *tasks, o.Workers*2)
	go func() {
		defer close(output)
		r, err := rardecode.OpenReader(o.Input, "")
		if err != nil {
			fmt.Fprintln(os.Stderr, err)
			os.Exit(1)

		}
		defer r.Close()

		for {
			f, err := r.Next()
			if err != nil && err != io.EOF {
				fmt.Fprintln(os.Stderr, err)
				os.Exit(1)

			}
			if f == nil {
				break
			}
			if idx, ok := indexedNames[f.Name]; ok {
				var b bytes.Buffer
				if !o.Dry {
					io.Copy(&b, r)
				}

				p, fn := filepath.Split(filepath.Clean(f.Name))

				output <- &tasks{
					Id:     idx,
					Reader: &b,
					Path:   p,
					Name:   fn,
				}
			}
		}
	}()

	return
}

func (o *Options) loadPdf() (totalImages int, output chan *tasks, err error) {
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
			var b bytes.Buffer

			if !o.Dry {
				img, err := pdfimage.Extract(pdf, i+1)
				if err != nil {
					fmt.Fprintln(os.Stderr, err)
					os.Exit(1)
				}

				err = tiff.Encode(&b, img, nil)
				if err != nil {
					fmt.Fprintln(os.Stderr, err)
					os.Exit(1)
				}
			}

			output <- &tasks{
				Id:     i,
				Reader: &b,
				Path:   "",
				Name:   fmt.Sprintf(pageFmt, i+1),
			}
		}
	}()

	return
}
