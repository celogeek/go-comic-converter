package main

import (
	"fmt"
	"os"
	"runtime/debug"

	"github.com/celogeek/go-comic-converter/v2/internal/converter"
	"github.com/celogeek/go-comic-converter/v2/internal/epub"
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

		fmt.Printf(`go-comic-converter
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
		cmd.Options.SaveDefault()
		fmt.Fprintf(
			os.Stderr,
			"%s%s\n\nSaving to %s\n",
			cmd.Options.Header(),
			cmd.Options.ShowDefault(),
			cmd.Options.FileName(),
		)
		return
	}

	if cmd.Options.Show {
		fmt.Fprintln(os.Stderr, cmd.Options.Header(), cmd.Options.ShowDefault())
		return
	}

	if cmd.Options.Reset {
		cmd.Options.ResetDefault()
		fmt.Fprintf(
			os.Stderr,
			"%s%s\n\nReset default to %s\n",
			cmd.Options.Header(),
			cmd.Options.ShowDefault(),
			cmd.Options.FileName(),
		)
		return
	}

	if err := cmd.Validate(); err != nil {
		cmd.Fatal(err)
	}

	fmt.Fprintln(os.Stderr, cmd.Options)

	if cmd.Options.Dry {
		return
	}

	profile := cmd.Options.GetProfile()
	if err := epub.NewEpub(&epub.EpubOptions{
		Input:                      cmd.Options.Input,
		Output:                     cmd.Options.Output,
		LimitMb:                    cmd.Options.LimitMb,
		Title:                      cmd.Options.Title,
		Author:                     cmd.Options.Author,
		StripFirstDirectoryFromToc: cmd.Options.StripFirstDirectoryFromToc,
		ImageOptions: &epub.ImageOptions{
			ViewWidth:           profile.Width,
			ViewHeight:          profile.Height,
			Quality:             cmd.Options.Quality,
			Crop:                cmd.Options.Crop,
			Palette:             profile.Palette,
			Brightness:          cmd.Options.Brightness,
			Contrast:            cmd.Options.Contrast,
			AutoRotate:          cmd.Options.AutoRotate,
			AutoSplitDoublePage: cmd.Options.AutoSplitDoublePage,
			NoBlankPage:         cmd.Options.NoBlankPage,
			Manga:               cmd.Options.Manga,
			HasCover:            cmd.Options.HasCover,
			AddPanelView:        cmd.Options.AddPanelView,
			Workers:             cmd.Options.Workers,
		},
	}).Write(); err != nil {
		fmt.Fprintf(os.Stderr, "Error: %v\n", err)
		os.Exit(1)
	}

	os.Exit(0)
}
