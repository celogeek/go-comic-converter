// Package converter profiles manage supported profiles for go-comic-converter.
package converter

import (
	"fmt"
	"strings"

	"github.com/celogeek/go-comic-converter/v3/internal/pkg/utils"
	"github.com/celogeek/go-comic-converter/v3/pkg/epuboptions"
)

type Profile struct {
	Code        string                 `json:"code"`
	Type        epuboptions.DeviceType `json:"type"`
	Year        int                    `json:"year,omitempty"`
	Width       int                    `json:"width"`
	Height      int                    `json:"height"`
	Description string                 `json:"description"`
	Features    Features               `json:"features"`
}

// Features of the device, they drive the default and the allowed options.
type Features struct {
	// Color is true when the screen displays colors: otherwise the grayscale can't be turned off.
	Color bool `json:"color"`
	// PanelView is true when the device has no native panel view: it is added to the pages by default.
	PanelView bool `json:"panel_view"`
	// Zoom is the default zoom of the images: they are stored bigger than the view (1 = the size of the view).
	// It is 1 when unset.
	Zoom float64 `json:"zoom"`
}

func (p Profile) String() string {
	return p.Code + " - " + p.Description + " - " + utils.IntToString(p.Width) + "x" + utils.IntToString(p.Height)
}

type Profiles []Profile

