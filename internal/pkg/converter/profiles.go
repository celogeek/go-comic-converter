// Package converter profiles manage supported profiles for go-comic-converter.
package converter

import (
	"fmt"
	"strings"

	"github.com/celogeek/go-comic-converter/v3/internal/pkg/utils"
)

// DeviceType groups profiles by device family.
type DeviceType string

const (
	Generic    DeviceType = "Generic"
	Kindle     DeviceType = "Kindle"
	Kobo       DeviceType = "Kobo"
	ReMarkable DeviceType = "reMarkable"
)

type Profile struct {
	Code        string     `json:"code"`
	Type        DeviceType `json:"type"`
	Year        int        `json:"year,omitempty"`
	Width       int        `json:"width"`
	Height      int        `json:"height"`
	Description string     `json:"description"`
	Features    Features   `json:"features"`
}

// Features of the device, they drive the default and the allowed options.
type Features struct {
	// Color is true when the screen displays colors: otherwise the grayscale can't be turned off.
	Color bool `json:"color"`
	// PanelView is true when the device has no native panel view: it is added to the pages by default.
	PanelView bool `json:"panel_view"`
}

func (p Profile) String() string {
	return p.Code + " - " + p.Description + " - " + utils.IntToString(p.Width) + "x" + utils.IntToString(p.Height)
}

type Profiles []Profile

