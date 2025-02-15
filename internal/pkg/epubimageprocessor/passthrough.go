package epubimageprocessor

import "github.com/celogeek/go-comic-converter/v3/internal/pkg/epubimage"

func (e EPUBImageProcessor) PassThrough() (images []epubimage.EPUBImage, err error) {
	images = make([]epubimage.EPUBImage, 0)
	return images, nil
}
