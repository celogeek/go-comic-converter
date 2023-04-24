package epub

import (
	"archive/zip"
	"bytes"
	"fmt"
	"image"
	"image/color"
	_ "image/jpeg"
	_ "image/png"
	"io"
	"io/fs"
	"os"
	"path/filepath"
	"sort"
	"strings"
	"sync"

	epubfilters "github.com/celogeek/go-comic-converter/v2/internal/epub/filters"
	epubimage "github.com/celogeek/go-comic-converter/v2/internal/epub/image"
	epubimagedata "github.com/celogeek/go-comic-converter/v2/internal/epub/imagedata"
	"github.com/celogeek/go-comic-converter/v2/internal/sortpath"
	"github.com/disintegration/gift"
	"github.com/nwaples/rardecode"
	pdfimage "github.com/raff/pdfreader/image"
	"github.com/raff/pdfreader/pdfread"
	"golang.org/x/image/tiff"
	_ "golang.org/x/image/webp"
)

type imageTask struct {
	Id     int
	Reader io.ReadCloser
	Path   string
	Name   string
}

func colorIsBlank(c color.Color) bool {
	g := color.GrayModel.Convert(c).(color.Gray)
	return g.Y >= 0xf0
}

func findMarging(img image.Image) image.Rectangle {
	imgArea := img.Bounds()

LEFT:
	for x := imgArea.Min.X; x < imgArea.Max.X; x++ {
		for y := imgArea.Min.Y; y < imgArea.Max.Y; y++ {
			if !colorIsBlank(img.At(x, y)) {
				break LEFT
			}
		}
		imgArea.Min.X++
	}

UP:
	for y := imgArea.Min.Y; y < imgArea.Max.Y; y++ {
		for x := imgArea.Min.X; x < imgArea.Max.X; x++ {
			if !colorIsBlank(img.At(x, y)) {
				break UP
			}
		}
		imgArea.Min.Y++
	}

RIGHT:
	for x := imgArea.Max.X - 1; x >= imgArea.Min.X; x-- {
		for y := imgArea.Min.Y; y < imgArea.Max.Y; y++ {
			if !colorIsBlank(img.At(x, y)) {
				break RIGHT
			}
		}
		imgArea.Max.X--
	}

BOTTOM:
	for y := imgArea.Max.Y - 1; y >= imgArea.Min.Y; y-- {
		for x := imgArea.Min.X; x < imgArea.Max.X; x++ {
			if !colorIsBlank(img.At(x, y)) {
				break BOTTOM
			}
		}
		imgArea.Max.Y--
	}

	return imgArea
}