// NewProfiles Initialize list of all supported profiles.
func NewProfiles() Profiles {
	return []Profile{
		// High Resolution for Tablet
		{"HR", Generic, 0, 2400, 3840, "High Resolution", Features{Color: true}},
		{"SR", Generic, 0, 1200, 1920, "Standard Resolution", Features{Color: true}},
		// Kindle, by year of release
		{"K1", Kindle, 2007, 600, 800, "Kindle 1", Features{PanelView: true}},
		{"K2", Kindle, 2009, 600, 800, "Kindle 2", Features{PanelView: true}},
		{"KDX", Kindle, 2009, 824, 1200, "Kindle DX/DXG", Features{PanelView: true}},
		{"K3", Kindle, 2010, 600, 800, "Kindle Keyboard", Features{PanelView: true}},
		{"K4", Kindle, 2011, 600, 800, "Kindle 4", Features{PanelView: true}},
		{"KT", Kindle, 2011, 600, 800, "Kindle Touch", Features{PanelView: true}},
		{"K5", Kindle, 2012, 600, 800, "Kindle 5", Features{PanelView: true}},
		{"KPW1", Kindle, 2012, 758, 1024, "Kindle Paperwhite 1", Features{PanelView: true}},
		{"KPW2", Kindle, 2013, 758, 1024, "Kindle Paperwhite 2", Features{PanelView: true}},
		{"K7", Kindle, 2014, 600, 800, "Kindle 7", Features{PanelView: true}},
		{"KV", Kindle, 2014, 1072, 1448, "Kindle Voyage", Features{PanelView: true}},
		{"KPW3", Kindle, 2015, 1072, 1448, "Kindle Paperwhite 3", Features{PanelView: true}},
		{"KO1", Kindle, 2016, 1072, 1448, "Kindle Oasis 1", Features{PanelView: true}},
		{"K8", Kindle, 2016, 600, 800, "Kindle 8", Features{PanelView: true}},
		{"KO2", Kindle, 2017, 1264, 1680, "Kindle Oasis 2", Features{PanelView: true}},
		{"KPW4", Kindle, 2018, 1072, 1448, "Kindle Paperwhite 4", Features{PanelView: true}},
		{"K10", Kindle, 2019, 600, 800, "Kindle 10", Features{PanelView: true}},
		{"KO3", Kindle, 2019, 1264, 1680, "Kindle Oasis 3", Features{PanelView: true}},
		{"KPW5", Kindle, 2021, 1236, 1648, "Kindle Paperwhite 5/Signature Edition", Features{PanelView: true}},
		{"K11", Kindle, 2022, 1072, 1448, "Kindle 11", Features{}},
		{"KS", Kindle, 2022, 1860, 2480, "Kindle Scribe", Features{}},
		{"K12", Kindle, 2024, 1072, 1448, "Kindle 12", Features{}},
		{"KPW12", Kindle, 2024, 1264, 1680, "Kindle Paperwhite 12/Signature Edition", Features{}},
		{"KCS", Kindle, 2024, 1264, 1680, "Kindle Colorsoft", Features{Color: true}},
		{"KS2", Kindle, 2024, 1860, 2480, "Kindle Scribe 2", Features{}},
		{"KS3", Kindle, 2025, 1980, 2640, "Kindle Scribe 3", Features{}},
		{"KSC", Kindle, 2025, 1980, 2640, "Kindle Scribe Colorsoft", Features{Color: true}},
		// Kobo, by year of release
		{"KoR", Kobo, 2010, 600, 800, "Kobo eReader", Features{}},
		{"KoT", Kobo, 2011, 600, 800, "Kobo Touch", Features{}},
		{"KoM", Kobo, 2012, 600, 800, "Kobo Mini", Features{}},
		{"KoG", Kobo, 2012, 758, 1024, "Kobo Glo", Features{}},
		{"KoA", Kobo, 2013, 758, 1024, "Kobo Aura", Features{}},
		{"KoAHD", Kobo, 2013, 1080, 1440, "Kobo Aura HD", Features{}},
		{"KoAH2O", Kobo, 2014, 1080, 1430, "Kobo Aura H2O", Features{}},
		{"KoGHD", Kobo, 2015, 1072, 1448, "Kobo Glo HD", Features{}},
		{"KoT2", Kobo, 2015, 600, 800, "Kobo Touch 2.0", Features{}},
		{"KoAO", Kobo, 2016, 1404, 1872, "Kobo Aura One", Features{}},
		{"KoA2", Kobo, 2016, 758, 1024, "Kobo Aura Edition 2", Features{}},
		{"KoAH2O2", Kobo, 2017, 1080, 1430, "Kobo Aura H2O Edition 2", Features{}},
		{"KoCHD", Kobo, 2018, 1072, 1448, "Kobo Clara HD", Features{}},
		{"KoF", Kobo, 2018, 1440, 1920, "Kobo Forma", Features{}},
		{"KoLH2O", Kobo, 2019, 1264, 1680, "Kobo Libra H2O", Features{}},
		{"KoN", Kobo, 2020, 758, 1024, "Kobo Nia", Features{}},
		{"KoE", Kobo, 2021, 1404, 1872, "Kobo Elipsa", Features{}},
		{"KoS", Kobo, 2021, 1440, 1920, "Kobo Sage", Features{}},
		{"KoL2", Kobo, 2021, 1264, 1680, "Kobo Libra 2", Features{}},
		{"KoC2E", Kobo, 2022, 1072, 1448, "Kobo Clara 2E", Features{}},
		{"KoE2", Kobo, 2023, 1404, 1872, "Kobo Elipsa 2E", Features{}},
		{"KoCBW", Kobo, 2024, 1072, 1448, "Kobo Clara BW", Features{}},
		{"KoCC", Kobo, 2024, 1072, 1448, "Kobo Clara Colour", Features{Color: true}},
		{"KoLC", Kobo, 2024, 1264, 1680, "Kobo Libra Colour", Features{Color: true}},
		// reMarkable
		{"RM1", ReMarkable, 2017, 1404, 1872, "reMarkable 1", Features{}},
		{"RM2", ReMarkable, 2020, 1404, 1872, "reMarkable 2", Features{}},
		{"RMP", ReMarkable, 2024, 1620, 2160, "reMarkable Paper Pro", Features{Color: true}},
		{"RMPM", ReMarkable, 2025, 954, 1696, "reMarkable Paper Pro Move", Features{Color: true}},
		{"RMPU", ReMarkable, 2026, 1404, 1872, "reMarkable Paper Pure", Features{}},
	}
}

func (p Profiles) String() string {
	s := make([]string, 0)
	var last DeviceType
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
			"      - %-7s %7s - %4d x %-4d - %s",
			v.Code,
			year,
			v.Width, v.Height,
			v.Description,
		))
	}
	return strings.Join(s, "\n")
}
