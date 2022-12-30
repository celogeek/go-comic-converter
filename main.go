package main

import (
	"flag"
	"fmt"
	"go-comic-converter/internal/epub"
	"os"
	"path/filepath"
	"runtime"
	"strings"
)

type Profile struct {
	Code        string
	Description string
	Width       int
	Height      int
}

var Profiles = []Profile{
	// Kindle
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
	// Kobo
	{"KoMT", "Kobo Mini/Touch", 600, 800},
	{"KoG", "Kobo Glo", 768, 1024},
	{"KoGHD", "Kobo Glo HD", 1072, 1448},
	{"KoA", "Kobo Aura", 758, 1024},
	{"KoAHD", "Kobo Aura HD", 1080, 1440},
	{"KoAH2O", "Kobo Aura H2O", 1080, 1430},
	{"KoAO", "Kobo Aura ONE", 1404, 1872},
	{"KoN", "Kobo Nia", 758, 1024},
	{"KoC", "Kobo Clara HD/Kobo Clara 2E", 1072, 1448},
	{"KoL", "Kobo Libra H2O/Kobo Libra 2", 1264, 1680},
	{"KoF", "Kobo Forma", 1440, 1920},
	{"KoS", "Kobo Sage", 1440, 1920},
	{"KoE", "Kobo Elipsa", 1404, 1872},
}
var ProfilesIdx = map[string]int{}

func init() {
	for i, p := range Profiles {
		ProfilesIdx[p.Code] = i
	}
}

type Option struct {
	Input    string
	Output   string
	Profile  string
	Author   string
	Title    string
	Quality  int
	NoCrop   bool
	LimitMb  int
	PrintMem bool
}

func (o *Option) String() string {
	var desc string
	var width, height int
	if i, ok := ProfilesIdx[o.Profile]; ok {
		profile := Profiles[i]
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
    Input   : %s
    Output  : %s
    Profile : %s - %s - %dx%d
    Author  : %s
    Title   : %s
    Quality : %d
    Crop    : %v
    LimitMb : %s
    PrintMem: %v
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
		o.PrintMem,
	)
}

func PrintMemUsage() {
	var m runtime.MemStats
	runtime.ReadMemStats(&m)

	bToMb := func(b uint64) uint64 { return b / 1024 / 1024 }

	// For info on each, see: https://golang.org/pkg/runtime/#MemStats
	fmt.Printf(`Memory Usage:
	Alloc     : %v MiB
	TotalAlloc: %v MiB
	Sys       : %v MiB
	NumGC     : %v
`,
		bToMb(m.Alloc),
		bToMb(m.TotalAlloc),
		bToMb(m.Sys),
		m.NumGC,
	)
}

func main() {
	availableProfiles := make([]string, 0)
	for _, p := range Profiles {
		availableProfiles = append(availableProfiles, fmt.Sprintf(
			"    - %-7s ( %9s ) - %s",
			p.Code,
			fmt.Sprintf("%dx%d", p.Width, p.Height),
			p.Description,
		))
	}

	opt := &Option{}
	flag.StringVar(&opt.Input, "input", "", "Source of comic to convert")
	flag.StringVar(&opt.Output, "output", "", "Output of the epub: (default [INPUT].epub)")
	flag.StringVar(&opt.Profile, "profile", "", fmt.Sprintf("Profile to use: \n%s", strings.Join(availableProfiles, "\n")))
	flag.StringVar(&opt.Author, "author", "GO Comic Converter", "Author of the epub")
	flag.StringVar(&opt.Title, "title", "", "Title of the epub")
	flag.IntVar(&opt.Quality, "quality", 85, "Quality of the image")
	flag.BoolVar(&opt.NoCrop, "nocrop", false, "Disable cropping")
	flag.IntVar(&opt.LimitMb, "limitmb", 0, "Limit size of the ePub: Default nolimit (0), Minimum 20")
	flag.BoolVar(&opt.PrintMem, "printmem", false, "Print memory usage")
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

	profileIdx, profileMatch := ProfilesIdx[opt.Profile]
	if !profileMatch {
		fmt.Println("Profile doesn't exists!")
		flag.Usage()
		os.Exit(1)
	}
	profile := Profiles[profileIdx]

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

	if opt.PrintMem {
		PrintMemUsage()
	}
	os.Exit(0)
}
