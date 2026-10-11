package epuboptions

import (
	"github.com/celogeek/go-comic-converter/v3/internal/pkg/utils"
)

// Limits of the aspect ratio (height/width) of the view: the profiles go from 1.1 to 1.8,
// anything outside gives a page too thin to read.
const (
	MinAspectRatio = 0.5
	MaxAspectRatio = 3.0
)

type View struct {
	Width       int     `yaml:"-" json:"width"`
	Height      int     `yaml:"-" json:"height"`
	AspectRatio float64 `yaml:"aspect_ratio" json:"aspect_ratio"`
	Color       Color   `yaml:"color" json:"color"`
}

func (v View) Dimension() string {
	return utils.IntToString(v.Width) + "x" + utils.IntToString(v.Height)
}

func (v View) Port() string {
	return "width=" + utils.IntToString(v.Width) + ",height=" + utils.IntToString(v.Height)
}

func (v View) Style() string {
	return "width:" + utils.IntToString(v.Width) + "px;height:" + utils.IntToString(v.Height) + "px;"
}
