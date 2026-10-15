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
	Features    epuboptions.Features   `json:"features"`
}

func (p Profile) String() string {
	return p.Code + " - " + p.Description + " - " + utils.IntToString(p.Width) + "x" + utils.IntToString(p.Height)
}

type Profiles []Profile

// NewProfiles Initialize list of all supported profiles.
func NewProfiles() Profiles {
	profiles := []Profile{
		// Kindle, the generic resolutions first, then by year of release, newest first
		{"HR", epuboptions.Kindle, 0, 2400, 3840, "Kindle High Resolution", epuboptions.Features{Manga: true, Color: true}},
		{"SR", epuboptions.Kindle, 0, 1200, 1920, "Kindle Standard Resolution", epuboptions.Features{Manga: true, Color: true}},
		{"KS3", epuboptions.Kindle, 2025, 1980, 2640, "Kindle Scribe 3", epuboptions.Features{Manga: true, Landscape: true}},
		{"KSC", epuboptions.Kindle, 2025, 1980, 2640, "Kindle Scribe Colorsoft", epuboptions.Features{Manga: true, Color: true, Landscape: true}},
		{"K12", epuboptions.Kindle, 2024, 1072, 1448, "Kindle 12", epuboptions.Features{Manga: true}},
		{"KPW12", epuboptions.Kindle, 2024, 1264, 1680, "Kindle Paperwhite 12/Signature Edition", epuboptions.Features{Manga: true}},
		{"KCS", epuboptions.Kindle, 2024, 1264, 1680, "Kindle Colorsoft", epuboptions.Features{Manga: true, Color: true}},
		{"KS2", epuboptions.Kindle, 2024, 1860, 2480, "Kindle Scribe 2", epuboptions.Features{Manga: true, Landscape: true}},
		{"K11", epuboptions.Kindle, 2022, 1072, 1448, "Kindle 11", epuboptions.Features{Manga: true}},
		{"KS", epuboptions.Kindle, 2022, 1860, 2480, "Kindle Scribe", epuboptions.Features{Manga: true, Landscape: true}},
		{"KPW5", epuboptions.Kindle, 2021, 1236, 1648, "Kindle Paperwhite 5/Signature Edition", epuboptions.Features{Manga: true, PanelView: true}},
		{"K10", epuboptions.Kindle, 2019, 600, 800, "Kindle 10", epuboptions.Features{Manga: true, PanelView: true}},
		{"KO3", epuboptions.Kindle, 2019, 1264, 1680, "Kindle Oasis 3", epuboptions.Features{Manga: true, PanelView: true}},
		{"KPW4", epuboptions.Kindle, 2018, 1072, 1448, "Kindle Paperwhite 4", epuboptions.Features{Manga: true, PanelView: true}},
		{"KO2", epuboptions.Kindle, 2017, 1264, 1680, "Kindle Oasis 2", epuboptions.Features{Manga: true, PanelView: true}},
		{"KO1", epuboptions.Kindle, 2016, 1072, 1448, "Kindle Oasis 1", epuboptions.Features{Manga: true, PanelView: true}},
		{"K8", epuboptions.Kindle, 2016, 600, 800, "Kindle 8", epuboptions.Features{Manga: true, PanelView: true}},
		{"KPW3", epuboptions.Kindle, 2015, 1072, 1448, "Kindle Paperwhite 3", epuboptions.Features{Manga: true, PanelView: true}},
		{"K7", epuboptions.Kindle, 2014, 600, 800, "Kindle 7", epuboptions.Features{Manga: true, PanelView: true}},
		{"KV", epuboptions.Kindle, 2014, 1072, 1448, "Kindle Voyage", epuboptions.Features{Manga: true, PanelView: true}},
		{"KPW2", epuboptions.Kindle, 2013, 758, 1024, "Kindle Paperwhite 2", epuboptions.Features{Manga: true, PanelView: true}},
		{"K5", epuboptions.Kindle, 2012, 600, 800, "Kindle 5", epuboptions.Features{Manga: true, PanelView: true}},
		{"KPW1", epuboptions.Kindle, 2012, 758, 1024, "Kindle Paperwhite 1", epuboptions.Features{Manga: true, PanelView: true}},
		{"K4", epuboptions.Kindle, 2011, 600, 800, "Kindle 4", epuboptions.Features{Manga: true, PanelView: true}},
		{"KT", epuboptions.Kindle, 2011, 600, 800, "Kindle Touch", epuboptions.Features{Manga: true, PanelView: true}},
		{"K3", epuboptions.Kindle, 2010, 600, 800, "Kindle Keyboard", epuboptions.Features{Manga: true, PanelView: true}},
		{"K2", epuboptions.Kindle, 2009, 600, 800, "Kindle 2", epuboptions.Features{Manga: true, PanelView: true}},
		{"KDX", epuboptions.Kindle, 2009, 824, 1200, "Kindle DX/DXG", epuboptions.Features{Manga: true, PanelView: true}},
		{"K1", epuboptions.Kindle, 2007, 600, 800, "Kindle 1", epuboptions.Features{Manga: true, PanelView: true}},
		// Kobo, by year of release, newest first
		{"KoCBW", epuboptions.Kobo, 2024, 1072, 1448, "Kobo Clara BW", epuboptions.Features{Manga: true}},
		{"KoCC", epuboptions.Kobo, 2024, 1072, 1448, "Kobo Clara Colour", epuboptions.Features{Manga: true, Color: true}},
		{"KoLC", epuboptions.Kobo, 2024, 1264, 1680, "Kobo Libra Colour", epuboptions.Features{Manga: true, Color: true}},
		{"KoE2", epuboptions.Kobo, 2023, 1404, 1872, "Kobo Elipsa 2E", epuboptions.Features{Manga: true}},
		{"KoC2E", epuboptions.Kobo, 2022, 1072, 1448, "Kobo Clara 2E", epuboptions.Features{Manga: true}},
		{"KoE", epuboptions.Kobo, 2021, 1404, 1872, "Kobo Elipsa", epuboptions.Features{Manga: true}},
		{"KoS", epuboptions.Kobo, 2021, 1440, 1920, "Kobo Sage", epuboptions.Features{Manga: true}},
		{"KoL2", epuboptions.Kobo, 2021, 1264, 1680, "Kobo Libra 2", epuboptions.Features{Manga: true}},
		{"KoN", epuboptions.Kobo, 2020, 758, 1024, "Kobo Nia", epuboptions.Features{Manga: true}},
		{"KoLH2O", epuboptions.Kobo, 2019, 1264, 1680, "Kobo Libra H2O", epuboptions.Features{Manga: true}},
		{"KoCHD", epuboptions.Kobo, 2018, 1072, 1448, "Kobo Clara HD", epuboptions.Features{Manga: true}},
		{"KoF", epuboptions.Kobo, 2018, 1440, 1920, "Kobo Forma", epuboptions.Features{Manga: true}},
		{"KoAH2O2", epuboptions.Kobo, 2017, 1080, 1430, "Kobo Aura H2O Edition 2", epuboptions.Features{Manga: true}},
		{"KoAO", epuboptions.Kobo, 2016, 1404, 1872, "Kobo Aura One", epuboptions.Features{Manga: true}},
		{"KoA2", epuboptions.Kobo, 2016, 758, 1024, "Kobo Aura Edition 2", epuboptions.Features{Manga: true}},
		{"KoGHD", epuboptions.Kobo, 2015, 1072, 1448, "Kobo Glo HD", epuboptions.Features{Manga: true}},
		{"KoT2", epuboptions.Kobo, 2015, 600, 800, "Kobo Touch 2.0", epuboptions.Features{Manga: true}},
		{"KoAH2O", epuboptions.Kobo, 2014, 1080, 1430, "Kobo Aura H2O", epuboptions.Features{Manga: true}},
		{"KoA", epuboptions.Kobo, 2013, 758, 1024, "Kobo Aura", epuboptions.Features{Manga: true}},
		{"KoAHD", epuboptions.Kobo, 2013, 1080, 1440, "Kobo Aura HD", epuboptions.Features{Manga: true}},
		{"KoM", epuboptions.Kobo, 2012, 600, 800, "Kobo Mini", epuboptions.Features{Manga: true}},
		{"KoG", epuboptions.Kobo, 2012, 758, 1024, "Kobo Glo", epuboptions.Features{Manga: true}},
		{"KoT", epuboptions.Kobo, 2011, 600, 800, "Kobo Touch", epuboptions.Features{Manga: true}},
		{"KoR", epuboptions.Kobo, 2010, 600, 800, "Kobo eReader", epuboptions.Features{Manga: true}},
		// reMarkable
		{"RMPU", epuboptions.ReMarkable, 2026, 1404, 1872, "reMarkable Paper Pure", epuboptions.Features{}},
		{"RMPM", epuboptions.ReMarkable, 2025, 954, 1696, "reMarkable Paper Pro Move", epuboptions.Features{Color: true}},
		{"RMP", epuboptions.ReMarkable, 2024, 1620, 2160, "reMarkable Paper Pro", epuboptions.Features{Color: true}},
		{"RM2", epuboptions.ReMarkable, 2020, 1404, 1872, "reMarkable 2", epuboptions.Features{}},
		{"RM1", epuboptions.ReMarkable, 2017, 1404, 1872, "reMarkable 1", epuboptions.Features{}},
		// Apple, the standard resolution first, then by year of the latest device of the profile, newest first. The panel view is handled by Apple Books
		{"AB", epuboptions.Apple, 0, 1640, 2360, "Apple Books Standard Resolution", epuboptions.Features{Manga: true, Color: true, Zoom: 1.5}},
		// TODO: provisional size taken from the leaks of the inner display (2713 x 1920), Apple has not published
		// the resolution of the iPhone Duo yet (on sale 2026-10-23). To update once it is released.
		{"ABD", epuboptions.Apple, 2026, 1920, 2713, "Apple Books on iPhone Duo", epuboptions.Features{Manga: true, Color: true, Zoom: 1.5}},
		{"ABP11", epuboptions.Apple, 2025, 1668, 2420, "Apple Books on iPad Pro 11\"", epuboptions.Features{Manga: true, Color: true, Zoom: 1.5}},
		{"ABP13", epuboptions.Apple, 2025, 2064, 2752, "Apple Books on iPad Pro 13\"", epuboptions.Features{Manga: true, Color: true, Zoom: 1.5}},
		{"ABI", epuboptions.Apple, 2025, 1320, 2868, "Apple Books on iPhone", epuboptions.Features{Manga: true, Color: true, Zoom: 1.5}},
		{"ABM", epuboptions.Apple, 2024, 1488, 2266, "Apple Books on iPad mini", epuboptions.Features{Manga: true, Color: true, Zoom: 1.5}},
	}

	for i := range profiles {
		if profiles[i].Features.Zoom == 0 {
			profiles[i].Features.Zoom = 1
		}
	}

	return profiles
}

