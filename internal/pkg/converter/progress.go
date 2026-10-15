package converter

import (
	"encoding/json"
	"fmt"
	"os"
	"sync"
	"time"

	"github.com/schollz/progressbar/v3"

	"github.com/celogeek/go-comic-converter/v3/internal/pkg/utils"
	"github.com/celogeek/go-comic-converter/v3/pkg/epuboptions"
)

// jobProgress follows a job, Set receives the number of items done.
type jobProgress interface {
	Set(num int) error
	Close() error
}

type jsonProgress struct {
	e           *json.Encoder
	total       int
	description string
	totalJob    int
	currentJob  int
}

func (p *jsonProgress) Set(num int) error {
	return p.e.Encode(map[string]any{
		"type": "epubprogress",
		"data": map[string]any{
			"epubprogress": map[string]any{
				"current": num,
				"total":   p.total,
			},
			"steps": map[string]any{
				"current": p.currentJob,
				"total":   p.totalJob,
			},
			"description": p.description,
		},
	})
}

func (p *jsonProgress) Close() error {
	return p.Set(p.total)
}

// newJobProgress creates the progress of a job, depending on the output mode.
func (o *Options) newJobProgress(total int, description string, currentJob, totalJob int) jobProgress {
	if o.Quiet {
		return progressbar.DefaultSilent(int64(total))
	}

	if o.Json {
		return &jsonProgress{
			e:           json.NewEncoder(os.Stdout),
			total:       total,
			description: description,
			currentJob:  currentJob,
			totalJob:    totalJob,
		}
	}

	fmtJob := utils.FormatNumberOfDigits(totalJob)
	return progressbar.NewOptions(total,
		progressbar.OptionSetWriter(os.Stderr),
		progressbar.OptionThrottle(65*time.Millisecond),
		progressbar.OptionOnCompletion(func() {
			utils.Println()
		}),
		progressbar.OptionSetDescription(fmt.Sprintf(
			"["+fmtJob+"/"+fmtJob+"] %-15s",
			currentJob,
			totalJob,
			description,
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

// jobCallback builds the callback of a job: the progress is created at the first call as the total is given there.
// It only moves forward, and it is closed once all the items are done.
func (o *Options) jobCallback(description string, currentJob, totalJob int, onFinish func()) epuboptions.Progress {
	var (
		mu       sync.Mutex
		progress jobProgress
		current  int
		finished bool
	)
	return func(done, total int) {
		mu.Lock()
		defer mu.Unlock()
		if finished {
			return
		}
		if progress == nil {
			progress = o.newJobProgress(total, description, currentJob, totalJob)
		}
		if done >= total {
			finished = true
			_ = progress.Close()
			if onFinish != nil {
				onFinish()
			}
			return
		}
		if done > current {
			current = done
			_ = progress.Set(done)
		}
	}
}

// BindProgress follows the processing and the writing of the epub with the progress bars of the command line.
func (o *Options) BindProgress() {
	description := "Processing"
	if o.Image.Format == "copy" {
		description = "Copying"
	}

	o.OnProcess = o.jobCallback(description, 1, 2, nil)
	o.OnWrite = o.jobCallback("Writing Part", 2, 2, func() {
		if !o.Json {
			utils.Println()
		}
	})
}
