package epubimage

// options for image transformation
type Options struct {
	Crop                bool
	CropRatioLeft       int
	CropRatioUp         int
	CropRatioRight      int
	CropRatioBottom     int
	ViewWidth           int
	ViewHeight          int
	Quality             int
	Brightness          int
	Contrast            int
	AutoRotate          bool
	AutoSplitDoublePage bool
	NoBlankPage         bool
	Manga               bool
	HasCover            bool
}