// String table of the profiles: a row per profile, with the features of the device.
func (p Profiles) String() string {
	rows := make([][]string, 0, len(p)+2)
	rows = append(rows, []string{"Device", "Code", "Year", "Resolution", "Zoom", "Color", "Panel view", "Manga", "Landscape", "Description"})
	rows = append(rows, nil) // separator
	for _, v := range p {
		year := ""
		if v.Year > 0 {
			year = utils.IntToString(v.Year)
		}
		rows = append(rows, []string{
			string(v.Type),
			v.Code,
			year,
			fmt.Sprintf("%4d x %-4d", v.Width, v.Height),
			"x" + utils.FloatToString(v.Features.Zoom, -1),
			yesOrNot(v.Features.Color),
			yesOrNot(v.Features.PanelView),
			yesOrNot(v.Features.Manga),
			yesOrNot(v.Features.Landscape),
			v.Description,
		})
	}

	widths := make([]int, len(rows[0]))
	for _, r := range rows {
		for i, c := range r {
			widths[i] = max(widths[i], len(c))
		}
	}

	lines := make([]string, 0, len(rows))
	for _, r := range rows {
		cells := make([]string, len(widths))
		for i := range widths {
			switch {
			case r == nil:
				cells[i] = strings.Repeat("-", widths[i])
			case i == len(widths)-1:
				cells[i] = r[i]
			default:
				cells[i] = fmt.Sprintf("%-*s", widths[i], r[i])
			}
		}
		lines = append(lines, strings.Join(cells, "  "))
	}

	return strings.Join(lines, "\n")
}

func yesOrNot(b bool) string {
	if b {
		return "yes"
	}
	return ""
}
