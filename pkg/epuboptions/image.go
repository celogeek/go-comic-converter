package epuboptions

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
	AppleBookCompatibility    bool   `yaml:"apple_book_compatibility" json:"apple_book_compatibility"`
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
