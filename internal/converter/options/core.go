package options

import (
	"fmt"
	"os"
	"path/filepath"

	"github.com/celogeek/go-comic-converter/internal/converter/profiles"
	"gopkg.in/yaml.v3"
)

type Options struct {
	Input               string `yaml:"-"`
	Output              string `yaml:"-"`
	Author              string `yaml:"-"`
	Title               string `yaml:"-"`
	Auto                bool   `yaml:"-"`
	Workers             int    `yaml:"-"`
	Dry                 bool   `yaml:"-"`
	Show                bool   `yaml:"-"`
	Save                bool   `yaml:"-"`
	Help                bool   `yaml:"-"`
	Profile             string `yaml:"profile"`
	Quality             int    `yaml:"quality"`
	Crop                bool   `yaml:"crop"`
	Brightness          int    `yaml:"brightness"`
	Contrast            int    `yaml:"contrast"`
	AutoRotate          bool   `yaml:"auto_rotate"`
	AutoSplitDoublePage bool   `yaml:"auto_split_double_page"`
	NoBlankPage         bool   `yaml:"no_blank_page"`
	Manga               bool   `yaml:"manga"`
	HasCover            bool   `yaml:"has_cover"`
	AddPanelView        bool   `yaml:"add_panel_view"`
	LimitMb             int    `yaml:"limit_mb"`

	profiles profiles.Profiles
}

func New() *Options {
	return &Options{
		Profile:             "",
		Quality:             85,
		Crop:                true,
		Brightness:          0,
		Contrast:            0,
		AutoRotate:          false,
		AutoSplitDoublePage: false,
		NoBlankPage:         false,
		Manga:               false,
		HasCover:            true,
		AddPanelView:        false,
		LimitMb:             0,
		profiles:            profiles.New(),
	}
}

func (o *Options) Header() string {
	return `Go Comic Converter

Options:`
}

func (o *Options) String() string {
	return fmt.Sprintf(`%s
    Input              : %s
    Output             : %s
    Author             : %s
    Title              : %s
    Workers            : %d%s
`,
		o.Header(),
		o.Input,
		o.Output,
		o.Author,
		o.Title,
		o.Workers,
		o.ShowDefault(),
	)
}

func (o *Options) FileName() string {
	home, _ := os.UserHomeDir()
	return filepath.Join(home, ".go-comic-converter.yaml")
}

func (o *Options) LoadDefault() error {
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

func (o *Options) ShowDefault() string {
	var profileDesc string
	profile := o.GetProfile()
	if profile != nil {
		profileDesc = fmt.Sprintf(
			"%s - %s - %dx%d - %d levels of gray",
			o.Profile,
			profile.Description,
			profile.Width,
			profile.Height,
			len(profile.Palette),
		)
	}
	limitmb := "nolimit"
	if o.LimitMb > 0 {
		limitmb = fmt.Sprintf("%d Mb", o.LimitMb)
	}

	return fmt.Sprintf(`
    Profile            : %s
    Quality            : %d
    Crop               : %v
    Brightness         : %d
    Contrast           : %d
    AutoRotate         : %v
    AutoSplitDoublePage: %v
    NoBlankPage        : %v
    Manga              : %v
    HasCover           : %v
    AddPanelView       : %v
    LimitMb            : %s`,
		profileDesc,
		o.Quality,
		o.Crop,
		o.Brightness,
		o.Contrast,
		o.AutoRotate,
		o.AutoSplitDoublePage,
		o.NoBlankPage,
		o.Manga,
		o.HasCover,
		o.AddPanelView,
		limitmb,
	)
}

func (o *Options) SaveDefault() error {
	f, err := os.Create(o.FileName())
	if err != nil {
		return err
	}
	defer f.Close()
	return yaml.NewEncoder(f).Encode(o)
}

func (o *Options) GetProfile() *profiles.Profile {
	return o.profiles.Get(o.Profile)
}

func (o *Options) AvailableProfiles() string {
	return o.profiles.String()
}
