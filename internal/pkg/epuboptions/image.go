package epuboptions

type Image struct {
	Crop                      Crop
	Quality                   int
	Brightness                int
	Contrast                  int
	AutoContrast              bool
	AutoRotate                bool
	AutoSplitDoublePage       bool
	KeepDoublePageIfSplit     bool
	KeepSplitDoublePageAspect bool
	NoBlankImage              bool
	Manga                     bool
	HasCover                  bool
	View                      View
	GrayScale                 bool
	GrayScaleMode             int
	Resize                    bool
	Format                    string
	AppleBookCompatibility    bool
}
