// Package epuboptions for EPUB creation.
package epuboptions

import (
	"os"

	"github.com/golang/freetype/truetype"
)

type EPUBOptions struct {
	// Output
	Input  string `yaml:"-" json:"input"`
	Output string `yaml:"-" json:"output"`
	Author string `yaml:"-" json:"author"`
	Title  string `yaml:"-" json:"title"`

	//Config
	TitlePage                  int            `yaml:"title_page" json:"title_page"`
	CoverCaption               bool           `yaml:"cover_caption" json:"cover_caption"`
	Font                       string         `yaml:"font" json:"font"`
	TrueTypeFont               *truetype.Font `yaml:"-" json:"-"`
	LimitMb                    int            `yaml:"limit_mb" json:"limit_mb"`
	StripFirstDirectoryFromToc bool           `yaml:"strip_first_directory" json:"strip_first_directory"`
	SortPathMode               int            `yaml:"sort_path_mode" json:"sort_path_mode"`
	Image                      Image          `yaml:"image" json:"image"`

	// Other
	Dry     bool `yaml:"-" json:"dry"`
	Workers int  `yaml:"-" json:"workers"`

	// Progress, both are optional
	OnProcess Progress `yaml:"-" json:"-"`
	OnWrite   Progress `yaml:"-" json:"-"`
}

// Progress is called each time an item is done, with the number of items done and the total of items.
//
// It can be called from several goroutines at the same time, so the done counts can reach it out of order.
type Progress func(done, total int)

// Processed reports the progress of the images processed.
func (o EPUBOptions) Processed(done, total int) {
	if o.OnProcess != nil {
		o.OnProcess(done, total)
	}
}

// Written reports the progress of the files written.
func (o EPUBOptions) Written(done, total int) {
	if o.OnWrite != nil {
		o.OnWrite(done, total)
	}
}

func (o EPUBOptions) WorkersRatio(pct int) (nbWorkers int) {
	nbWorkers = o.Workers * pct / 100
	if nbWorkers < 1 {
		nbWorkers = 1
	}
	return
}

// ImgStorage is the temporary file that holds the processed images, a dry run writes nothing.
func (o EPUBOptions) ImgStorage() string {
	if o.Dry {
		return os.DevNull
	}
	return o.Output + ".tmp"
}
