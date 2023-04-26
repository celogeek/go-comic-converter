package epubimage

// options for image transformation
type Options struct {
	Crop                bool
	CropRatio           int
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
