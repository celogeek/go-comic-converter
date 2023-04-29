/*
Convert CBZ/CBR/Dir into EPUB for e-reader devices (Kindle Devices, ...)

My goal is to make a simple, crossplatform, and fast tool to convert comics into EPUB.

EPUB is now support by Amazon through [SendToKindle](https://www.amazon.com/gp/sendtokindle/), by Email or by using the App. So I've made it simple to support the size limit constraint of those services.
*/
package main

import (
	"fmt"
	"os"
	"runtime/debug"

	"github.com/celogeek/go-comic-converter/v2/internal/converter"
	"github.com/celogeek/go-comic-converter/v2/internal/epub"
	epuboptions "github.com/celogeek/go-comic-converter/v2/internal/epub/options"
	"github.com/tcnksm/go-latest"
)

func main() {
	cmd := converter.New()
	if err := cmd.LoadConfig(); err != nil {
		cmd.Fatal(err)
	}
	cmd.InitParse()
	cmd.Parse()

	if cmd.Options.Version {
		bi, ok := debug.ReadBuildInfo()
		if !ok {
			fmt.Fprintln(os.Stderr, "failed to fetch current version")
			os.Exit(1)
		}

		githubTag := &latest.GithubTag{
			Owner:      "celogeek",
			Repository: "go-comic-converter",
		}
		v, err := githubTag.Fetch()
		if err != nil || len(v.Versions) < 1 {
			fmt.Fprintln(os.Stderr, "failed to fetch the latest version")
			os.Exit(1)
		}
		latest_version := v.Versions[0]

		fmt.Fprintf(os.Stderr, `go-comic-converter
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
			latest_version.Original(),
			latest_version.Segments()[0],
			latest_version.Original(),
		)
		return
	}

	if cmd.Options.Save {
		cmd.Options.SaveConfig()
		fmt.Fprintf(
			os.Stderr,
			"%s%s\n\nSaving to %s\n",
			cmd.Options.Header(),
			cmd.Options.ShowConfig(),
			cmd.Options.FileName(),
		)
		return
	}

	if cmd.Options.Show {
		fmt.Fprintln(os.Stderr, cmd.Options.Header(), cmd.Options.ShowConfig())
		return
	}

	if cmd.Options.Reset {
		cmd.Options.ResetConfig()
		fmt.Fprintf(
			os.Stderr,
			"%s%s\n\nReset default to %s\n",
			cmd.Options.Header(),
			cmd.Options.ShowConfig(),
			cmd.Options.FileName(),
		)
		return
	}

	if err := cmd.Validate(); err != nil {
		cmd.Fatal(err)
	}

	fmt.Fprintln(os.Stderr, cmd.Options)

	profile := cmd.Options.GetProfile()
	perfectWidth, perfectHeight := profile.PerfectDim()

	if err := epub.New(&epuboptions.Options{
		Input:                      cmd.Options.Input,
		Output:                     cmd.Options.Output,
		LimitMb:                    cmd.Options.LimitMb,
		Title:                      cmd.Options.Title,
		Author:                     cmd.Options.Author,
		StripFirstDirectoryFromToc: cmd.Options.StripFirstDirectoryFromToc,
		SortPathMode:               cmd.Options.SortPathMode,
		Workers:                    cmd.Options.Workers,
		Dry:                        cmd.Options.Dry,
		DryVerbose:                 cmd.Options.DryVerbose,
		Quiet:                      cmd.Options.Quiet,
		Image: &epuboptions.Image{
			Crop: &epuboptions.Crop{
				Enabled: cmd.Options.Crop,
				Left:    cmd.Options.CropRatioLeft,
				Up:      cmd.Options.CropRatioUp,
				Right:   cmd.Options.CropRatioRight,
				Bottom:  cmd.Options.CropRatioBottom,
			},
			Quality:             cmd.Options.Quality,
			Brightness:          cmd.Options.Brightness,
			Contrast:            cmd.Options.Contrast,
			AutoRotate:          cmd.Options.AutoRotate,
			AutoSplitDoublePage: cmd.Options.AutoSplitDoublePage,
			NoBlankImage:        cmd.Options.NoBlankImage,
			Manga:               cmd.Options.Manga,
			HasCover:            cmd.Options.HasCover,
			View: &epuboptions.View{
				Width:  perfectWidth,
				Height: perfectHeight,
			},
		},
	}).Write(); err != nil {
		fmt.Fprintf(os.Stderr, "Error: %v\n", err)
		os.Exit(1)
	}

	os.Exit(0)
}
