// Package epubprogress create a progress bar with custom settings.
package epubprogress

import (
	"encoding/json"
	"fmt"
	"os"
	"time"

	"github.com/schollz/progressbar/v3"

	"github.com/celogeek/go-comic-converter/v3/internal/pkg/utils"
)

type Options struct {
	Quiet       bool
	Json        bool
	Max         int
	Description string
	CurrentJob  int
	TotalJob    int
}

type EPUBProgress interface {
	Add(num int) error
	Close() error
}

func New(o Options) EPUBProgress {
	if o.Quiet {
		return progressbar.DefaultSilent(int64(o.Max))
	}

	if o.Json {
		return &jsonprogress{
			o: o,
			e: json.NewEncoder(os.Stdout),
		}
	}

	fmtJob := utils.FormatNumberOfDigits(o.TotalJob)
	return progressbar.NewOptions(o.Max,
		progressbar.OptionSetWriter(os.Stderr),
		progressbar.OptionThrottle(65*time.Millisecond),
		progressbar.OptionOnCompletion(func() {
			utils.Println()
		}),
		progressbar.OptionSetDescription(fmt.Sprintf(
			"["+fmtJob+"/"+fmtJob+"] %-15s",
			o.CurrentJob,
			o.TotalJob,
			o.Description,
		)),
		progressbar.OptionSetWidth(60),
		progressbar.OptionShowCount(),
		progressbar.OptionSetRenderBlankState(true),
		progressbar.OptionEnableColorCodes(true),
		progressbar.OptionSetTheme(progressbar.Theme{
			Saucer:        "[green]=[reset]",
			SaucerHead:    "[green]>[reset]",
			SaucerPadding: " ",
			BarStart:      "[",
			BarEnd:        "]",
		}),
	)
}
