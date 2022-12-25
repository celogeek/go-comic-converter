package main

import (
	"fmt"
	comicconverter "go-comic-converter/internal/comic-converter"
)

func main() {
	cv := comicconverter.
		New(comicconverter.ComicConverterOptions{
			Quality: 75,
		})

	for i := 1; i < 10; i++ {
		cv.
			Load(fmt.Sprintf("/Users/vincent/Downloads/0000%d.jpg", i)).
			CropMarging().
			Resize(1860, 2480).
			Save(fmt.Sprintf("/Users/vincent/Downloads/0000%d_gray.jpg", i))
	}

}
