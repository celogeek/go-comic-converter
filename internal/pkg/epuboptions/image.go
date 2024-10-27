package epuboptions

type Image struct {
	Crop                      Crop   `yaml:"crop"`
	Quality                   int    `yaml:"quality"`
	Brightness                int    `yaml:"brightness"`
	Contrast                  int    `yaml:"contrast"`
	AutoContrast              bool   `yaml:"auto_contrast"`
	AutoRotate                bool   `yaml:"auto_rotate"`
	AutoSplitDoublePage       bool   `yaml:"auto_split_double_page"`
	KeepDoublePageIfSplit     bool   `yaml:"keep_double_page_if_split"`
	KeepSplitDoublePageAspect bool   `yaml:"keep_split_double_page_aspect"`
	NoBlankImage              bool   `yaml:"no_blank_image"`
	Manga                     bool   `yaml:"manga"`
	HasCover                  bool   `yaml:"has_cover"`
	View                      View   `yaml:"view"`
	GrayScale                 bool   `yaml:"grayscale"`
	GrayScaleMode             int    `yaml:"grayscale_mode"` // 0 = normal, 1 = average, 2 = luminance
	Resize                    bool   `yaml:"resize"`
	Format                    string `yaml:"format"`
	AppleBookCompatibility    bool   `yaml:"apple_book_compatibility"`
}

func (i Image) MediaType() string {
	return "image/" + i.Format
}
