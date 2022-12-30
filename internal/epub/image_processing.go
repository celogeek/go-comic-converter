package epub

import (
	"archive/zip"
	"bytes"
	"fmt"
	imageconverter "go-comic-converter/internal/image-converter"
	"io"
	"io/fs"
	"os"
	"path/filepath"
	"runtime"
	"sort"
	"strings"
	"sync"

	"github.com/nwaples/rardecode"
	"github.com/schollz/progressbar/v3"
)

type Image struct {
	Id     int
	Data   *ImageData
	Width  int
	Height int
}

type imageTask struct {
	Id     int
	Reader io.ReadCloser
}

type readFakeCloser struct {
	io.Reader
}

func (rfc readFakeCloser) Close() error { return nil }

func LoadImages(path string, options *ImageOptions) ([]*Image, error) {
	images := make([]*Image, 0)

	fi, err := os.Stat(path)
	if err != nil {
		return nil, err
	}

	var (
		imageCount int
		imageInput chan *imageTask
	)

	if fi.IsDir() {
		imageCount, imageInput, err = loadDir(path)
	} else {
		switch ext := strings.ToLower(filepath.Ext(path)); ext {
		case ".cbz", "zip":
			imageCount, imageInput, err = loadCbz(path)
		case ".cbr", "rar":
			imageCount, imageInput, err = loadCbr(path)
		case ".pdf":
			err = fmt.Errorf("not implemented")
		default:
			err = fmt.Errorf("unknown file format (%s): support .cbz, .cbr, .pdf", ext)
		}
	}
	if err != nil {
		return nil, err
	}

	imageOutput := make(chan *Image)

	// processing
	wg := &sync.WaitGroup{}
	bar := progressbar.Default(int64(imageCount), "Processing")
	for i := 0; i < runtime.NumCPU(); i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			for img := range imageInput {
				data, w, h := imageconverter.Convert(
					img.Reader,
					options.Crop,
					options.ViewWidth,
					options.ViewHeight,
					options.Quality,
				)
				name := fmt.Sprintf("OEBPS/Images/%d.jpg", img.Id)
				if img.Id == 0 {
					name = "OEBPS/Images/cover.jpg"
				}
				imageOutput <- &Image{
					img.Id,
					newImageData(name, data),
					w,
					h,
				}
			}
		}()
	}

	go func() {
		wg.Wait()
		bar.Close()
		close(imageOutput)
	}()

	for image := range imageOutput {
		images = append(images, image)
		bar.Add(1)
	}

	if len(images) == 0 {
		return nil, fmt.Errorf("image not found")
	}

	sort.Slice(images, func(i, j int) bool {
		return images[i].Id < images[j].Id
	})

	return images, nil
}

func loadDir(input string) (int, chan *imageTask, error) {

	images := make([]string, 0)
	err := filepath.WalkDir(input, func(path string, d fs.DirEntry, err error) error {
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
		fmt.Println(err)
		os.Exit(1)
	}
	if len(images) == 0 {
		return 0, nil, fmt.Errorf("image not found")
	}

	sort.Strings(images)

	output := make(chan *imageTask)
	go func() {
		defer close(output)
		for i, img := range images {
			f, err := os.Open(img)
			if err != nil {
				fmt.Println(err)
				os.Exit(1)
			}
			output <- &imageTask{
				Id:     i,
				Reader: f,
			}
		}
	}()
	return len(images), output, nil
}

func loadCbz(input string) (int, chan *imageTask, error) {
	r, err := zip.OpenReader(input)
	if err != nil {
		return 0, nil, err
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
		return 0, nil, fmt.Errorf("no images found")
	}

	sort.SliceStable(images, func(i, j int) bool {
		return strings.Compare(images[i].Name, images[j].Name) < 0
	})

	output := make(chan *imageTask)
	go func() {
		defer close(output)
		for i, img := range images {
			f, err := img.Open()
			if err != nil {
				fmt.Println(err)
				os.Exit(1)
			}
			output <- &imageTask{
				Id:     i,
				Reader: f,
			}
		}
	}()
	return len(images), output, nil
}

func loadCbr(input string) (int, chan *imageTask, error) {
	rr, err := os.Open(input)
	if err != nil {
		return 0, nil, err
	}
	defer rr.Close()
	rs, err := rr.Stat()
	if err != nil {
		return 0, nil, err
	}

	bar := progressbar.DefaultBytes(rs.Size(), "Uncompressing rar")
	defer bar.Close()

	r, err := rardecode.NewReader(io.TeeReader(rr, bar), "")
	if err != nil {
		return 0, nil, err
	}

	type imageContent struct {
		Name string
		Data io.ReadCloser
	}

	images := make([]*imageContent, 0)

	for {
		f, err := r.Next()

		if f == nil {
			break
		}

		if err != nil {
			return 0, nil, err
		}

		if f.IsDir {
			continue
		}

		if strings.ToLower(filepath.Ext(f.Name)) != ".jpg" {
			continue
		}

		b := bytes.NewBuffer([]byte{})
		io.Copy(b, r)

		images = append(images, &imageContent{
			Name: f.Name,
			Data: readFakeCloser{b},
		})
	}

	if len(images) == 0 {
		return 0, nil, fmt.Errorf("no images found")
	}

	sort.SliceStable(images, func(i, j int) bool {
		return strings.Compare(images[i].Name, images[j].Name) < 0
	})

	output := make(chan *imageTask)
	go func() {
		defer close(output)
		for i, img := range images {
			output <- &imageTask{
				Id:     i,
				Reader: img.Data,
			}
		}
	}()
	return len(images), output, nil
}
