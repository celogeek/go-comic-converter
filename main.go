package main

import (
	"fmt"
	comicconverter "go-comic-converter/internal/comic-converter"
	"io/fs"
	"path/filepath"
	"runtime"
	"strings"
	"sync"
)

type Todo struct {
	Input  string
	Output string
}

func main() {
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
