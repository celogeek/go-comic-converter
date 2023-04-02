package profiles

import (
	"fmt"
	"image/color"
	"strings"

	"github.com/celogeek/go-comic-converter/internal/epub"
)

type Profile struct {
	Code        string
	Description string
	Width       int
	Height      int
	Palette     color.Palette
}

type Profiles []Profile

func New() Profiles {
	return []Profile{
		{"KS", "Kindle 1", 600, 670, epub.PALETTE_4},
		{"K11", "Kindle 11", 1072, 1448, epub.PALETTE_16},
		{"K2", "Kindle 2", 600, 670, epub.PALETTE_15},
		{"K34", "Kindle Keyboard/Touch", 600, 800, epub.PALETTE_16},
		{"K578", "Kindle", 600, 800, epub.PALETTE_16},
		{"KDX", "Kindle DX/DXG", 824, 1000, epub.PALETTE_16},
		{"KPW", "Kindle Paperwhite 1/2", 758, 1024, epub.PALETTE_16},
		{"KV", "Kindle Paperwhite 3/4/Voyage/Oasis", 1072, 1448, epub.PALETTE_16},
		{"KPW5", "Kindle Paperwhite 5/Signature Edition", 1236, 1648, epub.PALETTE_16},
		{"KO", "Kindle Oasis 2/3", 1264, 1680, epub.PALETTE_16},
		{"KS", "Kindle Scribe", 1860, 2480, epub.PALETTE_16},
		// Kobo
		{"KoMT", "Kobo Mini/Touch", 600, 800, epub.PALETTE_16},
		{"KoG", "Kobo Glo", 768, 1024, epub.PALETTE_16},
		{"KoGHD", "Kobo Glo HD", 1072, 1448, epub.PALETTE_16},
		{"KoA", "Kobo Aura", 758, 1024, epub.PALETTE_16},
		{"KoAHD", "Kobo Aura HD", 1080, 1440, epub.PALETTE_16},
		{"KoAH2O", "Kobo Aura H2O", 1080, 1430, epub.PALETTE_16},
		{"KoAO", "Kobo Aura ONE", 1404, 1872, epub.PALETTE_16},
		{"KoN", "Kobo Nia", 758, 1024, epub.PALETTE_16},
		{"KoC", "Kobo Clara HD/Kobo Clara 2E", 1072, 1448, epub.PALETTE_16},
		{"KoL", "Kobo Libra H2O/Kobo Libra 2", 1264, 1680, epub.PALETTE_16},
		{"KoF", "Kobo Forma", 1440, 1920, epub.PALETTE_16},
		{"KoS", "Kobo Sage", 1440, 1920, epub.PALETTE_16},
		{"KoE", "Kobo Elipsa", 1404, 1872, epub.PALETTE_16},
	}
}

func (p Profiles) String() string {
	s := make([]string, 0)
	for _, v := range p {
		s = append(s, fmt.Sprintf(
			"    - %-7s ( %9s ) - %2d levels of gray - %s",
			v.Code,
			fmt.Sprintf("%dx%d", v.Width, v.Height),
			len(v.Palette),
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
