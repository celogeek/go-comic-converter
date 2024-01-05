/*
Manage options with default value from config.
*/
package converter

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"gopkg.in/yaml.v3"
)

type converterOptions struct {
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
	Brightness                 int     `yaml:"brightness"`
	Contrast                   int     `yaml:"contrast"`
	AutoContrast               bool    `yaml:"auto_contrast"`
	AutoRotate                 bool    `yaml:"auto_rotate"`
	AutoSplitDoublePage        bool    `yaml:"auto_split_double_page"`
	KeepDoublePageIfSplitted   bool    `yaml:"keep_double_page_if_splitted"`
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
	profiles converterProfiles
}

// Initialize default options.
func newOptions() *converterOptions {
	return &converterOptions{
		Profile:                  "SR",
		Quality:                  85,
		Grayscale:                true,
		Crop:                     true,
		CropRatioLeft:            1,
		CropRatioUp:              1,
		CropRatioRight:           1,
		CropRatioBottom:          3,
		NoBlankImage:             true,
		HasCover:                 true,
		KeepDoublePageIfSplitted: true,
		SortPathMode:             1,
		ForegroundColor:          "000",
		BackgroundColor:          "FFF",
		Format:                   "jpeg",
		TitlePage:                1,
		profiles:                 newProfiles(),
	}
}

func (o *converterOptions) Header() string {
	return "Go Comic Converter\n\nOptions:"
}

func (o *converterOptions) String() string {
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

func (o *converterOptions) MarshalJSON() ([]byte, error) {
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
			out["keepdoublepageifsplitted"] = o.KeepDoublePageIfSplitted
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

// Config file: ~/.go-comic-converter.yaml
func (o *converterOptions) FileName() string {
	home, _ := os.UserHomeDir()
	return filepath.Join(home, ".go-comic-converter.yaml")
}

// Load config files
func (o *converterOptions) LoadConfig() error {
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
func (o *converterOptions) ShowConfig() string {
	var profileDesc string
	profile := o.GetProfile()
	if profile != nil {
		profileDesc = fmt.Sprintf(
			"%s - %s - %dx%d",
			o.Profile,
			profile.Description,
			profile.Width,
			profile.Height,
		)
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

	aspectRatio := "auto"
	if o.AspectRatio > 0 {
		aspectRatio = fmt.Sprintf("1:%.02f", o.AspectRatio)
	} else if o.AspectRatio < 0 {
		aspectRatio = fmt.Sprintf("1:%0.2f (device)", float64(profile.Height)/float64(profile.Width))
	}

	titlePage := ""
	switch o.TitlePage {
	case 0:
		titlePage = "never"
	case 1:
		titlePage = "always"
	case 2:
		titlePage = "when epub is splitted"
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
		{"Grayscale Mode", grayscaleMode, o.Grayscale},
		{"Crop", o.Crop, true},
		{"Crop Ratio", fmt.Sprintf("%d Left - %d Up - %d Right - %d Bottom", o.CropRatioLeft, o.CropRatioUp, o.CropRatioRight, o.CropRatioBottom), o.Crop},
		{"Brightness", o.Brightness, o.Brightness != 0},
		{"Contrast", o.Contrast, o.Contrast != 0},
		{"Auto Contrast", o.AutoContrast, true},
		{"Auto Rotate", o.AutoRotate, true},
		{"Auto Split DoublePage", o.AutoSplitDoublePage, o.PortraitOnly || !o.AppleBookCompatibility},
		{"Keep DoublePage If Splitted", o.KeepDoublePageIfSplitted, (o.PortraitOnly || !o.AppleBookCompatibility) && o.AutoSplitDoublePage},
		{"No Blank Image", o.NoBlankImage, true},
		{"Manga", o.Manga, true},
		{"Has Cover", o.HasCover, true},
		{"Limit", fmt.Sprintf("%d Mb", o.LimitMb), o.LimitMb != 0},
		{"Strip First Directory From Toc", o.StripFirstDirectoryFromToc, true},
		{"Sort Path Mode", sortpathmode, true},
		{"Foreground Color", fmt.Sprintf("#%s", o.ForegroundColor), true},
		{"Background Color", fmt.Sprintf("#%s", o.BackgroundColor), true},
		{"Resize", !o.NoResize, true},
		{"Aspect Ratio", aspectRatio, true},
		{"Portrait Only", o.PortraitOnly, true},
		{"Title Page", titlePage, true},
		{"Apple Book Compatibility", o.AppleBookCompatibility, !o.PortraitOnly},
	} {
		if v.Condition {
			b.WriteString(fmt.Sprintf("\n    %-32s: %v", v.Key, v.Value))
		}
	}
	return b.String()
}

// reset all settings to default value
func (o *converterOptions) ResetConfig() error {
	newOptions().SaveConfig()
	return o.LoadConfig()
}

// save all current settings as futur default value
func (o *converterOptions) SaveConfig() error {
	f, err := os.Create(o.FileName())
	if err != nil {
		return err
	}
	defer f.Close()
	return yaml.NewEncoder(f).Encode(o)
}

// shortcut to get current profile
func (o *converterOptions) GetProfile() *converterProfile {
	return o.profiles.Get(o.Profile)
}

// all available profiles
func (o *converterOptions) AvailableProfiles() string {
	return o.profiles.String()
}
