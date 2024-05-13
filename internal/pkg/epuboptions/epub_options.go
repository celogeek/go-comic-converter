// Package epuboptions EPUBOptions for EPUB creation.
package epuboptions

type EPUBOptions struct {
	Input                      string
	Output                     string
	Title                      string
	TitlePage                  int
	Author                     string
	LimitMb                    int
	StripFirstDirectoryFromToc bool
	Dry                        bool
	DryVerbose                 bool
	SortPathMode               int
	Quiet                      bool
	Json                       bool
	Workers                    int
	Image                      Image
}

func (o EPUBOptions) WorkersRatio(pct int) (nbWorkers int) {
	nbWorkers = o.Workers * pct / 100
	if nbWorkers < 1 {
		nbWorkers = 1
	}
	return
}

func (o EPUBOptions) ImgStorage() string {
	return o.Output + ".tmp"
}
