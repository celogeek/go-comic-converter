/*
Options for EPUB creation.
*/
package epuboptions

import "fmt"

type Crop struct {
	Enabled                 bool
	Left, Up, Right, Bottom int
}

type Color struct {
	Foreground, Background string
}

type View struct {
	Width, Height int
	Color         Color
}

type Image struct {
	Crop                *Crop
	Quality             int
	Brightness          int
	Contrast            int
	AutoRotate          bool
	AutoSplitDoublePage bool
	NoBlankImage        bool
	Manga               bool
	HasCover            bool
	View                *View
	GrayScale           bool
	Resize              bool
	Format              string
}

type Options struct {
	Input                      string
	Output                     string
	Title                      string
	Author                     string
	LimitMb                    int
	StripFirstDirectoryFromToc bool
	Dry                        bool
	DryVerbose                 bool
	SortPathMode               int
	Quiet                      bool
	Workers                    int
	Image                      *Image
}

func (o *Options) WorkersRatio(pct int) (nbWorkers int) {
	nbWorkers = o.Workers * pct / 100
	if nbWorkers < 1 {
		nbWorkers = 1
	}
	return
}

func (o *Options) ImgStorage() string {
	return fmt.Sprintf("%s.tmp", o.Output)
}
