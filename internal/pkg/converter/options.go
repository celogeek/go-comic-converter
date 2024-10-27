// Package converter options manage options with default value from config.
package converter

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"gopkg.in/yaml.v3"

	"github.com/celogeek/go-comic-converter/v2/internal/pkg/epuboptions"
	"github.com/celogeek/go-comic-converter/v2/internal/pkg/utils"
)

type Options struct {
	epuboptions.EPUBOptions

	// Config
	Profile                string  `yaml:"profile"`
	NoResize               bool    `yaml:"noresize"`
	Format                 string  `yaml:"format"`
	AspectRatio            float64 `yaml:"aspect_ratio"`
	PortraitOnly           bool    `yaml:"portrait_only"`
	AppleBookCompatibility bool    `yaml:"apple_book_compatibility"`

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
	Version bool `yaml:"-"`
	Help    bool `yaml:"-"`

	// Internal
	profiles Profiles
}

// NewOptions Initialize default options.
func NewOptions() *Options {
	return &Options{
		Profile: "SR",
		EPUBOptions: epuboptions.EPUBOptions{
			Image: epuboptions.Image{
				Quality:   85,
				GrayScale: true,
				Crop: epuboptions.Crop{
					Enabled: true,
					Left:    1,
					Up:      1,
					Right:   1,
					Bottom:  3,
				},
				NoBlankImage:              true,
				HasCover:                  true,
				KeepDoublePageIfSplit:     true,
				KeepSplitDoublePageAspect: true,
				View: epuboptions.View{
					Color: epuboptions.Color{
						Foreground: "000",
						Background: "FFF",
					},
				},
			},
			TitlePage:    1,
			SortPathMode: 1,
		},
		Format:   "jpeg",
		profiles: NewProfiles(),
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
		"grayscale":                  o.Image.GrayScale,
		"crop":                       o.Image.Crop.Enabled,
		"autocontrast":               o.Image.AutoContrast,
		"autorotate":                 o.Image.AutoRotate,
		"noblankimage":               o.Image.NoBlankImage,
		"manga":                      o.Image.Manga,
		"hascover":                   o.Image.HasCover,
		"stripfirstdirectoryfromtoc": o.StripFirstDirectoryFromToc,
		"sortpathmode":               o.SortPathMode,
		"foregroundcolor":            o.Image.View.Color.Foreground,
		"backgroundcolor":            o.Image.View.Color.Background,
		"resize":                     !o.NoResize,
		"aspectratio":                o.AspectRatio,
		"portraitonly":               o.PortraitOnly,
		"titlepage":                  o.TitlePage,
	}
	if o.Format == "jpeg" {
		out["quality"] = o.Image.Quality
	}
	if o.Image.GrayScale {
		out["grayscale_mode"] = o.Image.GrayScaleMode
	}
	if o.Image.Crop.Enabled {
		out["crop_ratio"] = map[string]any{
			"left":   o.Image.Crop.Left,
			"right":  o.Image.Crop.Right,
			"up":     o.Image.Crop.Up,
			"bottom": o.Image.Crop.Bottom,
		}
		out["crop_limit"] = o.Image.Crop.Limit
		out["crop_skip_if_limit_reached"] = o.Image.Crop.SkipIfLimitReached
	}
	if o.Image.Brightness != 0 {
		out["brightness"] = o.Image.Brightness
	}
	if o.Image.Contrast != 0 {
		out["contrast"] = o.Image.Contrast
	}
	if o.PortraitOnly || !o.AppleBookCompatibility {
		out["autosplitdoublepage"] = o.Image.AutoSplitDoublePage
		if o.Image.AutoSplitDoublePage {
			out["keepdoublepageifsplit"] = o.Image.KeepDoublePageIfSplit
			out["keepsplitdoublepageaspect"] = o.Image.KeepSplitDoublePageAspect
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
	switch o.Image.GrayScaleMode {
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
		{"Quality", o.Image.Quality, o.Format == "jpeg"},
		{"Grayscale", o.Image.GrayScale, true},
		{"Grayscale mode", grayscaleMode, o.Image.GrayScale},
		{"Crop", o.Image.Crop.Enabled, true},
		{"Crop ratio",
			utils.IntToString(o.Image.Crop.Left) + " Left - " +
				utils.IntToString(o.Image.Crop.Up) + " Up - " +
				utils.IntToString(o.Image.Crop.Right) + " Right - " +
				utils.IntToString(o.Image.Crop.Bottom) + " Bottom - " +
				"Limit " + utils.IntToString(o.Image.Crop.Limit) + "% - " +
				"Skip " + utils.BoolToString(o.Image.Crop.SkipIfLimitReached),
			o.Image.Crop.Enabled},
		{"Brightness", o.Image.Brightness, o.Image.Brightness != 0},
		{"Contrast", o.Image.Contrast, o.Image.Contrast != 0},
		{"Auto contrast", o.Image.AutoContrast, true},
		{"Auto rotate", o.Image.AutoRotate, true},
		{"Auto split double page", o.Image.AutoSplitDoublePage, o.PortraitOnly || !o.AppleBookCompatibility},
		{"Keep double page if split", o.Image.KeepDoublePageIfSplit, (o.PortraitOnly || !o.AppleBookCompatibility) && o.Image.AutoSplitDoublePage},
		{"Keep split double page aspect", o.Image.KeepSplitDoublePageAspect, (o.PortraitOnly || !o.AppleBookCompatibility) && o.Image.AutoSplitDoublePage},
		{"No blank image", o.Image.NoBlankImage, true},
		{"Manga", o.Image.Manga, true},
		{"Has cover", o.Image.HasCover, true},
		{"Limit", utils.IntToString(o.LimitMb) + " Mb", o.LimitMb != 0},
		{"Strip first directory from toc", o.StripFirstDirectoryFromToc, true},
		{"Sort path mode", sortpathmode, true},
		{"Foreground color", "#" + o.Image.View.Color.Foreground, true},
		{"Background color", "#" + o.Image.View.Color.Background, true},
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
