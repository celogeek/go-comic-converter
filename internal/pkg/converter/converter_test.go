package converter

import (
	"path/filepath"
	"testing"
)

func TestValidate(t *testing.T) {
	for _, c := range []struct {
		name    string
		mutate  func(c *Converter)
		wantErr bool
	}{
		{"defaults", func(c *Converter) {}, false},
		{"workers 1", func(c *Converter) { c.Options.Workers = 1 }, false},
		{"workers 0", func(c *Converter) { c.Options.Workers = 0 }, true},
		{"workers negative", func(c *Converter) { c.Options.Workers = -1 }, true},
		{"quality 1", func(c *Converter) { c.Options.Image.Quality = 1 }, false},
		{"quality 100", func(c *Converter) { c.Options.Image.Quality = 100 }, false},
		{"quality 0", func(c *Converter) { c.Options.Image.Quality = 0 }, true},
		{"quality 101", func(c *Converter) { c.Options.Image.Quality = 101 }, true},
	} {
		t.Run(c.name, func(t *testing.T) {
			conv := New()
			conv.Options.Input = t.TempDir()
			conv.Options.Output = filepath.Join(t.TempDir(), "out.epub")
			conv.Options.Workers = 1 // the default comes from the flags
			c.mutate(conv)

			err := conv.Validate()
			if (err != nil) != c.wantErr {
				t.Errorf("got error %v, want error %v", err, c.wantErr)
			}
		})
	}
}
