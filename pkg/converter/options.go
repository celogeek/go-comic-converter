// Package converter options manage options with default value from config.
package converter

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"gopkg.in/yaml.v3"

	"github.com/celogeek/go-comic-converter/v2/pkg/utils"
)

type Options struct {
	// Output
	Input  string `yaml:"-"`
	Output string `yaml:"-"`
	Author string `yaml:"-"`
	Title  string `yaml:"-"`

	// Config
	Profile                    string  `yaml:"profile"`
	Quality                    int     `yaml:"quality"`
	Grayscale                  bool    `yaml:"grayscale"`
	GrayscaleMode              int     `yaml:"grayscale_mode"` // 0 = normal, 1 = average, 2 = luminance
	Crop                       bool    `yaml:"crop"`
	CropRatioLeft              int     `yaml:"crop_ratio_left"`
	CropRatioUp                int     `yaml:"crop_ratio_up"`
	CropRatioRight             int     `yaml:"crop_ratio_right"`
	CropRatioBottom            int     `yaml:"crop_ratio_bottom"`
	CropLimit                  int     `yaml:"crop_limit"`
	CropSkipIfLimitReached     bool    `yaml:"crop_skip_if_limit_reached"`
	Brightness                 int     `yaml:"brightness"`
	Contrast                   int     `yaml:"contrast"`
	AutoContrast               bool    `yaml:"auto_contrast"`
	AutoRotate                 bool    `yaml:"auto_rotate"`
	AutoSplitDoublePage        bool    `yaml:"auto_split_double_page"`
	KeepDoublePageIfSplit      bool    `yaml:"keep_double_page_if_split"`
	KeepSplitDoublePageAspect  bool    `yaml:"keep_split_double_page_aspect"`
	NoBlankImage               bool    `yaml:"no_blank_image"`
	Manga                      bool    `yaml:"manga"`
	HasCover                   bool    `yaml:"has_cover"`
	LimitMb                    int     `yaml:"limit_mb"`
	StripFirstDirectoryFromToc bool    `yaml:"strip_first_directory_from_toc"`
	SortPathMode               int     `yaml:"sort_path_mode"`
	ForegroundColor            string  `yaml:"foreground_color"`
	BackgroundColor            string  `yaml:"background_color"`
	NoResize                   bool    `yaml:"noresize"`
	Format                     string  `yaml:"format"`
	AspectRatio                float64 `yaml:"aspect_ratio"`
	PortraitOnly               bool    `yaml:"portrait_only"`
	AppleBookCompatibility     bool    `yaml:"apple_book_compatibility"`
	TitlePage                  int     `yaml:"title_page"`

	// Default Config
	Show  bool `yaml:"-"`
	Save  bool `yaml:"-"`
	Reset bool `yaml:"-"`

	// Shortcut
	Auto         bool `yaml:"-"`
	NoFilter     bool `yaml:"-"`
	MaxQuality   bool `yaml:"-"`
	BestQuality  bool `yaml:"-"`
	GreatQuality bool `yaml:"-"`
	GoodQuality  bool `yaml:"-"`

	// Other
	Workers    int  `yaml:"-"`
	Dry        bool `yaml:"-"`
	DryVerbose bool `yaml:"-"`
	Quiet      bool `yaml:"-"`
	Json       bool `yaml:"-"`
	Version    bool `yaml:"-"`
	Help       bool `yaml:"-"`

	// Internal
	profiles Profiles
}

// NewOptions Initialize default options.
func NewOptions() *Options {
	return &Options{
		Profile:                   "SR",
		Quality:                   85,
		Grayscale:                 true,
		Crop:                      true,
		CropRatioLeft:             1,
		CropRatioUp:               1,
		CropRatioRight:            1,
		CropRatioBottom:           3,
		NoBlankImage:              true,
		HasCover:                  true,
		KeepDoublePageIfSplit:     true,
		KeepSplitDoublePageAspect: true,
		SortPathMode:              1,
		ForegroundColor:           "000",
		BackgroundColor:           "FFF",
		Format:                    "jpeg",
		TitlePage:                 1,
		profiles:                  NewProfiles(),
	}
}

