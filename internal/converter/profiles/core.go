package profiles

import (
	"fmt"
	"strings"
)

type Profile struct {
	Code        string
	Description string
	Width       int
	Height      int
}

const perfectRatio = 1.5

func (p Profile) PerfectDim() (int, int) {
	width, height := float64(p.Width), float64(p.Height)
	perfectWidth, perfectHeight := height/perfectRatio, width*perfectRatio
	if perfectWidth > width {
		perfectWidth = width
	} else {
		perfectHeight = height
	}
	return int(perfectWidth), int(perfectHeight)
}

type Profiles []Profile

func New() Profiles {
	return []Profile{
		{"K1", "Kindle 1", 600, 670},
		{"K11", "Kindle 11", 1072, 1448},
		{"K2", "Kindle 2", 600, 670},
		{"K34", "Kindle Keyboard/Touch", 600, 800},
		{"K578", "Kindle", 600, 800},
		{"KDX", "Kindle DX/DXG", 824, 1000},
		{"KPW", "Kindle Paperwhite 1/2", 758, 1024},
		{"KV", "Kindle Paperwhite 3/4/Voyage/Oasis", 1072, 1448},
		{"KPW5", "Kindle Paperwhite 5/Signature Edition", 1236, 1648},
		{"KO", "Kindle Oasis 2/3", 1264, 1680},
		{"KS", "Kindle Scribe", 1860, 2480},
		// Kobo
		{"KoMT", "Kobo Mini/Touch", 600, 800},
		{"KoG", "Kobo Glo", 768, 1024},
		{"KoGHD", "Kobo Glo HD", 1072, 1448},
		{"KoA", "Kobo Aura", 758, 1024},
		{"KoAHD", "Kobo Aura HD", 1080, 1440},
		{"KoAH2O", "Kobo Aura H2O", 1080, 1430},
		{"KoAO", "Kobo Aura ONE", 1404, 1872},
		{"KoN", "Kobo Nia", 758, 1024},
		{"KoC", "Kobo Clara HD/Kobo Clara 2E", 1072, 1448},
		{"KoL", "Kobo Libra H2O/Kobo Libra 2", 1264, 1680},
		{"KoF", "Kobo Forma", 1440, 1920},
		{"KoS", "Kobo Sage", 1440, 1920},
		{"KoE", "Kobo Elipsa", 1404, 1872},
	}
}

func (p Profiles) String() string {
	s := make([]string, 0)
	for _, v := range p {
		s = append(s, fmt.Sprintf(
			"    - %-7s ( %9s ) - %s",
			v.Code,
			fmt.Sprintf("%dx%d", v.Width, v.Height),
			v.Description,
		))
	}
	return strings.Join(s, "\n")
}

func (p Profiles) Get(name string) *Profile {
	for _, profile := range p {
		if profile.Code == name {
			return &profile
		}
	}
	return nil
}
