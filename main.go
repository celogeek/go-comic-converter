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

	"github.com/celogeek/go-comic-converter/v3/internal/pkg/converter"
	"github.com/celogeek/go-comic-converter/v3/internal/pkg/utils"
	"github.com/celogeek/go-comic-converter/v3/pkg/epub"
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
	if err != nil {
		utils.Fatalln("failed to fetch the latest version")
	}
	if len(v.Versions) < 1 {
		utils.Fatalln("no versions found")
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
}

func generate(cmd *converter.Converter) {
	if err := cmd.Validate(); err != nil {
		cmd.Fatal(err)
	}

	if profile := cmd.Options.GetProfile(); profile != nil {
		cmd.Options.Image.View.Width = profile.Width
		cmd.Options.Image.View.Height = profile.Height
	}

	if cmd.Options.Json {
		_ = json.NewEncoder(os.Stdout).Encode(map[string]any{
			"type": "options", "data": cmd.Options,
		})
	} else {
		utils.Println(cmd.Options)
	}

	if err := epub.New(cmd.Options.EPUBOptions).Write(); err != nil {
		utils.Fatalf("Error: %v\n", err)
	}
	if !cmd.Options.Dry {
		cmd.Stats()
	}
}
