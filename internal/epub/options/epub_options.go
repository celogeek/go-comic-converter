/*
Options for EPUB creation.
*/
package epuboptions

type Crop struct {
	Enabled                 bool
	Left, Up, Right, Bottom int
}

type View struct {
	Width, Height int
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
