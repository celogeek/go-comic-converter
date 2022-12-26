package main

import (
	"flag"
	"fmt"
	"go-comic-converter/internal/epub"
	"path/filepath"
	"strings"
)

type Profile struct {
	Width  int
	Height int
}

var Profiles = map[string]Profile{
	"KS": {1860, 2480},
}

type Option struct {
	Input   string
	Output  string
	Profile string
	Author  string
	Title   string
	Quality int
}

func (o *Option) String() string {
	var width, height int
	profile, profileMatch := Profiles[o.Profile]
	if profileMatch {
		width = profile.Width
		height = profile.Height
	}

	return fmt.Sprintf(`Options:
	Input  : %s
	Output : %s
	Profile: %s (%dx%d)
	Author : %s
	Title  : %s
	Quality: %d
`,
		o.Input,
		o.Output,
		o.Profile,
		width,
		height,
		o.Author,
		o.Title,
		o.Quality,
	)
}

func main() {
	availableProfiles := make([]string, 0)
	for k := range Profiles {
		availableProfiles = append(availableProfiles, k)
	}

	opt := &Option{}
	flag.StringVar(&opt.Input, "input", "", "Source of comic to convert")
	flag.StringVar(&opt.Output, "output", "", "Output of the epub")
	flag.StringVar(&opt.Profile, "profile", "", fmt.Sprintf("Profile to use: %s", strings.Join(availableProfiles, ", ")))
	flag.StringVar(&opt.Author, "author", "GO Comic Converter", "Author of the epub")
	flag.StringVar(&opt.Title, "title", "", "Title of the epub")
	flag.IntVar(&opt.Quality, "quality", 75, "Quality of the image: Default 75")
	flag.Parse()

	if opt.Input == "" || opt.Output == "" {
		fmt.Println("Missing input or output!")
		flag.Usage()
		return
	}
	profile, profileMatch := Profiles[opt.Profile]
	if !profileMatch {
		fmt.Println("Profile doesn't exists!")
		flag.Usage()
		return
	}

	if opt.Title == "" {
		opt.Title = filepath.Base(opt.Input)
	}

	fmt.Println(opt)

	err := epub.NewEpub(opt.Output).
		SetSize(profile.Width, profile.Height).
		SetQuality(opt.Quality).
		SetTitle(opt.Input).
		SetAuthor(opt.Author).
		LoadDir(opt.Input).
		Write()

	if err != nil {
		panic(err)
	}
}