func (o *Options) Header() string {
	return "Go Comic Converter\n\nOptions:"
}

func (o *Options) String() string {
	var b strings.Builder
	b.WriteString(o.Header())
	for _, v := range []struct {
		K string
		V any
	}{
		{"Input", o.Input},
		{"Output", o.Output},
		{"Author", o.Author},
		{"Title", o.Title},
		{"Workers", o.Workers},
	} {
		b.WriteString(fmt.Sprintf("\n    %-32s: %v", v.K, v.V))
	}
	b.WriteString(o.ShowConfig())
	b.WriteRune('\n')
	return b.String()
}

func (o *Options) MarshalJSON() ([]byte, error) {
	out := map[string]any{
		"input":                      o.Input,
		"output":                     o.Output,
		"author":                     o.Author,
		"title":                      o.Title,
		"workers":                    o.Workers,
		"profile":                    o.GetProfile(),
		"format":                     o.Format,
		"grayscale":                  o.Grayscale,
		"crop":                       o.Crop,
		"autocontrast":               o.AutoContrast,
		"autorotate":                 o.AutoRotate,
		"noblankimage":               o.NoBlankImage,
		"manga":                      o.Manga,
		"hascover":                   o.HasCover,
		"stripfirstdirectoryfromtoc": o.StripFirstDirectoryFromToc,
		"sortpathmode":               o.SortPathMode,
		"foregroundcolor":            o.ForegroundColor,
		"backgroundcolor":            o.BackgroundColor,
		"resize":                     !o.NoResize,
		"aspectratio":                o.AspectRatio,
		"portraitonly":               o.PortraitOnly,
		"titlepage":                  o.TitlePage,
	}
	if o.Format == "jpeg" {
		out["quality"] = o.Quality
	}
	if o.Grayscale {
		out["grayscale_mode"] = o.GrayscaleMode
	}
	if o.Crop {
		out["crop_ratio"] = map[string]any{
			"left":   o.CropRatioLeft,
			"right":  o.CropRatioRight,
			"up":     o.CropRatioUp,
			"bottom": o.CropRatioBottom,
		}
		out["crop_limit"] = o.CropLimit
		out["crop_skip_if_limit_reached"] = o.CropSkipIfLimitReached
	}
	if o.Brightness != 0 {
		out["brightness"] = o.Brightness
	}
	if o.Contrast != 0 {
		out["contrast"] = o.Contrast
	}
	if o.PortraitOnly || !o.AppleBookCompatibility {
		out["autosplitdoublepage"] = o.AutoSplitDoublePage
		if o.AutoSplitDoublePage {
			out["keepdoublepageifsplit"] = o.KeepDoublePageIfSplit
			out["keepsplitdoublepageaspect"] = o.KeepSplitDoublePageAspect
		}
	}
	if o.LimitMb != 0 {
		out["limitmb"] = o.LimitMb
	}
	if !o.PortraitOnly {
		out["applebookcompatibility"] = o.AppleBookCompatibility
	}
	return json.Marshal(out)
}

// FileName Config file: ~/.go-comic-converter.yaml
func (o *Options) FileName() string {
	home, _ := os.UserHomeDir()
	return filepath.Join(home, ".go-comic-converter.yaml")
}

// LoadConfig Load config files
func (o *Options) LoadConfig() error {
	f, err := os.Open(o.FileName())
	if err != nil {
		return nil
	}
	defer func(f *os.File) {
		_ = f.Close()
	}(f)
	err = yaml.NewDecoder(f).Decode(o)
	if err != nil && err.Error() != "EOF" {
		return err
	}

	return nil
}

