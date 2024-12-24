package epuboptions

type Image struct {
	Crop                      Crop   `yaml:"crop" json:"crop"`
	Quality                   int    `yaml:"quality" json:"quality"`
	Brightness                int    `yaml:"brightness" json:"brightness"`
	Contrast                  int    `yaml:"contrast" json:"contrast"`
	AutoContrast              bool   `yaml:"auto_contrast" json:"auto_contrast"`
	AutoRotate                bool   `yaml:"auto_rotate" json:"auto_rotate"`
	AutoSplitDoublePage       bool   `yaml:"auto_split_double_page" json:"auto_split_double_page"`
	KeepDoublePageIfSplit     bool   `yaml:"keep_double_page_if_split" json:"keep_double_page_if_split"`
	KeepSplitDoublePageAspect bool   `yaml:"keep_split_double_page_aspect" json:"keep_split_double_page_aspect"`
	NoBlankImage              bool   `yaml:"no_blank_image" json:"no_blank_image"`
	Manga                     bool   `yaml:"manga" json:"manga"`
	HasCover                  bool   `yaml:"has_cover" json:"has_cover"`
	View                      View   `yaml:"view" json:"view"`
	GrayScale                 bool   `yaml:"grayscale" json:"grayscale"`
	GrayScaleMode             int    `yaml:"grayscale_mode" json:"gray_scale_mode"` // 0 = normal, 1 = average, 2 = luminance
	Resize                    bool   `yaml:"resize" json:"resize"`
	Format                    string `yaml:"format" json:"format"`
	AppleBookCompatibility    bool   `yaml:"apple_book_compatibility" json:"apple_book_compatibility"`
}

func (i Image) MediaType() string {
	return "image/" + i.Format
}
