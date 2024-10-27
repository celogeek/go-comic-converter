// Package epuboptions for EPUB creation.
package epuboptions

type EPUBOptions struct {
	// Output
	Input  string `yaml:"-"`
	Output string `yaml:"-"`
	Author string `yaml:"-"`
	Title  string `yaml:"-"`

	//Config
	TitlePage                  int   `yaml:"title_page"`
	LimitMb                    int   `yaml:"limit_mb"`
	StripFirstDirectoryFromToc bool  `yaml:"strip_first_directory"`
	SortPathMode               int   `yaml:"sort_path_mode"`
	Image                      Image `yaml:"image"`

	// Other
	Dry        bool `yaml:"-"`
	DryVerbose bool `yaml:"-"`
	Quiet      bool `yaml:"-"`
	Json       bool `yaml:"-"`
	Workers    int  `yaml:"-"`
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
