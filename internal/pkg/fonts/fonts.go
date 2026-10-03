// Package fonts provides the font used to render the cover and title page.
package fonts

import (
	_ "embed"
	"fmt"
	"os"

	"github.com/golang/freetype/truetype"
)

// Literata is licensed under the SIL Open Font License 1.1 (see OFL.txt).
//
//go:embed "Literata-Regular.ttf"
var literata []byte

// Default is the embedded font, used when no custom font is provided.
var Default = mustParse(literata)

func mustParse(b []byte) *truetype.Font {
	f, err := truetype.Parse(b)
	if err != nil {
		panic(err)
	}
	return f
}

// Load returns the TrueType font at path, or the Default font if path is empty.
func Load(path string) (*truetype.Font, error) {
	if path == "" {
		return Default, nil
	}
	b, err := os.ReadFile(path)
	if err != nil {
		return nil, fmt.Errorf("font: %w", err)
	}
	f, err := truetype.Parse(b)
	if err != nil {
		return nil, fmt.Errorf("font %s: %w", path, err)
	}
	return f, nil
}