// ShowConfig Get current settings for fields that can be saved
func (o *Options) ShowConfig() string {
	var profileDesc string
	profile := o.GetProfile()
	if profile != nil {
		profileDesc = profile.String()
	}

	sortpathmode := ""
	switch o.SortPathMode {
	case 0:
		sortpathmode = "path=alpha, file=alpha"
	case 1:
		sortpathmode = "path=alphanumeric, file=alpha"
	case 2:
		sortpathmode = "path=alphanumeric, file=alphanumeric"
	}

	aspectRatio := "auto"
	if o.AspectRatio > 0 {
		aspectRatio = "1:" + utils.FloatToString(o.AspectRatio, 2)
	} else if o.AspectRatio < 0 {
		if profile != nil {
			aspectRatio = "1:" + utils.FloatToString(float64(profile.Height)/float64(profile.Width), 2) + " (device)"
		} else {
			aspectRatio = "1:?? (device)"
		}
	}

	titlePage := ""
	switch o.TitlePage {
	case 0:
		titlePage = "never"
	case 1:
		titlePage = "always"
	case 2:
		titlePage = "when epub is split"
	}

	grayscaleMode := "normal"
	switch o.GrayscaleMode {
	case 1:
		grayscaleMode = "average"
	case 2:
		grayscaleMode = "luminance"
	}

	var b strings.Builder
	for _, v := range []struct {
		Key       string
		Value     any
		Condition bool
	}{
		{"Profile", profileDesc, true},
		{"Format", o.Format, true},
		{"Quality", o.Quality, o.Format == "jpeg"},
		{"Grayscale", o.Grayscale, true},
		{"Grayscale mode", grayscaleMode, o.Grayscale},
		{"Crop", o.Crop, true},
		{"Crop ratio",
			utils.IntToString(o.CropRatioLeft) + " Left - " +
				utils.IntToString(o.CropRatioUp) + " Up - " +
				utils.IntToString(o.CropRatioRight) + " Right - " +
				utils.IntToString(o.CropRatioBottom) + " Bottom - " +
				"Limit " + utils.IntToString(o.CropLimit) + "% - " +
				"Skip " + utils.BoolToString(o.CropSkipIfLimitReached),
			o.Crop},
		{"Brightness", o.Brightness, o.Brightness != 0},
		{"Contrast", o.Contrast, o.Contrast != 0},
		{"Auto contrast", o.AutoContrast, true},
		{"Auto rotate", o.AutoRotate, true},
		{"Auto split double page", o.AutoSplitDoublePage, o.PortraitOnly || !o.AppleBookCompatibility},
		{"Keep double page if split", o.KeepDoublePageIfSplit, (o.PortraitOnly || !o.AppleBookCompatibility) && o.AutoSplitDoublePage},
		{"Keep split double page aspect", o.KeepSplitDoublePageAspect, (o.PortraitOnly || !o.AppleBookCompatibility) && o.AutoSplitDoublePage},
		{"No blank image", o.NoBlankImage, true},
		{"Manga", o.Manga, true},
		{"Has cover", o.HasCover, true},
		{"Limit", utils.IntToString(o.LimitMb) + " Mb", o.LimitMb != 0},
		{"Strip first directory from toc", o.StripFirstDirectoryFromToc, true},
		{"Sort path mode", sortpathmode, true},
		{"Foreground color", "#" + o.ForegroundColor, true},
		{"Background color", "#" + o.BackgroundColor, true},
		{"Resize", !o.NoResize, true},
		{"Aspect ratio", aspectRatio, true},
		{"Portrait only", o.PortraitOnly, true},
		{"Title page", titlePage, true},
		{"Apple book compatibility", o.AppleBookCompatibility, !o.PortraitOnly},
	} {
		if v.Condition {
			b.WriteString(fmt.Sprintf("\n    %-32s: %v", v.Key, v.Value))
		}
	}
	return b.String()
}

// ResetConfig reset all settings to default value
func (o *Options) ResetConfig() error {
	if err := NewOptions().SaveConfig(); err != nil {
		return err
	}
	return o.LoadConfig()
}

// SaveConfig save all current settings as default value
func (o *Options) SaveConfig() error {
	f, err := os.Create(o.FileName())
	if err != nil {
		return err
	}
	defer func(f *os.File) {
		_ = f.Close()
	}(f)
	return yaml.NewEncoder(f).Encode(o)
}

// GetProfile shortcut to get current profile
func (o *Options) GetProfile() *Profile {
	if p, ok := o.profiles[o.Profile]; ok {
		return &p
	}
	return nil
}

// AvailableProfiles all available profiles
func (o *Options) AvailableProfiles() string {
	return o.profiles.String()
}
