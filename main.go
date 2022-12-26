package main

import (
	"fmt"
	"go-comic-converter/internal/epub"
	imageconverter "go-comic-converter/internal/image-converter"
	"io/fs"
	"path/filepath"
	"runtime"
	"sort"
	"strings"
	"sync"

	epub2 "github.com/bmaupin/go-epub"
)

type File struct {
	Path         string
	Name         string
	Title        string
	Data         string
	InternalPath string
}

func addImages(doc *epub2.Epub, imagesPath []string) {
	wg := &sync.WaitGroup{}
	todos := make(chan string, runtime.NumCPU())
	imageResult := make(chan *File)

	wg.Add(runtime.NumCPU())
	for i := 0; i < runtime.NumCPU(); i++ {
		go func() {
			defer wg.Done()
			for imagePath := range todos {
				name := filepath.Base(imagePath)
				ext := filepath.Ext(name)
				title := name[0 : len(name)-len(ext)]
				imageResult <- &File{
					Path:  imagePath,
					Name:  name,
					Title: title,
					Data:  imageconverter.Convert(imagePath, true, 1860, 2480, 75),
				}
			}
		}()
	}
	go func() {
		for _, imagePath := range imagesPath {
			todos <- imagePath
		}
		close(todos)
		wg.Wait()
		close(imageResult)
	}()

	results := make([]*File, 0)
	for result := range imageResult {
		fmt.Println(result.Name)
		internalPath, _ := doc.AddImage(result.Data, result.Name)
		result.InternalPath = internalPath
		results = append(results, result)
	}
	sort.SliceStable(results, func(i, j int) bool {
		return strings.Compare(
			results[i].Path, results[j].Path,
		) < 0
	})
	for i, result := range results {
		if i == 0 {
			doc.SetCover(result.InternalPath, "")
		} else {
			doc.AddSection(
				fmt.Sprintf("<img src=\"%s\" />", result.InternalPath),
				result.Title,
				fmt.Sprintf("%s.xhtml", result.Title),
				"../css/cover.css",
			)
		}
	}
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

func main2() {
	imagesPath := getImages("/Users/vincent/Downloads/Bleach T01 (Tite KUBO) [eBook officiel 1920]")

	doc := epub2.NewEpub("Bleach T01 (Tite KUBO) [eBook officiel 1920]")
	doc.SetAuthor("Bachelier Vincent")

	addImages(doc, imagesPath)

	if err := doc.Write("/Users/vincent/Downloads/test.epub"); err != nil {
		panic(err)
	}

}

func main() {
	err := epub.NewEpub("/Users/vincent/Downloads/test.epub").
		SetSize(1860, 2480).
		SetQuality(75).
		SetTitle("Bleach T01 (Tite KUBO) [eBook officiel 1920]").
		SetAuthor("Bachelier Vincent").
		LoadDir("/Users/vincent/Downloads/Bleach T01 (Tite KUBO) [eBook officiel 1920]").
		Write()

	if err != nil {
		panic(err)
	}
}