func (e *ePub) LoadImages() ([]*epubimage.Image, error) {
	images := make([]*epubimage.Image, 0)

	fi, err := os.Stat(e.Input)
	if err != nil {
		return nil, err
	}

	var (
		imageCount int
		imageInput chan *imageTask
	)

	if fi.IsDir() {
		imageCount, imageInput, err = loadDir(e.Input, e.SortPathMode)
	} else {
		switch ext := strings.ToLower(filepath.Ext(e.Input)); ext {
		case ".cbz", ".zip":
			imageCount, imageInput, err = loadCbz(e.Input, e.SortPathMode)
		case ".cbr", ".rar":
			imageCount, imageInput, err = loadCbr(e.Input, e.SortPathMode)
		case ".pdf":
			imageCount, imageInput, err = loadPdf(e.Input)
		default:
			err = fmt.Errorf("unknown file format (%s): support .cbz, .zip, .cbr, .rar, .pdf", ext)
		}
	}
	if err != nil {
		return nil, err
	}

	if e.Dry {
		for img := range imageInput {
			img.Reader.Close()
			images = append(images, &epubimage.Image{
				Id:   img.Id,
				Path: img.Path,
				Name: img.Name,
			})
		}

		return images, nil
	}

	imageOutput := make(chan *epubimage.Image)

	// processing
	bar := e.NewBar(imageCount, "Processing", 1, 2)
	wg := &sync.WaitGroup{}

	for i := 0; i < e.Workers; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()

			for img := range imageInput {
				// Decode image
				src, _, err := image.Decode(img.Reader)
				img.Reader.Close()
				if err != nil {
					bar.Clear()
					fmt.Fprintf(os.Stderr, "error processing image %s%s: %s\n", img.Path, img.Name, err)
					os.Exit(1)
				}

				if e.Image.Crop {
					g := gift.New(gift.Crop(findMarging(src)))
					newSrc := image.NewNRGBA(g.Bounds(src.Bounds()))
					g.Draw(newSrc, src)
					src = newSrc
				}

				g := epubimage.NewGift(e.Options.Image)

				// Convert image
				dst := image.NewGray(g.Bounds(src.Bounds()))
				g.Draw(dst, src)

				var raw image.Image
				if img.Id == 0 {
					raw = dst
				}

				imageOutput <- &epubimage.Image{
					Id:      img.Id,
					Part:    0,
					Raw:     raw,
					Data:    epubimagedata.New(img.Id, 0, dst, e.Image.Quality),
					Width:   dst.Bounds().Dx(),
					Height:  dst.Bounds().Dy(),
					IsCover: img.Id == 0,
					DoublePage: src.Bounds().Dx() > src.Bounds().Dy() &&
						src.Bounds().Dx() > e.Image.ViewHeight &&
						src.Bounds().Dy() > e.Image.ViewWidth,
					Path: img.Path,
					Name: img.Name,
				}

				// Auto split double page
				// Except for cover
				// Only if the src image have width > height and is bigger than the view
				if (!e.Image.HasCover || img.Id > 0) &&
					e.Image.AutoSplitDoublePage &&
					src.Bounds().Dx() > src.Bounds().Dy() &&
					src.Bounds().Dx() > e.Image.ViewHeight &&
					src.Bounds().Dy() > e.Image.ViewWidth {
					gifts := epubimage.NewGiftSplitDoublePage(e.Options.Image)
					for i, g := range gifts {
						part := i + 1
						dst := image.NewGray(g.Bounds(src.Bounds()))
						g.Draw(dst, src)

						imageOutput <- &epubimage.Image{
							Id:         img.Id,
							Part:       part,
							Data:       epubimagedata.New(img.Id, part, dst, e.Image.Quality),
							Width:      dst.Bounds().Dx(),
							Height:     dst.Bounds().Dy(),
							IsCover:    false,
							DoublePage: false,
							Path:       img.Path,
							Name:       img.Name,
						}
					}
				}
			}
		}()
	}

	go func() {
		wg.Wait()
		close(imageOutput)
	}()

	for img := range imageOutput {
		if !(e.Image.NoBlankPage && img.Width == 1 && img.Height == 1) {
			images = append(images, img)
		}
		if img.Part == 0 {
			bar.Add(1)
		}
	}
	bar.Close()

	if len(images) == 0 {
		return nil, fmt.Errorf("image not found")
	}

	return images, nil
}

func isSupportedImage(path string) bool {
	switch strings.ToLower(filepath.Ext(path)) {
	case ".jpg", ".jpeg", ".png", ".webp":
		{
			return true
		}
	}
	return false
}

