/*
Convert CBZ/CBR/Dir into EPUB for e-reader devices (Kindle Devices, ...)

My goal is to make a simple, cross-platform, and fast tool to convert comics into EPUB.

EPUB is now support by Amazon through [SendToKindle](https://www.amazon.com/gp/sendtokindle/), by Email or by using the App. So I've made it simple to support the size limit constraint of those services.
*/
package main

import (
	"encoding/json"
	"os"
	"runtime/debug"

	"github.com/tcnksm/go-latest"

	"github.com/celogeek/go-comic-converter/v2/pkg/converter"
	"github.com/celogeek/go-comic-converter/v2/pkg/epub"
	"github.com/celogeek/go-comic-converter/v2/pkg/epuboptions"
	"github.com/celogeek/go-comic-converter/v2/pkg/utils"
)

func main() {
	cmd := converter.New()
	if err := cmd.LoadConfig(); err != nil {
		cmd.Fatal(err)
	}
	cmd.InitParse()
	cmd.Parse()

	switch {
	case cmd.Options.Version:
		version()
	case cmd.Options.Save:
		save(cmd)
	case cmd.Options.Show:
		show(cmd)
	case cmd.Options.Reset:
		reset(cmd)
	default:
		generate(cmd)
	}

}

func version() {
	bi, ok := debug.ReadBuildInfo()
	if !ok {
		utils.Fatalln("failed to fetch current version")
	}

	githubTag := &latest.GithubTag{
		Owner:      "celogeek",
		Repository: "go-comic-converter",
	}
	v, err := githubTag.Fetch()
	if err != nil || len(v.Versions) < 1 {
		utils.Fatalln("failed to fetch the latest version")
	}
	latestVersion := v.Versions[0]

	utils.Printf(`go-comic-converter
  Path             : %s
  Sum              : %s
  Version          : %s
  Available Version: %s

To install the latest version:
$ go install github.com/celogeek/go-comic-converter/v%d@%s
`,
		bi.Main.Path,
		bi.Main.Sum,
		bi.Main.Version,
		latestVersion.Original(),
		latestVersion.Segments()[0],
		latestVersion.Original(),
	)
}

func save(cmd *converter.Converter) {
	if err := cmd.Options.SaveConfig(); err != nil {
		cmd.Fatal(err)
	}
	utils.Printf(
		"%s%s\n\nSaving to %s\n",
		cmd.Options.Header(),
		cmd.Options.ShowConfig(),
		cmd.Options.FileName(),
	)
}

func show(cmd *converter.Converter) {
	utils.Println(cmd.Options.Header(), cmd.Options.ShowConfig())
}

func reset(cmd *converter.Converter) {
	if err := cmd.Options.ResetConfig(); err != nil {
		cmd.Fatal(err)
	}
	utils.Printf(
		"%s%s\n\nReset default to %s\n",
		cmd.Options.Header(),
		cmd.Options.ShowConfig(),
		cmd.Options.FileName(),
	)
	if err := cmd.Options.ResetConfig(); err != nil {
		cmd.Fatal(err)
	}
	utils.Printf(
		"%s%s\n\nReset default to %s\n",
		cmd.Options.Header(),
		cmd.Options.ShowConfig(),
		cmd.Options.FileName(),
	)

}

func generate(cmd *converter.Converter) {
	if err := cmd.Validate(); err != nil {
		cmd.Fatal(err)
	}

	if cmd.Options.Json {
		_ = json.NewEncoder(os.Stdout).Encode(map[string]any{
			"type": "options", "data": cmd.Options,
		})
	} else {
		utils.Println(cmd.Options)
	}

	profile := cmd.Options.GetProfile()

	if err := epub.New(epuboptions.EPUBOptions{
		Input:                      cmd.Options.Input,
		Output:                     cmd.Options.Output,
		LimitMb:                    cmd.Options.LimitMb,
		Title:                      cmd.Options.Title,
		TitlePage:                  cmd.Options.TitlePage,
		Author:                     cmd.Options.Author,
		StripFirstDirectoryFromToc: cmd.Options.StripFirstDirectoryFromToc,
		SortPathMode:               cmd.Options.SortPathMode,
		Workers:                    cmd.Options.Workers,
		Dry:                        cmd.Options.Dry,
		DryVerbose:                 cmd.Options.DryVerbose,
		Quiet:                      cmd.Options.Quiet,
		Json:                       cmd.Options.Json,
		Image: epuboptions.Image{
			Crop: epuboptions.Crop{
				Enabled:            cmd.Options.Crop,
				Left:               cmd.Options.CropRatioLeft,
				Up:                 cmd.Options.CropRatioUp,
				Right:              cmd.Options.CropRatioRight,
				Bottom:             cmd.Options.CropRatioBottom,
				Limit:              cmd.Options.CropLimit,
				SkipIfLimitReached: cmd.Options.CropSkipIfLimitReached,
			},
			Quality:                   cmd.Options.Quality,
			Brightness:                cmd.Options.Brightness,
			Contrast:                  cmd.Options.Contrast,
			AutoContrast:              cmd.Options.AutoContrast,
			AutoRotate:                cmd.Options.AutoRotate,
			AutoSplitDoublePage:       cmd.Options.AutoSplitDoublePage,
			KeepDoublePageIfSplit:     cmd.Options.KeepDoublePageIfSplit,
			KeepSplitDoublePageAspect: cmd.Options.KeepSplitDoublePageAspect,
			NoBlankImage:              cmd.Options.NoBlankImage,
			Manga:                     cmd.Options.Manga,
			HasCover:                  cmd.Options.HasCover,
			View: epuboptions.View{
				Width:        profile.Width,
				Height:       profile.Height,
				AspectRatio:  cmd.Options.AspectRatio,
				PortraitOnly: cmd.Options.PortraitOnly,
				Color: epuboptions.Color{
					Foreground: cmd.Options.ForegroundColor,
					Background: cmd.Options.BackgroundColor,
				},
			},
			GrayScale:              cmd.Options.Grayscale,
			GrayScaleMode:          cmd.Options.GrayscaleMode,
			Resize:                 !cmd.Options.NoResize,
			Format:                 cmd.Options.Format,
			AppleBookCompatibility: cmd.Options.AppleBookCompatibility,
		},
	}).Write(); err != nil {
		utils.Fatalf("Error: %v\n", err)
	}
	if !cmd.Options.Dry {
		cmd.Stats()
	}
}
