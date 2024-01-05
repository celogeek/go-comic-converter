/*
create a progress bar with custom settings.
*/
package epubprogress

import (
	"encoding/json"
	"fmt"
	"os"
	"time"

	"github.com/schollz/progressbar/v3"
)

type Options struct {
	Quiet       bool
	Json        bool
	Max         int
	Description string
	CurrentJob  int
	TotalJob    int
}

type epubProgress interface {
	Add(num int) error
	Close() error
}

func New(o Options) epubProgress {
	if o.Quiet {
		return progressbar.DefaultSilent(int64(o.Max))
	}

	if o.Json {
		return &epubProgressJson{
			o: o,
			e: json.NewEncoder(os.Stdout),
		}
	}

	fmtJob := fmt.Sprintf("%%0%dd", len(fmt.Sprint(o.TotalJob)))
	fmtDesc := fmt.Sprintf("[%s/%s] %%-15s", fmtJob, fmtJob)
	return progressbar.NewOptions(o.Max,
		progressbar.OptionSetWriter(os.Stderr),
		progressbar.OptionThrottle(65*time.Millisecond),
		progressbar.OptionOnCompletion(func() {
			fmt.Fprint(os.Stderr, "\n")
		}),
		progressbar.OptionSetDescription(fmt.Sprintf(fmtDesc, o.CurrentJob, o.TotalJob, o.Description)),
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
