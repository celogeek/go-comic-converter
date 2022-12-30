package main

import (
	"flag"
	"fmt"
	"go-comic-converter/internal/epub"
	"os"
	"path/filepath"
	"strings"
)

type Profile struct {
	Description string
	Width       int
	Height      int
}

var Profiles = map[string]Profile{
	"KS": {"Kindle Scribe", 1860, 2480},
}

type Option struct {
	Input   string
	Output  string
	Profile string
	Author  string
	Title   string
	Quality int
	NoCrop  bool
	LimitMb int
}

func (o *Option) String() string {
	var desc string
	var width, height int
	if profile, ok := Profiles[o.Profile]; ok {
		desc = profile.Description
		width = profile.Width
		height = profile.Height
	}
	limitmb := "nolimit"
	if o.LimitMb > 0 {
		limitmb = fmt.Sprintf("%d Mb", o.LimitMb)
	}

	return fmt.Sprintf(`Go Comic Converter

Options:
    Input  : %s
    Output : %s
    Profile: %s - %s - %dx%d
    Author : %s
    Title  : %s
    Quality: %d
    Crop   : %v
    LimitMb: %s
`,
		o.Input,
		o.Output,
		o.Profile,
		desc,
		width,
		height,
		o.Author,
		o.Title,
		o.Quality,
		!o.NoCrop,
		limitmb,
	)
}

func main() {
	availableProfiles := make([]string, 0)
	for k := range Profiles {
		availableProfiles = append(availableProfiles, k)
	}

	opt := &Option{}
	flag.StringVar(&opt.Input, "input", "", "Source of comic to convert")
	flag.StringVar(&opt.Output, "output", "", "Output of the epub: (default [INPUT].epub)")
	flag.StringVar(&opt.Profile, "profile", "", fmt.Sprintf("Profile to use: %s", strings.Join(availableProfiles, ", ")))
	flag.StringVar(&opt.Author, "author", "GO Comic Converter", "Author of the epub")
	flag.StringVar(&opt.Title, "title", "", "Title of the epub")
	flag.IntVar(&opt.Quality, "quality", 85, "Quality of the image")
	flag.BoolVar(&opt.NoCrop, "nocrop", false, "Disable cropping")
	flag.IntVar(&opt.LimitMb, "limitmb", 0, "Limit size of the ePub: Default nolimit (0), Minimum 20")
	flag.Parse()

	if opt.Input == "" {
		fmt.Println("Missing input or output!")
		flag.Usage()
		os.Exit(1)
	}

	if opt.Output == "" {
		fi, err := os.Stat(opt.Input)
		if err != nil {
			fmt.Println(err)
			flag.Usage()
			os.Exit(1)
		}
		if fi.IsDir() {
			opt.Output = fmt.Sprintf("%s.epub", filepath.Clean(opt.Input))
		} else {
			ext := filepath.Ext(opt.Input)
			opt.Output = fmt.Sprintf("%s.epub", opt.Input[0:len(opt.Input)-len(ext)])
		}
	}

	profile, profileMatch := Profiles[opt.Profile]
	if !profileMatch {
		fmt.Println("Profile doesn't exists!")
		flag.Usage()
		os.Exit(1)
	}

	if opt.LimitMb > 0 && opt.LimitMb < 20 {
		fmt.Println("LimitMb should be 0 or >= 20")
		flag.Usage()
		os.Exit(1)
	}

	if opt.Title == "" {
		opt.Title = filepath.Base(opt.Input)
	}

	fmt.Println(opt)

	if err := epub.NewEpub(&epub.EpubOptions{
		Input:   opt.Input,
		Output:  opt.Output,
		LimitMb: opt.LimitMb,
		Title:   opt.Title,
		Author:  opt.Author,
		ImageOptions: &epub.ImageOptions{
			ViewWidth:  profile.Width,
			ViewHeight: profile.Height,
			Quality:    opt.Quality,
			Crop:       !opt.NoCrop,
		},
	}).Write(); err != nil {
		fmt.Printf("Error: %v\n", err)
		os.Exit(1)
	}

	os.Exit(0)
}
