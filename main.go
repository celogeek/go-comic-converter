package main

import (
	"fmt"
	comicconverter "go-comic-converter/internal/comic-converter"
	"io/fs"
	"path/filepath"
	"runtime"
	"sort"
	"strings"
	"sync"

	"github.com/bmaupin/go-epub"
)

type Todo struct {
	Input  string
	Output string
}

func addImages(doc *epub.Epub, imagesPath string) {
	wg := &sync.WaitGroup{}
	todos := make(chan Todo, runtime.NumCPU())

	wg.Add(runtime.NumCPU())
	for i := 0; i < runtime.NumCPU(); i++ {
		go func() {
			defer wg.Done()
			for todo := range todos {
				fmt.Printf("Processing %s\n", todo.Input)
				comicconverter.Save(
					comicconverter.Resize(
						comicconverter.CropMarging(
							comicconverter.Load(todo.Input),
						), 1860, 2480), todo.Output, 75,
				)
			}
		}()
	}

	dirname := "/Users/vincent/Downloads/Bleach T01 (Tite KUBO) [eBook officiel 1920]"
	filepath.WalkDir(dirname, func(path string, d fs.DirEntry, err error) error {
		if d.IsDir() {
			return nil
		}
		input := path
		ext := filepath.Ext(path)
		if strings.ToLower(ext) != ".jpg" {
			return nil
		}
		output := fmt.Sprintf("%s_gray%s", input[0:len(input)-len(ext)], ext)

		todos <- Todo{input, output}

		return nil
	})

	close(todos)

	wg.Wait()
}

func getImages(dirname string) []string {
	images := make([]string, 0)
	filepath.WalkDir(dirname, func(path string, d fs.DirEntry, err error) error {
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
	sort.Strings(images)
	return images
}

func main() {
	imagesPath := getImages("/Users/vincent/Downloads/Bleach T01 (Tite KUBO) [eBook officiel 1920]")

	doc := epub.NewEpub("Bleach T01 (Tite KUBO) [eBook officiel 1920]")
	doc.SetAuthor("Bachelier Vincent")

	for i, imagePath := range imagesPath {
		fmt.Printf("%04d / %04d\n", i+1, len(imagesPath))
		name := filepath.Base(imagePath)
		ext := filepath.Ext(name)
		title := name[0 : len(name)-len(ext)]

		img := comicconverter.Convert(imagePath, true, 1860, 2480, 75)
		if i == 0 {
			doc.SetCover(img, "")
		} else {
			imgPath, _ := doc.AddImage(img, name)
			doc.AddSection(fmt.Sprintf("<img src=\"%s\" />", imgPath), title, fmt.Sprintf("%s.xhtml", title), "../css/cover.css")
		}
	}

	if err := doc.Write("/Users/vincent/Downloads/test.epub"); err != nil {
		panic(err)
	}

}
