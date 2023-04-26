/*
Manage options with default value from config.
*/
package options

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/celogeek/go-comic-converter/v2/internal/converter/profiles"
	"gopkg.in/yaml.v3"
)

type Options struct {
	// Output
	Input  string `yaml:"-"`
	Output string `yaml:"-"`
	Author string `yaml:"-"`
	Title  string `yaml:"-"`

	// Config
	Profile                    string `yaml:"profile"`
	Quality                    int    `yaml:"quality"`
	Crop                       bool   `yaml:"crop"`
	Brightness                 int    `yaml:"brightness"`
	Contrast                   int    `yaml:"contrast"`
	Auto                       bool   `yaml:"-"`
	AutoRotate                 bool   `yaml:"auto_rotate"`
	AutoSplitDoublePage        bool   `yaml:"auto_split_double_page"`
	NoBlankPage                bool   `yaml:"no_blank_page"`
	Manga                      bool   `yaml:"manga"`
	HasCover                   bool   `yaml:"has_cover"`
	LimitMb                    int    `yaml:"limit_mb"`
	StripFirstDirectoryFromToc bool   `yaml:"strip_first_directory_from_toc"`
	SortPathMode               int    `yaml:"sort_path_mode"`

	// Default Config
	Show  bool `yaml:"-"`
	Save  bool `yaml:"-"`
	Reset bool `yaml:"-"`

	// Other
	Workers    int  `yaml:"-"`
	Dry        bool `yaml:"-"`
	DryVerbose bool `yaml:"-"`
	Quiet      bool `yaml:"-"`
	Version    bool `yaml:"-"`
	Help       bool `yaml:"-"`

	// Internal
	profiles profiles.Profiles
}

// Initialize default options.
func New() *Options {
	return &Options{
		Profile:                    "",
		Quality:                    85,
		Crop:                       true,
		Brightness:                 0,
		Contrast:                   0,
		AutoRotate:                 false,
		AutoSplitDoublePage:        false,
		NoBlankPage:                false,
		Manga:                      false,
		HasCover:                   true,
		LimitMb:                    0,
		StripFirstDirectoryFromToc: false,
		SortPathMode:               1,
		profiles:                   profiles.New(),
	}
}

func (o *Options) Header() string {
	return `Go Comic Converter

Options:`
}

func (o *Options) String() string {
	return fmt.Sprintf(`%s
    Input                     : %s
    Output                    : %s
    Author                    : %s
    Title                     : %s
    Workers                   : %d%s
`,
		o.Header(),
		o.Input,
		o.Output,
		o.Author,
		o.Title,
		o.Workers,
		o.ShowConfig(),
	)
}

// Config file: ~/.go-comic-converter.yaml
func (o *Options) FileName() string {
	home, _ := os.UserHomeDir()
	return filepath.Join(home, ".go-comic-converter.yaml")
}

// Load config files
func (o *Options) LoadConfig() error {
	f, err := os.Open(o.FileName())
	if err != nil {
		return nil
	}
	defer f.Close()
	err = yaml.NewDecoder(f).Decode(o)
	if err != nil && err.Error() != "EOF" {
		return err
	}

	return nil
}

// Get current settings for fields that can be saved
func (o *Options) ShowConfig() string {
	var profileDesc, viewDesc string
	profile := o.GetProfile()
	if profile != nil {
		profileDesc = fmt.Sprintf(
			"%s - %s - %dx%d",
			o.Profile,
			profile.Description,
			profile.Width,
			profile.Height,
		)

		perfectWidth, perfectHeight := profile.PerfectDim()
		viewDesc = fmt.Sprintf(
			"%dx%d",
			perfectWidth,
			perfectHeight,
		)
	}
	limitmb := "nolimit"
	if o.LimitMb > 0 {
		limitmb = fmt.Sprintf("%d Mb", o.LimitMb)
	}

	sortpathmode := ""
	switch o.SortPathMode {
	case 0:
		sortpathmode = "path=alpha, file=alpha"
	case 1:
		sortpathmode = "path=alphanum, file=alpha"
	case 2:
		sortpathmode = "path=alphanum, file=alphanum"
	}

	return fmt.Sprintf(`
    Profile                   : %s
    ViewRatio                 : 1:%s
    View                      : %s
    Quality                   : %d
    Crop                      : %v
    Brightness                : %d
    Contrast                  : %d
    AutoRotate                : %v
    AutoSplitDoublePage       : %v
    NoBlankPage               : %v
    Manga                     : %v
    HasCover                  : %v
    LimitMb                   : %s
    StripFirstDirectoryFromToc: %v
    SortPathMode              : %s`,
		profileDesc,
		strings.TrimRight(fmt.Sprintf("%f", profiles.PerfectRatio), "0"),
		viewDesc,
		o.Quality,
		o.Crop,
		o.Brightness,
		o.Contrast,
		o.AutoRotate,
		o.AutoSplitDoublePage,
		o.NoBlankPage,
		o.Manga,
		o.HasCover,
		limitmb,
		o.StripFirstDirectoryFromToc,
		sortpathmode,
	)
}

// reset all settings to default value
func (o *Options) ResetConfig() error {
	New().SaveConfig()
	return o.LoadConfig()
}

// save all current settings as futur default value
func (o *Options) SaveConfig() error {
	f, err := os.Create(o.FileName())
	if err != nil {
		return err
	}
	defer f.Close()
	return yaml.NewEncoder(f).Encode(o)
}

// shortcut to get current profile
func (o *Options) GetProfile() *profiles.Profile {
	return o.profiles.Get(o.Profile)
}

// all available profiles
func (o *Options) AvailableProfiles() string {
	return o.profiles.String()
}
