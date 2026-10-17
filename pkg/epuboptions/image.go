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

// Reading modes: the direction of the pages, and the webtoon that is a different kind of source.
const (
	ModeComic   = 0 // left to right
	ModeManga   = 1 // right to left, when the profile supports it
	ModeWebtoon = 2 // long vertical strips cut into pages: no double page, no crop, no blank page removal
)

// Features of the device, they drive the default and the allowed options.
type Features struct {
	// Color is true when the screen displays colors: otherwise the grayscale can't be turned off.
	Color bool `json:"color"`
	// PanelView is true when the device has no native panel view: it is added to the pages by default.
	PanelView bool `json:"panel_view"`
	// Manga is true when the device supports the manga mode (right to left), it is not the case of the reMarkable.
	Manga bool `json:"manga"`
	// Landscape is true when the device has the landscape only mode (Kindle Scribe).
	Landscape bool `json:"landscape"`
	// Zoom is the default zoom of the images: they are stored bigger than the view (1 = the size of the view).
	// It is 1 when unset.
	Zoom float64 `json:"zoom"`
}

// Double page modes: a double page (width > height) is never kept untouched.
const (
	DoublePageSplit       = 0 // split in 2 pages
	DoublePageSplitRotate = 1 // rotated double page, then split in 2 pages
	DoublePageRotate      = 2 // rotated double page only
)

type Image struct {
	Crop    Crop `yaml:"crop" json:"crop"`
	Quality int  `yaml:"quality" json:"quality"`
	// JXLEffort is the effort of the jxl encoder, from 1 (fast) to 7 (smaller files).
	JXLEffort                 int  `yaml:"jxl_effort" json:"jxl_effort"`
	Brightness                int  `yaml:"brightness" json:"brightness"`
	Contrast                  int  `yaml:"contrast" json:"contrast"`
	AutoContrast              bool `yaml:"auto_contrast" json:"auto_contrast"`
	DoublePageMode            int  `yaml:"double_page_mode" json:"double_page_mode"` // 0 = split, 1 = split + rotate, 2 = rotate
	KeepSplitDoublePageAspect bool `yaml:"keep_split_double_page_aspect" json:"keep_split_double_page_aspect"`
	CombineSplitDoublePage    bool `yaml:"combine_split_double_page" json:"combine_split_double_page"`
	NoBlankImage              bool `yaml:"no_blank_image" json:"no_blank_image"`
	// Mode is the reading mode: comic, manga or webtoon.
	Mode          int    `yaml:"mode" json:"mode"`
	Landscape     bool   `yaml:"landscape" json:"landscape"`
	View          View   `yaml:"view" json:"view"`
	GrayScale     bool   `yaml:"grayscale" json:"grayscale"`
	GrayScaleMode int    `yaml:"grayscale_mode" json:"gray_scale_mode"` // 0 = normal, 1 = average, 2 = luminance
	Resize        bool   `yaml:"resize" json:"resize"`
	Format        string `yaml:"format" json:"format"`
	// Zoom of the stored images: they are bigger than the view, for a better quality when the reader zooms.
	// 0 is the zoom of the profile.
	Zoom float64 `yaml:"zoom" json:"zoom"`
	// Features are the features of the profile, they come only from the profile.
	Features Features `yaml:"-" json:"features"`
	// Device is the device family of the profile, it comes only from the profile.
	Device DeviceType `yaml:"-" json:"device"`
}

// doublePageMode the double page mode in use: the landscape only mode doesn't support the rotation, the double page is split.
func (i Image) doublePageMode() int {
	if i.LandscapeMode() {
		return DoublePageSplit
	}
	return i.DoublePageMode
}

// SplitDoublePage tells if a double page is split in 2 pages.
func (i Image) SplitDoublePage() bool {
	return i.doublePageMode() != DoublePageRotate
}

// RotateDoublePage tells if a double page is rotated.
func (i Image) RotateDoublePage() bool {
	return i.doublePageMode() != DoublePageSplit
}

// KeepDoublePage tells if the double page itself (not only its parts) is part of the output.
func (i Image) KeepDoublePage() bool {
	return i.doublePageMode() != DoublePageSplit
}

// MangaMode tells if the pages are read from right to left: the option, and a profile that supports it.
func (i Image) MangaMode() bool {
	return i.Mode == ModeManga && i.Features.Manga
}

// WebtoonMode tells if the images are long vertical strips: they are joined, and cut into pages of the size of the view.
//
// The options about the double pages, the crop and the blank pages don't apply.
func (i Image) WebtoonMode() bool {
	return i.Mode == ModeWebtoon
}

// CropPages tells if the margins of the pages are cropped.
func (i Image) CropPages() bool {
	return i.Crop.Enabled && !i.WebtoonMode()
}

// RemoveBlankPages tells if the blank pages are removed.
func (i Image) RemoveBlankPages() bool {
	return i.NoBlankImage && !i.WebtoonMode()
}

// CombineDoublePages tells if the pages that are the 2 halves of a double page are merged.
func (i Image) CombineDoublePages() bool {
	return i.CombineSplitDoublePage && !i.WebtoonMode()
}

// WebtoonAspectRatio aspect ratio (height/width) of the pages a webtoon is cut into: the one of the view when it is set,
// otherwise the one of the device (the pages have no aspect ratio of their own to take the one of the source).
func (i Image) WebtoonAspectRatio() float64 {
	if ar := i.View.AspectRatio; ar > 0 {
		return min(max(ar, MinAspectRatio), MaxAspectRatio)
	}
	if i.View.Width > 0 && i.View.Height > 0 {
		return float64(i.View.Height) / float64(i.View.Width)
	}
	return 4.0 / 3
}

// IsGrayScale tells if the images are in grayscale: the option, or a profile without color screen.
func (i Image) IsGrayScale() bool {
	return i.GrayScale || !i.Features.Color
}

// LandscapeMode tells if the pages are spread left and right, with a blank page to keep them in pairs.
//
// It is only available on the profiles with the feature (Kindle Scribe),
// the other devices balance the pages by themselves.
func (i Image) LandscapeMode() bool {
	return i.Landscape && i.Features.Landscape
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
	if i.Features.Zoom > 0 {
		return i.Features.Zoom
	}
	return 1
}
