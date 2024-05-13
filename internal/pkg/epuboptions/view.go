package epuboptions

import (
	"github.com/celogeek/go-comic-converter/v2/internal/pkg/utils"
)

type View struct {
	Width, Height int
	AspectRatio   float64
	PortraitOnly  bool
	Color         Color
}

func (v View) Dimension() string {
	return utils.IntToString(v.Width) + "x" + utils.IntToString(v.Height)
}

func (v View) Port() string {
	return "width=" + utils.IntToString(v.Width) + ",height=" + utils.IntToString(v.Height)
}
