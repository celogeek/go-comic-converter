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
}

func (p Profile) String() string {
	return p.Code + " - " + p.Description + " - " + utils.IntToString(p.Width) + "x" + utils.IntToString(p.Height)
}

type Profiles []Profile

// NewProfiles Initialize list of all supported profiles.
func NewProfiles() Profiles {
	return []Profile{
		// High Resolution for Tablet
		{"HR", Generic, 0, 2400, 3840, "High Resolution"},
		{"SR", Generic, 0, 1200, 1920, "Standard Resolution"},
		// Kindle, by year of release
		{"K1", Kindle, 2007, 600, 800, "Kindle 1"},
		{"K2", Kindle, 2009, 600, 800, "Kindle 2"},
		{"KDX", Kindle, 2009, 824, 1200, "Kindle DX/DXG"},
		{"K3", Kindle, 2010, 600, 800, "Kindle Keyboard"},
		{"K4", Kindle, 2011, 600, 800, "Kindle 4"},
		{"KT", Kindle, 2011, 600, 800, "Kindle Touch"},
		{"K5", Kindle, 2012, 600, 800, "Kindle 5"},
		{"KPW1", Kindle, 2012, 758, 1024, "Kindle Paperwhite 1"},
		{"KPW2", Kindle, 2013, 758, 1024, "Kindle Paperwhite 2"},
		{"K7", Kindle, 2014, 600, 800, "Kindle 7"},
		{"KV", Kindle, 2014, 1072, 1448, "Kindle Voyage"},
		{"KPW3", Kindle, 2015, 1072, 1448, "Kindle Paperwhite 3"},
		{"KO1", Kindle, 2016, 1072, 1448, "Kindle Oasis 1"},
		{"K8", Kindle, 2016, 600, 800, "Kindle 8"},
		{"KO2", Kindle, 2017, 1264, 1680, "Kindle Oasis 2"},
		{"KPW4", Kindle, 2018, 1072, 1448, "Kindle Paperwhite 4"},
		{"K10", Kindle, 2019, 600, 800, "Kindle 10"},
		{"KO3", Kindle, 2019, 1264, 1680, "Kindle Oasis 3"},
		{"KPW5", Kindle, 2021, 1236, 1648, "Kindle Paperwhite 5/Signature Edition"},
		{"K11", Kindle, 2022, 1072, 1448, "Kindle 11"},
		{"KS", Kindle, 2022, 1860, 2480, "Kindle Scribe"},
		{"K12", Kindle, 2024, 1072, 1448, "Kindle 12"},
		{"KPW12", Kindle, 2024, 1264, 1680, "Kindle Paperwhite 12/Signature Edition"},
		{"KCS", Kindle, 2024, 1264, 1680, "Kindle Colorsoft"},
		{"KS2", Kindle, 2024, 1860, 2480, "Kindle Scribe 2"},
		{"KS3", Kindle, 2025, 1980, 2640, "Kindle Scribe 3"},
		{"KSC", Kindle, 2025, 1980, 2640, "Kindle Scribe Colorsoft"},
		// Kobo, by year of release
		{"KoR", Kobo, 2010, 600, 800, "Kobo eReader"},
		{"KoT", Kobo, 2011, 600, 800, "Kobo Touch"},
		{"KoM", Kobo, 2012, 600, 800, "Kobo Mini"},
		{"KoG", Kobo, 2012, 758, 1024, "Kobo Glo"},
		{"KoA", Kobo, 2013, 758, 1024, "Kobo Aura"},
		{"KoAHD", Kobo, 2013, 1080, 1440, "Kobo Aura HD"},
		{"KoAH2O", Kobo, 2014, 1080, 1430, "Kobo Aura H2O"},
		{"KoGHD", Kobo, 2015, 1072, 1448, "Kobo Glo HD"},
		{"KoT2", Kobo, 2015, 600, 800, "Kobo Touch 2.0"},
		{"KoAO", Kobo, 2016, 1404, 1872, "Kobo Aura One"},
		{"KoA2", Kobo, 2016, 758, 1024, "Kobo Aura Edition 2"},
		{"KoAH2O2", Kobo, 2017, 1080, 1430, "Kobo Aura H2O Edition 2"},
		{"KoCHD", Kobo, 2018, 1072, 1448, "Kobo Clara HD"},
		{"KoF", Kobo, 2018, 1440, 1920, "Kobo Forma"},
		{"KoLH2O", Kobo, 2019, 1264, 1680, "Kobo Libra H2O"},
		{"KoN", Kobo, 2020, 758, 1024, "Kobo Nia"},
		{"KoE", Kobo, 2021, 1404, 1872, "Kobo Elipsa"},
		{"KoS", Kobo, 2021, 1440, 1920, "Kobo Sage"},
		{"KoL2", Kobo, 2021, 1264, 1680, "Kobo Libra 2"},
		{"KoC2E", Kobo, 2022, 1072, 1448, "Kobo Clara 2E"},
		{"KoE2", Kobo, 2023, 1404, 1872, "Kobo Elipsa 2E"},
		{"KoCBW", Kobo, 2024, 1072, 1448, "Kobo Clara BW"},
		{"KoCC", Kobo, 2024, 1072, 1448, "Kobo Clara Colour"},
		{"KoLC", Kobo, 2024, 1264, 1680, "Kobo Libra Colour"},
		// reMarkable
		{"RM1", ReMarkable, 2017, 1404, 1872, "reMarkable 1"},
		{"RM2", ReMarkable, 2020, 1404, 1872, "reMarkable 2"},
		{"RMP", ReMarkable, 2024, 1620, 2160, "reMarkable Paper Pro"},
		{"RMPM", ReMarkable, 2025, 954, 1696, "reMarkable Paper Pro Move"},
		{"RMPU", ReMarkable, 2026, 1404, 1872, "reMarkable Paper Pure"},
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
