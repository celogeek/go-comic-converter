package epuboptions

import "github.com/celogeek/go-comic-converter/v3/internal/pkg/epubimage"

// Limits of the zoom of the stored images, relative to the size of the view.
const (
	MinZoom = 1.0
	MaxZoom = 3.0
)

// DeviceType groups profiles by device family.
type DeviceType string

const (
	Kindle     DeviceType = "Kindle"
	Kobo       DeviceType = "Kobo"
	ReMarkable DeviceType = "reMarkable"
	Apple      DeviceType = "Apple"
)

// Double page modes: a double page (width > height) is never kept untouched.
const (
	DoublePageSplit       = 0 // split in 2 pages
	DoublePageSplitRotate = 1 // rotated double page, then split in 2 pages
	DoublePageRotate      = 2 // rotated double page only
)

type Image struct {
	Crop                      Crop   `yaml:"crop" json:"crop"`
	Quality                   int    `yaml:"quality" json:"quality"`
	Brightness                int    `yaml:"brightness" json:"brightness"`
	Contrast                  int    `yaml:"contrast" json:"contrast"`
	AutoContrast              bool   `yaml:"auto_contrast" json:"auto_contrast"`
	DoublePageMode            int    `yaml:"double_page_mode" json:"double_page_mode"` // 0 = split, 1 = split + rotate, 2 = rotate
	KeepSplitDoublePageAspect bool   `yaml:"keep_split_double_page_aspect" json:"keep_split_double_page_aspect"`
	CombineSplitDoublePage    bool   `yaml:"combine_split_double_page" json:"combine_split_double_page"`
	NoBlankImage              bool   `yaml:"no_blank_image" json:"no_blank_image"`
	Manga                     bool   `yaml:"manga" json:"manga"`
	AddPanelView              bool   `yaml:"-" json:"add_panel_view"`
	View                      View   `yaml:"view" json:"view"`
	GrayScale                 bool   `yaml:"grayscale" json:"grayscale"`
	GrayScaleMode             int    `yaml:"grayscale_mode" json:"gray_scale_mode"` // 0 = normal, 1 = average, 2 = luminance
	Resize                    bool   `yaml:"resize" json:"resize"`
	Format                    string `yaml:"format" json:"format"`
	// Zoom of the stored images: they are bigger than the view, for a better quality when the reader zooms.
	// 0 is the zoom of the profile.
	Zoom float64 `yaml:"zoom" json:"zoom"`
	// ProfileZoom is the zoom of the profile, it comes only from the profile.
	ProfileZoom float64 `yaml:"-" json:"profile_zoom"`
	// Device is the device family of the profile, it comes only from the profile.
	Device DeviceType `yaml:"-" json:"device"`
}

// SplitDoublePage tells if a double page is split in 2 pages.
func (i Image) SplitDoublePage() bool {
	return i.DoublePageMode != DoublePageRotate
}

// RotateDoublePage tells if a double page is rotated.
func (i Image) RotateDoublePage() bool {
	return i.DoublePageMode != DoublePageSplit
}

// KeepDoublePage tells if the double page itself (not only its parts) is part of the output.
func (i Image) KeepDoublePage() bool {
	return i.DoublePageMode != DoublePageSplit
}

// deviceTiles max size of the file an image is stored in, for the devices with a constraint.
var deviceTiles = map[DeviceType]epubimage.TileSize{
	Kindle: epubimage.TileKindle,
	Apple:  epubimage.TileApple,
}

// Tile max size of the file an image is stored in, bigger images are tiled.
//
// Only the devices with a constraint have one, the other devices get no tile (the zero value).
func (i Image) Tile() epubimage.TileSize {
	return deviceTiles[i.Device]
}

// ZoomFactor zoom of the stored images: the option, or the zoom of the profile, or 1.
func (i Image) ZoomFactor() float64 {
	if i.Zoom > 0 {
		return i.Zoom
	}
	if i.ProfileZoom > 0 {
		return i.ProfileZoom
	}
	return 1
}
