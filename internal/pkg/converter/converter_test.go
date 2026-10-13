package converter

import (
	"errors"
	"github.com/celogeek/go-comic-converter/v3/pkg/epuboptions"
	"os"
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
		{"grayscale off with color profile", func(c *Converter) { c.Options.Profile = "KCS"; c.Options.Image.GrayScale = false }, false},
		{"grayscale on with color profile", func(c *Converter) { c.Options.Profile = "KCS"; c.Options.Image.GrayScale = true }, false},
		{"grayscale off with the default profile", func(c *Converter) { c.Options.Image.GrayScale = false }, false},
		{"grayscale on with grayscale profile", func(c *Converter) { c.Options.Profile = "K1"; c.Options.Image.GrayScale = true }, false},
		{"grayscale off with grayscale profile", func(c *Converter) { c.Options.Profile = "K1"; c.Options.Image.GrayScale = false }, true},
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

// the panel view only comes from the profile.
func TestParsePanelView(t *testing.T) {
	for _, c := range []struct {
		name string
		args []string
		want bool
	}{
		{"old kindle", []string{"-profile", "K1"}, true},
		{"recent kindle", []string{"-profile", "KS"}, false},
		{"kobo", []string{"-profile", "KoCC"}, false},
		{"apple", []string{"-profile", "AB"}, false},
	} {
		t.Run(c.name, func(t *testing.T) {
			conv := New()
			conv.InitParse()
			old := os.Args
			defer func() { os.Args = old }()
			os.Args = append([]string{"go-comic-converter"}, c.args...)
			conv.Parse()
			if got := conv.Options.Image.AddPanelView; got != c.want {
				t.Errorf("got panel view %v, want %v", got, c.want)
			}
		})
	}
}

// a wrong profile is reported with a dedicated error, so only the profiles are listed.
func TestValidateProfileError(t *testing.T) {
	conv := New()
	conv.Options.Input = t.TempDir()
	conv.Options.Output = filepath.Join(t.TempDir(), "out.epub")
	conv.Options.Workers = 1

	conv.Options.Profile = "NOPE"
	var unknown *UnknownProfileError
	if err := conv.Validate(); !errors.As(err, &unknown) || unknown.Profile != "NOPE" {
		t.Errorf("got error %v, want an unknown profile error", err)
	}

	conv.Options.Profile = ""
	if err := conv.Validate(); !errors.Is(err, ErrProfileMissing) {
		t.Errorf("got error %v, want a missing profile error", err)
	}
}

// the device of the profile is passed to the options, and drives the tile size and the Apple Books mode.
func TestParseDevice(t *testing.T) {
	for _, c := range []struct {
		profile    string
		device     epuboptions.DeviceType
		tileWidth  int
		tileHeight int
	}{
		{"KS", epuboptions.Kindle, 1200, 1920},
		{"KoCC", epuboptions.Kobo, 0, 0},
		{"SR", epuboptions.Kindle, 1200, 1920},
		{"HR", epuboptions.Kindle, 1200, 1920},
		{"AB", epuboptions.Apple, 2288, 2432},
		{"ABP13", epuboptions.Apple, 2288, 2432},
		{"ABI", epuboptions.Apple, 2288, 2432},
		{"ABD", epuboptions.Apple, 2288, 2432},
		{"RM2", epuboptions.ReMarkable, 0, 0},
	} {
		t.Run(c.profile, func(t *testing.T) {
			conv := New()
			conv.InitParse()
			old := os.Args
			defer func() { os.Args = old }()
			os.Args = []string{"go-comic-converter", "-profile", c.profile}
			conv.Parse()

			img := conv.Options.Image
			if img.Device != c.device {
				t.Errorf("got device %q, want %q", img.Device, c.device)
			}
			if tile := img.Tile(); tile.Width != c.tileWidth || tile.Height != c.tileHeight {
				t.Errorf("got tile %dx%d, want %dx%d", tile.Width, tile.Height, c.tileWidth, c.tileHeight)
			}
			if c.device == epuboptions.Apple && (img.DoublePageMode != epuboptions.DoublePageSplit || !img.KeepSplitDoublePageAspect) {
				t.Errorf("apple books must split the double page and keep its aspect")
			}
		})
	}
}
