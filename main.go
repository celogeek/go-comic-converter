package main

import (
	"fmt"
	"os"

	"github.com/celogeek/go-comic-converter/v2/internal/converter"
	"github.com/celogeek/go-comic-converter/v2/internal/epub"
)

func main() {
	cmd := converter.New()
	if err := cmd.LoadConfig(); err != nil {
		cmd.Fatal(err)
	}
	cmd.InitParse()
	cmd.Parse()

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
		Input:   cmd.Options.Input,
		Output:  cmd.Options.Output,
		LimitMb: cmd.Options.LimitMb,
		Title:   cmd.Options.Title,
		Author:  cmd.Options.Author,
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
