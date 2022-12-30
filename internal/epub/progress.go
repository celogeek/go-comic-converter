package epub

import (
	"fmt"
	"os"

	"github.com/schollz/progressbar/v3"
)

func NewBar(max int, description string, currentJob, totalJob int) *progressbar.ProgressBar {
	fmtJob := fmt.Sprintf("%%0%dd", len(fmt.Sprint(totalJob)))
	fmtDesc := fmt.Sprintf("[%s/%s] %%-15s", fmtJob, fmtJob)
	return progressbar.NewOptions(max,
		progressbar.OptionSetWriter(os.Stderr),
		progressbar.OptionOnCompletion(func() {
			fmt.Fprint(os.Stderr, "\n")
		}),
		progressbar.OptionSetDescription(fmt.Sprintf(fmtDesc, currentJob, totalJob, description)),
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
