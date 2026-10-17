package converter

import (
	"errors"
	"github.com/celogeek/go-comic-converter/v3/pkg/epuboptions"
	"os"
	"path/filepath"
	"strings"
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
		{"jxl effort 1", func(c *Converter) { c.Options.Image.JXLEffort = 1 }, false},
		{"jxl effort 7", func(c *Converter) { c.Options.Image.JXLEffort = 7 }, false},
		{"jxl effort 0", func(c *Converter) { c.Options.Image.JXLEffort = 0 }, true},
		{"jxl effort 8", func(c *Converter) { c.Options.Image.JXLEffort = 8 }, true},
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
		{"landscape with a profile that has it", func(c *Converter) { c.Options.Profile = "KS"; c.Options.Image.Landscape = true }, false},
		{"landscape off with a profile that has not it", func(c *Converter) { c.Options.Profile = "SR"; c.Options.Image.Landscape = false }, false},
		{"landscape with a kindle without it", func(c *Converter) { c.Options.Profile = "SR"; c.Options.Image.Landscape = true }, true},
		{"landscape with a kobo", func(c *Converter) { c.Options.Profile = "KoE"; c.Options.Image.Landscape = true }, true},
		{"jxl with apple books", func(c *Converter) { c.Options.Profile = "AB"; c.Options.Image.Format = "jxl" }, false},
		{"jxl with a kindle", func(c *Converter) { c.Options.Profile = "SR"; c.Options.Image.Format = "jxl" }, true},
		{"jxl with a kobo", func(c *Converter) { c.Options.Profile = "KoE"; c.Options.Image.Format = "jxl" }, true},
		{"jxl with a remarkable", func(c *Converter) { c.Options.Profile = "RMP"; c.Options.Image.Format = "jxl" }, true},
		{"manga with a reMarkable", func(c *Converter) { c.Options.Profile = "RM2"; c.Options.Image.Mode = epuboptions.ModeManga }, true},
		{"normal with a reMarkable", func(c *Converter) { c.Options.Profile = "RM2"; c.Options.Image.Mode = epuboptions.ModeComic }, false},
		{"manga with a kobo", func(c *Converter) { c.Options.Profile = "KoE"; c.Options.Image.Mode = epuboptions.ModeManga }, false},
		{"webtoon with a reMarkable", func(c *Converter) {
			c.Options.Profile = "RM2"
			c.Options.Image.Mode = epuboptions.ModeWebtoon
			c.Options.Image.View.AspectRatio = -1
		}, false},
		{"webtoon with jpeg", func(c *Converter) {
			c.Options.Image.Mode = epuboptions.ModeWebtoon
			c.Options.Image.View.AspectRatio = -1
			c.Options.Image.Format = "jpeg"
		}, false},
		{"webtoon with copy", func(c *Converter) {
			c.Options.Image.Mode = epuboptions.ModeWebtoon
			c.Options.Image.View.AspectRatio = -1
			c.Options.Image.Format = "copy"
		}, true},
		{"webtoon with the aspect ratio of the device", func(c *Converter) {
			c.Options.Image.Mode = epuboptions.ModeWebtoon
			c.Options.Image.View.AspectRatio = -1
		}, false},
		{"webtoon with an aspect ratio", func(c *Converter) {
			c.Options.Image.Mode = epuboptions.ModeWebtoon
			c.Options.Image.View.AspectRatio = 1.6
		}, false},
		{"webtoon with the aspect ratio of the source", func(c *Converter) {
			c.Options.Image.Mode = epuboptions.ModeWebtoon
			c.Options.Image.View.AspectRatio = 0
		}, true},
		{"mode unknown", func(c *Converter) { c.Options.Image.Mode = 3 }, true},
		{"landscape with the split", func(c *Converter) {
			c.Options.Profile = "KS"
			c.Options.Image.Landscape = true
			c.Options.Image.DoublePageMode = 0
		}, false},
		{"landscape with split + rotate", func(c *Converter) {
			c.Options.Profile = "KS"
			c.Options.Image.Landscape = true
			c.Options.Image.DoublePageMode = 1
		}, true},
		{"landscape with rotate", func(c *Converter) {
			c.Options.Profile = "KS"
			c.Options.Image.Landscape = true
			c.Options.Image.DoublePageMode = 2
		}, true},
		{"rotate without landscape", func(c *Converter) { c.Options.Profile = "KS"; c.Options.Image.DoublePageMode = 2 }, false},
		{"zoom default", func(c *Converter) { c.Options.Image.Zoom = 0 }, false},
		{"zoom min", func(c *Converter) { c.Options.Image.Zoom = 1 }, false},
		{"zoom max", func(c *Converter) { c.Options.Image.Zoom = 3 }, false},
		{"zoom too small", func(c *Converter) { c.Options.Image.Zoom = 0.5 }, true},
		{"zoom too big", func(c *Converter) { c.Options.Image.Zoom = 3.1 }, true},
		{"zoom negative", func(c *Converter) { c.Options.Image.Zoom = -1 }, true},
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
			if got := conv.Options.Image.Features.PanelView; got != c.want {
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

// every profile has a zoom: the Apple Books profiles are stored 1.5 times bigger, the others as the view.
func TestProfilesZoom(t *testing.T) {
	for _, p := range NewProfiles() {
		want := 1.0
		if p.Type == epuboptions.Apple {
			want = 1.5
		}
		if p.Features.Zoom != want {
			t.Errorf("%s: got zoom %v, want %v", p.Code, p.Features.Zoom, want)
		}
	}
}

// the zoom comes from the profile, unless it is set.
func TestParseZoom(t *testing.T) {
	for _, c := range []struct {
		name string
		args []string
		want float64
	}{
		{"kindle", []string{"-profile", "KS"}, 1},
		{"apple", []string{"-profile", "AB"}, 1.5},
		{"apple with zoom", []string{"-profile", "AB", "-zoom", "2"}, 2},
		{"apple with zoom 1", []string{"-profile", "AB", "-zoom", "1"}, 1},
		{"kindle with zoom", []string{"-profile", "KS", "-zoom", "1.5"}, 1.5},
	} {
		t.Run(c.name, func(t *testing.T) {
			conv := New()
			conv.InitParse()
			old := os.Args
			defer func() { os.Args = old }()
			os.Args = append([]string{"go-comic-converter"}, c.args...)
			conv.Parse()
			if got := conv.Options.Image.ZoomFactor(); got != c.want {
				t.Errorf("got zoom %v, want %v", got, c.want)
			}
		})
	}
}

// the profiles are listed as a table, a row per profile.
func TestProfilesString(t *testing.T) {
	lines := strings.Split(NewProfiles().String(), "\n")
	if want := len(NewProfiles()) + 2; len(lines) != want {
		t.Fatalf("got %d lines, want %d", len(lines), want)
	}
	for _, h := range []string{"Device", "Code", "Year", "Resolution", "Zoom", "Description", "Features"} {
		if !strings.Contains(lines[0], h) {
			t.Errorf("header %q missing: %q", h, lines[0])
		}
	}
	for _, l := range lines {
		if strings.TrimRight(l, " ") != l {
			t.Errorf("trailing spaces: %q", l)
		}
	}
	if strings.Trim(lines[1], "- ") != "" {
		t.Errorf("missing separator: %q", lines[1])
	}
	for _, c := range []struct {
		code string
		want []string
	}{
		{"K1", []string{"Kindle", "2007", "600 x 800", "x1", "Kindle 1", "panel view, manga"}},
		{"KSC", []string{"Kindle", "2025", "1980 x 2640", "x1", "Kindle Scribe Colorsoft", "color, manga, landscape"}},
		{"AB", []string{"Apple", "1640 x 2360", "x1.5", "Apple Books Standard Resolution", "color, manga, jxl"}},
		{"RM1", []string{"reMarkable", "2017", "1404 x 1872", "x1", "reMarkable 1"}},
	} {
		var row string
		for _, l := range lines[2:] {
			if f := strings.Fields(l); len(f) > 1 && f[1] == c.code {
				row = l
			}
		}
		if row == "" {
			t.Errorf("%s: no row", c.code)
		}
		for _, w := range c.want {
			if !strings.Contains(row, w) {
				t.Errorf("%s: %q missing from %q", c.code, w, row)
			}
		}
	}
}

func TestIgnoreConfig(t *testing.T) {
	for _, c := range []struct {
		args []string
		want bool
	}{
		{nil, false},
		{[]string{"-profile", "KS"}, false},
		{[]string{"-ignore-config"}, true},
		{[]string{"--ignore-config"}, true},
		{[]string{"-h", "-ignore-config"}, true},
		{[]string{"-ignore-config=1"}, true},
		{[]string{"-ignore-config=true"}, true},
		{[]string{"-ignore-config=0"}, false},
		{[]string{"-ignore-config", "-ignore-config=0"}, false},
		{[]string{"-ignore-config=0", "-ignore-config"}, true},
		// after the end of the flags, or as a value
		{[]string{"--", "-ignore-config"}, false},
		{[]string{"-title", "ignore-config"}, false},
		{[]string{"-ignore-configs"}, false},
	} {
		if got := ignoreConfig(c.args); got != c.want {
			t.Errorf("%v: got %v, want %v", c.args, got, c.want)
		}
	}
}

// the saved config gives the default of the options (and of the help), unless it is ignored.
func TestLoadConfigIgnored(t *testing.T) {
	home := t.TempDir()
	t.Setenv("HOME", home)
	if err := os.WriteFile(filepath.Join(home, ".go-comic-converter.yaml"), []byte("profile: KS\nepuboptions:\n  image:\n    quality: 50\n"), 0o600); err != nil {
		t.Fatal(err)
	}

	for _, c := range []struct {
		name        string
		args        []string
		wantProfile string
		wantQuality int
	}{
		{"saved config", []string{"-h"}, "KS", 50},
		{"ignored", []string{"-h", "-ignore-config"}, "SR", 85},
		{"ignored with a flag", []string{"-ignore-config", "-profile", "K1"}, "SR", 85},
	} {
		t.Run(c.name, func(t *testing.T) {
			old := os.Args
			defer func() { os.Args = old }()
			os.Args = append([]string{"go-comic-converter"}, c.args...)

			conv := New()
			if err := conv.LoadConfig(); err != nil {
				t.Fatal(err)
			}
			if conv.Options.Profile != c.wantProfile {
				t.Errorf("got profile %q, want %q", conv.Options.Profile, c.wantProfile)
			}
			if conv.Options.Image.Quality != c.wantQuality {
				t.Errorf("got quality %d, want %d", conv.Options.Image.Quality, c.wantQuality)
			}
		})
	}
}

// every device supports the manga mode, but the reMarkable.
func TestProfilesManga(t *testing.T) {
	for _, p := range NewProfiles() {
		if want := p.Type != epuboptions.ReMarkable; p.Features.Manga != want {
			t.Errorf("profile %s: manga = %v, want %v", p.Code, p.Features.Manga, want)
		}
	}
}