func loadDir(input string, sortpathmode int) (int, chan *imageTask, error) {
	images := make([]string, 0)
	input = filepath.Clean(input)
	err := filepath.WalkDir(input, func(path string, d fs.DirEntry, err error) error {
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
	if len(images) == 0 {
		return 0, nil, fmt.Errorf("image not found")
	}

	sort.Sort(sortpath.By(images, sortpathmode))

	output := make(chan *imageTask)
	go func() {
		defer close(output)
		for i, img := range images {
			f, err := os.Open(img)
			if err != nil {
				fmt.Fprintln(os.Stderr, err)
				os.Exit(1)
			}

			p, fn := filepath.Split(img)
			if p == input {
				p = ""
			} else {
				p = p[len(input)+1:]
			}
			output <- &imageTask{
				Id:     i,
				Reader: f,
				Path:   p,
				Name:   fn,
			}
		}
	}()
	return len(images), output, nil
}

func loadCbz(input string, sortpathmode int) (int, chan *imageTask, error) {
	r, err := zip.OpenReader(input)
	if err != nil {
		return 0, nil, err
	}

	images := make([]*zip.File, 0)
	for _, f := range r.File {
		if !f.FileInfo().IsDir() && isSupportedImage(f.Name) {
			images = append(images, f)
		}
	}
	if len(images) == 0 {
		r.Close()
		return 0, nil, fmt.Errorf("no images found")
	}

	names := []string{}
	for _, img := range images {
		names = append(names, img.Name)
	}
	sort.Sort(sortpath.By(names, sortpathmode))

	indexedNames := make(map[string]int)
	for i, name := range names {
		indexedNames[name] = i
	}

	output := make(chan *imageTask)
	go func() {
		defer close(output)
		for _, img := range images {
			f, err := img.Open()
			if err != nil {
				fmt.Fprintln(os.Stderr, err)
				os.Exit(1)
			}
			p, fn := filepath.Split(filepath.Clean(img.Name))
			output <- &imageTask{
				Id:     indexedNames[img.Name],
				Reader: f,
				Path:   p,
				Name:   fn,
			}
		}
	}()
	return len(images), output, nil
}

func loadCbr(input string, sortpathmode int) (int, chan *imageTask, error) {
	// listing and indexing
	rl, err := rardecode.OpenReader(input, "")
	if err != nil {
		return 0, nil, err
	}
	names := make([]string, 0)
	for {
		f, err := rl.Next()

		if err != nil && err != io.EOF {
			rl.Close()
			return 0, nil, err
		}

		if f == nil {
			break
		}

		if !f.IsDir && isSupportedImage(f.Name) {
			names = append(names, f.Name)
		}
	}
	rl.Close()

	if len(names) == 0 {
		return 0, nil, fmt.Errorf("no images found")
	}

	sort.Sort(sortpath.By(names, sortpathmode))

	indexedNames := make(map[string]int)
	for i, name := range names {
		indexedNames[name] = i
	}

	// send file to the queue
	output := make(chan *imageTask)
	go func() {
		defer close(output)
		r, err := rardecode.OpenReader(input, "")
		if err != nil {
			panic(err)
		}
		defer r.Close()

		for {
			f, err := r.Next()
			if err != nil && err != io.EOF {
				panic(err)
			}
			if f == nil {
				break
			}
			if idx, ok := indexedNames[f.Name]; ok {
				b := bytes.NewBuffer([]byte{})
				io.Copy(b, r)

				p, fn := filepath.Split(filepath.Clean(f.Name))

				output <- &imageTask{
					Id:     idx,
					Reader: io.NopCloser(b),
					Path:   p,
					Name:   fn,
				}
			}
		}
	}()

	return len(names), output, nil
}

func loadPdf(input string) (int, chan *imageTask, error) {
	pdf := pdfread.Load(input)
	if pdf == nil {
		return 0, nil, fmt.Errorf("can't read pdf")
	}

	nbPages := len(pdf.Pages())
	pageFmt := fmt.Sprintf("page %%0%dd", len(fmt.Sprintf("%d", nbPages)))
	output := make(chan *imageTask)
	go func() {
		defer close(output)
		defer pdf.Close()
		for i := 0; i < nbPages; i++ {
			img, err := pdfimage.Extract(pdf, i+1)
			if err != nil {
				panic(err)
			}

			b := bytes.NewBuffer([]byte{})
			err = tiff.Encode(b, img, nil)
			if err != nil {
				panic(err)
			}

			output <- &imageTask{
				Id:     i,
				Reader: io.NopCloser(b),
				Path:   "",
				Name:   fmt.Sprintf(pageFmt, i+1),
			}
		}
	}()

	return nbPages, output, nil
}

func (e *ePub) coverTitleImageData(title string, img *epubimage.Image, currentPart, totalPart int) *epubimagedata.ImageData {
	// Create a blur version of the cover
	g := gift.New(epubfilters.CoverTitle(title))
	dst := image.NewGray(g.Bounds(img.Raw.Bounds()))
	g.Draw(dst, img.Raw)

	return epubimagedata.NewRaw("OEBPS/Images/title.jpg", dst, e.Image.Quality)
}
