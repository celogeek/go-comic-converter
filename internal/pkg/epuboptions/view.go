package epuboptions

import (
	"github.com/celogeek/go-comic-converter/v3/internal/pkg/utils"
)

type View struct {
	Width        int     `yaml:"-" json:"width"`
	Height       int     `yaml:"-" json:"height"`
	AspectRatio  float64 `yaml:"aspect_ratio" json:"aspect_ratio"`
	PortraitOnly bool    `yaml:"portrait_only" json:"portrait_only"`
	Color        Color   `yaml:"color" json:"color"`
}

func (v View) Dimension() string {
	return utils.IntToString(v.Width) + "x" + utils.IntToString(v.Height)
}

func (v View) Port() string {
	return "width=" + utils.IntToString(v.Width) + ",height=" + utils.IntToString(v.Height)
}