// NewProfiles Initialize list of all supported profiles.
func NewProfiles() Profiles {
	profiles := []Profile{
		// Kindle, the generic resolutions first, then by year of release, newest first
		{"HR", epuboptions.Kindle, 0, 2400, 3840, "Kindle High Resolution", Features{Color: true}},
		{"SR", epuboptions.Kindle, 0, 1200, 1920, "Kindle Standard Resolution", Features{Color: true}},
		{"KS3", epuboptions.Kindle, 2025, 1980, 2640, "Kindle Scribe 3", Features{}},
		{"KSC", epuboptions.Kindle, 2025, 1980, 2640, "Kindle Scribe Colorsoft", Features{Color: true}},
		{"K12", epuboptions.Kindle, 2024, 1072, 1448, "Kindle 12", Features{}},
		{"KPW12", epuboptions.Kindle, 2024, 1264, 1680, "Kindle Paperwhite 12/Signature Edition", Features{}},
		{"KCS", epuboptions.Kindle, 2024, 1264, 1680, "Kindle Colorsoft", Features{Color: true}},
		{"KS2", epuboptions.Kindle, 2024, 1860, 2480, "Kindle Scribe 2", Features{}},
		{"K11", epuboptions.Kindle, 2022, 1072, 1448, "Kindle 11", Features{}},
		{"KS", epuboptions.Kindle, 2022, 1860, 2480, "Kindle Scribe", Features{}},
		{"KPW5", epuboptions.Kindle, 2021, 1236, 1648, "Kindle Paperwhite 5/Signature Edition", Features{PanelView: true}},
		{"K10", epuboptions.Kindle, 2019, 600, 800, "Kindle 10", Features{PanelView: true}},
		{"KO3", epuboptions.Kindle, 2019, 1264, 1680, "Kindle Oasis 3", Features{PanelView: true}},
		{"KPW4", epuboptions.Kindle, 2018, 1072, 1448, "Kindle Paperwhite 4", Features{PanelView: true}},
		{"KO2", epuboptions.Kindle, 2017, 1264, 1680, "Kindle Oasis 2", Features{PanelView: true}},
		{"KO1", epuboptions.Kindle, 2016, 1072, 1448, "Kindle Oasis 1", Features{PanelView: true}},
		{"K8", epuboptions.Kindle, 2016, 600, 800, "Kindle 8", Features{PanelView: true}},
		{"KPW3", epuboptions.Kindle, 2015, 1072, 1448, "Kindle Paperwhite 3", Features{PanelView: true}},
		{"K7", epuboptions.Kindle, 2014, 600, 800, "Kindle 7", Features{PanelView: true}},
		{"KV", epuboptions.Kindle, 2014, 1072, 1448, "Kindle Voyage", Features{PanelView: true}},
		{"KPW2", epuboptions.Kindle, 2013, 758, 1024, "Kindle Paperwhite 2", Features{PanelView: true}},
		{"K5", epuboptions.Kindle, 2012, 600, 800, "Kindle 5", Features{PanelView: true}},
		{"KPW1", epuboptions.Kindle, 2012, 758, 1024, "Kindle Paperwhite 1", Features{PanelView: true}},
		{"K4", epuboptions.Kindle, 2011, 600, 800, "Kindle 4", Features{PanelView: true}},
		{"KT", epuboptions.Kindle, 2011, 600, 800, "Kindle Touch", Features{PanelView: true}},
		{"K3", epuboptions.Kindle, 2010, 600, 800, "Kindle Keyboard", Features{PanelView: true}},
		{"K2", epuboptions.Kindle, 2009, 600, 800, "Kindle 2", Features{PanelView: true}},
		{"KDX", epuboptions.Kindle, 2009, 824, 1200, "Kindle DX/DXG", Features{PanelView: true}},
		{"K1", epuboptions.Kindle, 2007, 600, 800, "Kindle 1", Features{PanelView: true}},
		// Kobo, by year of release, newest first
		{"KoCBW", epuboptions.Kobo, 2024, 1072, 1448, "Kobo Clara BW", Features{}},
		{"KoCC", epuboptions.Kobo, 2024, 1072, 1448, "Kobo Clara Colour", Features{Color: true}},
		{"KoLC", epuboptions.Kobo, 2024, 1264, 1680, "Kobo Libra Colour", Features{Color: true}},
		{"KoE2", epuboptions.Kobo, 2023, 1404, 1872, "Kobo Elipsa 2E", Features{}},
		{"KoC2E", epuboptions.Kobo, 2022, 1072, 1448, "Kobo Clara 2E", Features{}},
		{"KoE", epuboptions.Kobo, 2021, 1404, 1872, "Kobo Elipsa", Features{}},
		{"KoS", epuboptions.Kobo, 2021, 1440, 1920, "Kobo Sage", Features{}},
		{"KoL2", epuboptions.Kobo, 2021, 1264, 1680, "Kobo Libra 2", Features{}},
		{"KoN", epuboptions.Kobo, 2020, 758, 1024, "Kobo Nia", Features{}},
		{"KoLH2O", epuboptions.Kobo, 2019, 1264, 1680, "Kobo Libra H2O", Features{}},
		{"KoCHD", epuboptions.Kobo, 2018, 1072, 1448, "Kobo Clara HD", Features{}},
		{"KoF", epuboptions.Kobo, 2018, 1440, 1920, "Kobo Forma", Features{}},
		{"KoAH2O2", epuboptions.Kobo, 2017, 1080, 1430, "Kobo Aura H2O Edition 2", Features{}},
		{"KoAO", epuboptions.Kobo, 2016, 1404, 1872, "Kobo Aura One", Features{}},
		{"KoA2", epuboptions.Kobo, 2016, 758, 1024, "Kobo Aura Edition 2", Features{}},
		{"KoGHD", epuboptions.Kobo, 2015, 1072, 1448, "Kobo Glo HD", Features{}},
		{"KoT2", epuboptions.Kobo, 2015, 600, 800, "Kobo Touch 2.0", Features{}},
		{"KoAH2O", epuboptions.Kobo, 2014, 1080, 1430, "Kobo Aura H2O", Features{}},
		{"KoA", epuboptions.Kobo, 2013, 758, 1024, "Kobo Aura", Features{}},
		{"KoAHD", epuboptions.Kobo, 2013, 1080, 1440, "Kobo Aura HD", Features{}},
		{"KoM", epuboptions.Kobo, 2012, 600, 800, "Kobo Mini", Features{}},
		{"KoG", epuboptions.Kobo, 2012, 758, 1024, "Kobo Glo", Features{}},
		{"KoT", epuboptions.Kobo, 2011, 600, 800, "Kobo Touch", Features{}},
		{"KoR", epuboptions.Kobo, 2010, 600, 800, "Kobo eReader", Features{}},
		// reMarkable
		{"RMPU", epuboptions.ReMarkable, 2026, 1404, 1872, "reMarkable Paper Pure", Features{}},
		{"RMPM", epuboptions.ReMarkable, 2025, 954, 1696, "reMarkable Paper Pro Move", Features{Color: true}},
		{"RMP", epuboptions.ReMarkable, 2024, 1620, 2160, "reMarkable Paper Pro", Features{Color: true}},
		{"RM2", epuboptions.ReMarkable, 2020, 1404, 1872, "reMarkable 2", Features{}},
		{"RM1", epuboptions.ReMarkable, 2017, 1404, 1872, "reMarkable 1", Features{}},
		// Apple, the standard resolution first, then by year of the latest device of the profile, newest first. The panel view is handled by Apple Books
		{"AB", epuboptions.Apple, 0, 1640, 2360, "Apple Books Standard Resolution", Features{Color: true, Zoom: 1.5}},
		// TODO: provisional size taken from the leaks of the inner display (2713 x 1920), Apple has not published
		// the resolution of the iPhone Duo yet (on sale 2026-10-23). To update once it is released.
		{"ABD", epuboptions.Apple, 2026, 1920, 2713, "Apple Books on iPhone Duo", Features{Color: true, Zoom: 1.5}},
		{"ABP11", epuboptions.Apple, 2025, 1668, 2420, "Apple Books on iPad Pro 11\"", Features{Color: true, Zoom: 1.5}},
		{"ABP13", epuboptions.Apple, 2025, 2064, 2752, "Apple Books on iPad Pro 13\"", Features{Color: true, Zoom: 1.5}},
		{"ABI", epuboptions.Apple, 2025, 1320, 2868, "Apple Books on iPhone", Features{Color: true, Zoom: 1.5}},
		{"ABM", epuboptions.Apple, 2024, 1488, 2266, "Apple Books on iPad mini", Features{Color: true, Zoom: 1.5}},
	}

	for i := range profiles {
		if profiles[i].Features.Zoom == 0 {
			profiles[i].Features.Zoom = 1
		}
	}

	return profiles
}

func (p Profiles) String() string {
	s := make([]string, 0)
	var last epuboptions.DeviceType
	for _, v := range p {
		if v.Type != last {
			s = append(s, "    "+string(v.Type)+":")
			last = v.Type
		}
		year := ""
		if v.Year > 0 {
			year = " (" + utils.IntToString(v.Year) + ")"
		}
		s = append(s, fmt.Sprintf(
			"      - %-7s %7s - %4d x %-4d - x%-4s - %s",
			v.Code,
			year,
			v.Width, v.Height,
			utils.FloatToString(v.Features.Zoom, -1),
			v.Description,
		))
	}
	return strings.Join(s, "\n")
}
