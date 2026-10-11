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
		{"aspect ratio device", func(c *Converter) { c.Options.Image.View.AspectRatio = -1 }, false},
		{"aspect ratio source", func(c *Converter) { c.Options.Image.View.AspectRatio = 0 }, false},
		{"aspect ratio 1.6", func(c *Converter) { c.Options.Image.View.AspectRatio = 1.6 }, false},
		{"aspect ratio min", func(c *Converter) { c.Options.Image.View.AspectRatio = 0.5 }, false},
		{"aspect ratio max", func(c *Converter) { c.Options.Image.View.AspectRatio = 3 }, false},
		{"aspect ratio too small", func(c *Converter) { c.Options.Image.View.AspectRatio = 0.0001 }, true},
		{"aspect ratio too big", func(c *Converter) { c.Options.Image.View.AspectRatio = 1000 }, true},
		{"aspect ratio negative", func(c *Converter) { c.Options.Image.View.AspectRatio = -2 }, true},
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

// the extension of the output is case insensitive: a name ending with .EPUB is a file, not a directory.
func TestValidateOutputExtension(t *testing.T) {
	for _, name := range []string{"out.epub", "out.EPUB", "out.Epub"} {
		t.Run(name, func(t *testing.T) {
			conv := New()
			conv.Options.Input = t.TempDir()
			conv.Options.Output = filepath.Join(t.TempDir(), name)
			conv.Options.Workers = 1
			want := conv.Options.Output

			if err := conv.Validate(); err != nil {
				t.Fatal(err)
			}
			if conv.Options.Output != want {
				t.Errorf("got output %q, want %q", conv.Options.Output, want)
			}
		})
	}
}
