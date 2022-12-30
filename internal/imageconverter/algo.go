package imageconverter

import (
	"image/color"
	"sort"
)

var ALGO_GRAY = map[string]func(color.Color, color.Palette) color.Gray{
	"default": func(c color.Color, p color.Palette) color.Gray {
		return p.Convert(c).(color.Gray)
	},
	"mean": func(c color.Color, p color.Palette) color.Gray {
		r, g, b, _ := c.RGBA()
		y := float64(r+g+b) / 3
		return p.Convert(color.Gray16{Y: uint16(y)}).(color.Gray)
	},
	"luma": func(c color.Color, p color.Palette) color.Gray {
		r, g, b, _ := c.RGBA()
		y := (0.2126*float64(r) + 0.7152*float64(g) + 0.0722*float64(b))
		return p.Convert(color.Gray16{Y: uint16(y)}).(color.Gray)
	},
	"luster": func(c color.Color, p color.Palette) color.Gray {
		r, g, b, _ := c.RGBA()
		arr := []float64{float64(r), float64(g), float64(b)}
		sort.Float64s(arr)
		y := (arr[0] + arr[2]) / 2
		return p.Convert(color.Gray16{Y: uint16(y)}).(color.Gray)
	},
}
