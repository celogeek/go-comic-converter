package converter

import (
	"os"
	"path/filepath"
	"reflect"
	"slices"
	"strings"
	"testing"
)

// parse the arguments with the saved config of the home directory.
func parseArgs(t *testing.T, args ...string) *Converter {
	t.Helper()
	old := os.Args
	t.Cleanup(func() { os.Args = old })
	os.Args = append([]string{"go-comic-converter"}, args...)

	conv := New()
	if err := conv.LoadConfig(); err != nil {
		t.Fatal(err)
	}
	conv.InitParse()
	conv.Parse()
	return conv
}

func TestCommandLineArgs(t *testing.T) {
	home := t.TempDir()
	t.Setenv("HOME", home)

	// the options at their default are not in the command line, except the profile
	args := parseArgs(t, "-profile", "AB", "-manga", "-title", "My Comic", "-input", "/tmp/in").CommandLineArgs()
	want := []string{"-ignore-config", "-input=/tmp/in", "-title=My Comic", "-profile=AB", "-manga=1"}
	if !reflect.DeepEqual(args, want) {
		t.Errorf("got %v, want %v", args, want)
	}

	// the default profile is kept
	if args := parseArgs(t).CommandLineArgs(); !reflect.DeepEqual(args, []string{"-ignore-config", "-profile=SR"}) {
		t.Errorf("got %v", args)
	}

	// the options that differ from the default, the zoom is compared to the one of the profile
	for _, c := range []struct {
		name     string
		args     []string
		want     []string
		unwanted []string
	}{
		{"quality", []string{"-quality", "70", "-crop=0"}, []string{"-quality=70", "-crop=0"}, []string{"-grayscale", "-manga"}},
		{"zoom of the profile", []string{"-profile", "AB", "-zoom", "1.5"}, nil, []string{"-zoom"}},
		{"zoom 1 on apple", []string{"-profile", "AB", "-zoom", "1"}, []string{"-zoom=1"}, nil},
		{"zoom on kindle", []string{"-profile", "KS", "-zoom", "2"}, []string{"-zoom=2"}, nil},
		{"zoom 1 on kindle", []string{"-profile", "KS", "-zoom", "1"}, nil, []string{"-zoom"}},
		{"shortcut", []string{"-goodquality"}, []string{"-quality=90"}, []string{"-format", "-grayscale=", "-resize"}},
		{"grayscale off", []string{"-profile", "KCS", "-grayscale=0"}, []string{"-profile=KCS", "-grayscale=0"}, nil},
	} {
		t.Run(c.name, func(t *testing.T) {
			args := parseArgs(t, c.args...).CommandLineArgs()
			for _, w := range c.want {
				if !slices.Contains(args, w) {
					t.Errorf("%q missing from %v", w, args)
				}
			}
			for _, u := range c.unwanted {
				for _, a := range args {
					if strings.HasPrefix(a, u) {
						t.Errorf("%q must not be in %v", a, args)
					}
				}
			}
			if args[0] != "-ignore-config" {
				t.Errorf("ignore-config must be first: %v", args)
			}
		})
	}
}

// the command line gives the same options, whatever the saved config.
func TestCommandLineSameOptions(t *testing.T) {
	home := t.TempDir()
	t.Setenv("HOME", home)
	if err := os.WriteFile(
		filepath.Join(home, ".go-comic-converter.yaml"),
		[]byte("profile: KS\nepuboptions:\n  image:\n    quality: 50\n    zoom: 2\n    manga: true\n"),
		0o600,
	); err != nil {
		t.Fatal(err)
	}

	for _, c := range []struct {
		name string
		args []string
	}{
		{"saved config", nil},
		{"profile and flags", []string{"-profile", "AB", "-quality", "70", "-crop=0", "-title", "It's a comic"}},
		{"shortcut", []string{"-profile", "KCS", "-goodquality"}},
		{"zoom of the profile", []string{"-profile", "AB"}},
		{"zoom set", []string{"-profile", "AB", "-zoom", "1.25"}},
	} {
		t.Run(c.name, func(t *testing.T) {
			want := parseArgs(t, c.args...)
			got := parseArgs(t, want.CommandLineArgs()...)

			if !reflect.DeepEqual(want.Options.EPUBOptions, got.Options.EPUBOptions) {
				t.Errorf("options differ:\nwant %+v\ngot  %+v", want.Options.EPUBOptions, got.Options.EPUBOptions)
			}
			if want.Options.Profile != got.Options.Profile {
				t.Errorf("profile %q, want %q", got.Options.Profile, want.Options.Profile)
			}
			if want.Options.Image.ZoomFactor() != got.Options.Image.ZoomFactor() {
				t.Errorf("zoom %v, want %v", got.Options.Image.ZoomFactor(), want.Options.Image.ZoomFactor())
			}
		})
	}
}

// ignore-config gives the built-in default, and the command line is made of it.
func TestCommandLineIgnoreConfig(t *testing.T) {
	home := t.TempDir()
	t.Setenv("HOME", home)
	if err := os.WriteFile(
		filepath.Join(home, ".go-comic-converter.yaml"),
		[]byte("epuboptions:\n  image:\n    quality: 50\n"),
		0o600,
	); err != nil {
		t.Fatal(err)
	}

	if args := parseArgs(t).CommandLineArgs(); !slices.Contains(args, "-quality=50") {
		t.Errorf("the saved config must be in %v", args)
	}
	if args := parseArgs(t, "-ignore-config").CommandLineArgs(); !reflect.DeepEqual(args, []string{"-ignore-config", "-profile=SR"}) {
		t.Errorf("the default must not be in %v", args)
	}
}

func TestCommandLineQuote(t *testing.T) {
	conv := parseArgs(t, "-ignore-config", "-title", "Bob's comic", "-input", "/tmp/My Comic.cbz", "-profile", "KS")
	got := conv.CommandLine()
	for _, want := range []string{
		"go-comic-converter -ignore-config \\\n",
		"  -input='/tmp/My Comic.cbz' \\\n",
		`  -title='Bob'\''s comic' \` + "\n",
		"  -profile=KS",
	} {
		if !strings.Contains(got, want) {
			t.Errorf("%q missing from:\n%s", want, got)
		}
	}
	if strings.HasSuffix(got, "\\") {
		t.Errorf("the last line must not continue:\n%s", got)
	}
}
