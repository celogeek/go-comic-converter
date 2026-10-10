// Package converter profiles manage supported profiles for go-comic-converter.
package converter

import (
	"fmt"
	"strings"

	"github.com/celogeek/go-comic-converter/v3/internal/pkg/utils"
)

type Profile struct {
	Code        string `json:"code"`
	Description string `json:"description"`
	Width       int    `json:"width"`
	Height      int    `json:"height"`
}

func (p Profile) String() string {
	return p.Code + " - " + p.Description + " - " + utils.IntToString(p.Width) + "x" + utils.IntToString(p.Height)
}

type Profiles []Profile

// NewProfiles Initialize list of all supported profiles.
func NewProfiles() Profiles {
	return []Profile{
		// High Resolution for Tablet
		{"HR", "High Resolution", 2400, 3840},
		{"SR", "Standard Resolution", 1200, 1920},
		//Kindle
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
		// reMarkable
		{"RM1", "reMarkable 1", 1404, 1872},
		{"RM2", "reMarkable 2", 1404, 1872},
	}
}

func (p Profiles) String() string {
	s := make([]string, 0)
	for _, v := range p {
		s = append(s, fmt.Sprintf(
			"    - %-7s - %4d x %-4d - %s",
			v.Code,
			v.Width, v.Height,
			v.Description,
		))
	}
	return strings.Join(s, "\n")
}
