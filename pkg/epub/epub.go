package epub

import (
	"github.com/celogeek/go-comic-converter/v2/internal/epub"
	"github.com/celogeek/go-comic-converter/v2/pkg/epuboptions"
)

func Generate(options *epuboptions.EPUBOptions) error {
	return epub.New(options).Write()
}
