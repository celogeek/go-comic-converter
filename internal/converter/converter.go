/*
Converter Helper to parse and prepare options for go-comic-converter.

It use goflag with additional feature:
  - Keep original order
  - Support section
*/
package converter

import (
	"encoding/json"
	"errors"
	"flag"
	"fmt"
	"os"
	"path/filepath"
	"regexp"
	"runtime"
	"strings"
	"time"
)

type converter struct {
	Options *converterOptions
	Cmd     *flag.FlagSet

	order           []converterOrder
	isZeroValueErrs []error
	startAt         time.Time
}

// Create a new parser
func New() *converter {
	options := newOptions()
	cmd := flag.NewFlagSet("go-comic-converter", flag.ExitOnError)
	conv := &converter{
		Options: options,
		Cmd:     cmd,
		order:   make([]converterOrder, 0),
		startAt: time.Now(),
	}

	var cmdOutput strings.Builder
	cmd.SetOutput(&cmdOutput)
	cmd.Usage = func() {
		fmt.Fprintf(os.Stderr, "Usage of %s:\n", filepath.Base(os.Args[0]))
		for _, o := range conv.order {
			switch v := o.(type) {
			case converterOrderSection:
				fmt.Fprintf(os.Stderr, "\n%s:\n", o.Value())
			case converterOrderName:
				fmt.Fprintln(os.Stderr, conv.Usage(v.isString, cmd.Lookup(v.Value())))
			}
		}
		if cmdOutput.Len() > 0 {
			fmt.Fprintf(os.Stderr, "\nError: %s", cmdOutput.String())
		}
	}

	return conv
}

// Load default options (config + default)
func (c *converter) LoadConfig() error {
	return c.Options.LoadConfig()
}

// Initialize the parser with all section and parameter.
func (c *converter) InitParse() {
	c.addSection("Output")
	c.addStringParam(&c.Options.Input, "input", "", "Source of comic to convert: directory, cbz, zip, cbr, rar, pdf")
	c.addStringParam(&c.Options.Output, "output", "", "Output of the EPUB (directory or EPUB): (default [INPUT].epub)")
	c.addStringParam(&c.Options.Author, "author", "GO Comic Converter", "Author of the EPUB")
	c.addStringParam(&c.Options.Title, "title", "", "Title of the EPUB")

	c.addSection("Config")
	c.addStringParam(&c.Options.Profile, "profile", c.Options.Profile, fmt.Sprintf("Profile to use: \n%s", c.Options.AvailableProfiles()))
	c.addIntParam(&c.Options.Quality, "quality", c.Options.Quality, "Quality of the image")
	c.addBoolParam(&c.Options.Grayscale, "grayscale", c.Options.Grayscale, "Grayscale image. Ideal for eInk devices.")
	c.addIntParam(&c.Options.GrayscaleMode, "grayscale-mode", c.Options.GrayscaleMode, "Grayscale Mode\n0 = normal\n1 = average\n2 = luminance")
	c.addBoolParam(&c.Options.Crop, "crop", c.Options.Crop, "Crop images")
	c.addIntParam(&c.Options.CropRatioLeft, "crop-ratio-left", c.Options.CropRatioLeft, "Crop ratio left: ratio of pixels allow to be non blank while cutting on the left.")
	c.addIntParam(&c.Options.CropRatioUp, "crop-ratio-up", c.Options.CropRatioUp, "Crop ratio up: ratio of pixels allow to be non blank while cutting on the top.")
	c.addIntParam(&c.Options.CropRatioRight, "crop-ratio-right", c.Options.CropRatioRight, "Crop ratio right: ratio of pixels allow to be non blank while cutting on the right.")
	c.addIntParam(&c.Options.CropRatioBottom, "crop-ratio-bottom", c.Options.CropRatioBottom, "Crop ratio bottom: ratio of pixels allow to be non blank while cutting on the bottom.")
	c.addIntParam(&c.Options.Brightness, "brightness", c.Options.Brightness, "Brightness readjustement: between -100 and 100, > 0 lighter, < 0 darker")
	c.addIntParam(&c.Options.Contrast, "contrast", c.Options.Contrast, "Contrast readjustement: between -100 and 100, > 0 more contrast, < 0 less contrast")
	c.addBoolParam(&c.Options.AutoContrast, "autocontrast", c.Options.AutoContrast, "Improve contrast automatically")
	c.addBoolParam(&c.Options.AutoRotate, "autorotate", c.Options.AutoRotate, "Auto Rotate page when width > height")
	c.addBoolParam(&c.Options.AutoSplitDoublePage, "autosplitdoublepage", c.Options.AutoSplitDoublePage, "Auto Split double page when width > height")
	c.addBoolParam(&c.Options.KeepDoublePageIfSplitted, "keepdoublepageifsplitted", c.Options.KeepDoublePageIfSplitted, "Keep the double page if splitted")
	c.addBoolParam(&c.Options.NoBlankImage, "noblankimage", c.Options.NoBlankImage, "Remove blank image")
	c.addBoolParam(&c.Options.Manga, "manga", c.Options.Manga, "Manga mode (right to left)")
	c.addBoolParam(&c.Options.HasCover, "hascover", c.Options.HasCover, "Has cover. Indicate if your comic have a cover. The first page will be used as a cover and include after the title.")
	c.addIntParam(&c.Options.LimitMb, "limitmb", c.Options.LimitMb, "Limit size of the EPUB: Default nolimit (0), Minimum 20")
	c.addBoolParam(&c.Options.StripFirstDirectoryFromToc, "strip", c.Options.StripFirstDirectoryFromToc, "Strip first directory from the TOC if only 1")
	c.addIntParam(&c.Options.SortPathMode, "sort", c.Options.SortPathMode, "Sort path mode\n0 = alpha for path and file\n1 = alphanum for path and alpha for file\n2 = alphanum for path and file")
	c.addStringParam(&c.Options.ForegroundColor, "foreground-color", c.Options.ForegroundColor, "Foreground color in hexa format RGB. Black=000, White=FFF")
	c.addStringParam(&c.Options.BackgroundColor, "background-color", c.Options.BackgroundColor, "Background color in hexa format RGB. Black=000, White=FFF, Light Gray=DDD, Dark Gray=777")
	c.addBoolParam(&c.Options.NoResize, "noresize", c.Options.NoResize, "Do not reduce image size if exceed device size")
	c.addStringParam(&c.Options.Format, "format", c.Options.Format, "Format of output images: jpeg (lossy), png (lossless)")
	c.addFloatParam(&c.Options.AspectRatio, "aspect-ratio", c.Options.AspectRatio, "Aspect ratio (height/width) of the output\n -1 = same as device\n  0 = same as source\n1.6 = amazon advice for kindle")
	c.addBoolParam(&c.Options.PortraitOnly, "portrait-only", c.Options.PortraitOnly, "Portrait only: force orientation to portrait only.")
	c.addIntParam(&c.Options.TitlePage, "titlepage", c.Options.TitlePage, "Title page\n0 = never\n1 = always\n2 = only if epub is splitted")

	c.addSection("Default config")
	c.addBoolParam(&c.Options.Show, "show", false, "Show your default parameters")
	c.addBoolParam(&c.Options.Save, "save", false, "Save your parameters as default")
	c.addBoolParam(&c.Options.Reset, "reset", false, "Reset your parameters to default")

	c.addSection("Shortcut")
	c.addBoolParam(&c.Options.Auto, "auto", false, "Activate all automatic options")
	c.addBoolParam(&c.Options.NoFilter, "nofilter", false, "Deactivate all filters")
	c.addBoolParam(&c.Options.MaxQuality, "maxquality", false, "Max quality: color png + noresize")
	c.addBoolParam(&c.Options.BestQuality, "bestquality", false, "Max quality: color jpg q100 + noresize")
	c.addBoolParam(&c.Options.GreatQuality, "greatquality", false, "Max quality: grayscale jpg q90 + noresize")
	c.addBoolParam(&c.Options.GoodQuality, "goodquality", false, "Max quality: grayscale jpg q90")

	c.addSection("Compatibility")
	c.addBoolParam(&c.Options.AppleBookCompatibility, "applebookcompatibility", c.Options.AppleBookCompatibility, "Apple book compatibility")

	c.addSection("Other")
	c.addIntParam(&c.Options.Workers, "workers", runtime.NumCPU(), "Number of workers")
	c.addBoolParam(&c.Options.Dry, "dry", false, "Dry run to show all options")
	c.addBoolParam(&c.Options.DryVerbose, "dry-verbose", false, "Display also sorted files after the TOC")
	c.addBoolParam(&c.Options.Quiet, "quiet", false, "Disable progress bar")
	c.addBoolParam(&c.Options.Json, "json", false, "Output progression and information in Json format")
	c.addBoolParam(&c.Options.Version, "version", false, "Show current and available version")
	c.addBoolParam(&c.Options.Help, "help", false, "Show this help message")
}

// Parse all parameters
func (c *converter) Parse() {
	c.Cmd.Parse(os.Args[1:])
	if c.Options.Help {
		c.Cmd.Usage()
		os.Exit(0)
	}

	if c.Options.Auto {
		c.Options.AutoContrast = true
		c.Options.AutoRotate = true
		c.Options.AutoSplitDoublePage = true
	}

	if c.Options.MaxQuality {
		c.Options.Format = "png"
		c.Options.Grayscale = false
		c.Options.NoResize = true
	} else if c.Options.BestQuality {
		c.Options.Format = "jpeg"
		c.Options.Quality = 100
		c.Options.Grayscale = false
		c.Options.NoResize = true
	} else if c.Options.GreatQuality {
		c.Options.Format = "jpeg"
		c.Options.Quality = 90
		c.Options.Grayscale = true
		c.Options.NoResize = true
	} else if c.Options.GoodQuality {
		c.Options.Format = "jpeg"
		c.Options.Quality = 90
		c.Options.Grayscale = true
		c.Options.NoResize = false
	}

	if c.Options.NoFilter {
		c.Options.Crop = false
		c.Options.Brightness = 0
		c.Options.Contrast = 0
		c.Options.AutoContrast = false
		c.Options.AutoRotate = false
		c.Options.NoBlankImage = false
		c.Options.NoResize = true
	}

	if c.Options.AppleBookCompatibility {
		c.Options.AutoSplitDoublePage = true
		c.Options.KeepDoublePageIfSplitted = false
	}
}

// Check parameters
func (c *converter) Validate() error {
	// Check input
	if c.Options.Input == "" {
		return errors.New("missing input")
	}

	fi, err := os.Stat(c.Options.Input)
	if err != nil {
		return err
	}

	// Check Output
	var defaultOutput string
	inputBase := filepath.Clean(c.Options.Input)
	if fi.IsDir() {
		defaultOutput = fmt.Sprintf("%s.epub", inputBase)
	} else {
		ext := filepath.Ext(inputBase)
		defaultOutput = fmt.Sprintf("%s.epub", inputBase[0:len(inputBase)-len(ext)])
	}

	if c.Options.Output == "" {
		c.Options.Output = defaultOutput
	}

	c.Options.Output = filepath.Clean(c.Options.Output)
	if filepath.Ext(c.Options.Output) == ".epub" {
		fo, err := os.Stat(filepath.Dir(c.Options.Output))
		if err != nil {
			return err
		}
		if !fo.IsDir() {
			return errors.New("parent of the output is not a directory")
		}
	} else {
		fo, err := os.Stat(c.Options.Output)
		if err != nil {
			return err
		}
		if !fo.IsDir() {
			return errors.New("output must be an existing dir or end with .epub")
		}
		c.Options.Output = filepath.Join(
			c.Options.Output,
			filepath.Base(defaultOutput),
		)
	}

	// Title
	if c.Options.Title == "" {
		ext := filepath.Ext(defaultOutput)
		c.Options.Title = filepath.Base(defaultOutput[0 : len(defaultOutput)-len(ext)])
	}

	// Profile
	if c.Options.Profile == "" {
		return errors.New("profile missing")
	}

	if p := c.Options.GetProfile(); p == nil {
		return fmt.Errorf("profile %q doesn't exists", c.Options.Profile)
	}

	// LimitMb
	if c.Options.LimitMb < 20 && c.Options.LimitMb != 0 {
		return errors.New("limitmb should be 0 or >= 20")
	}

	// Brightness
	if c.Options.Brightness < -100 || c.Options.Brightness > 100 {
		return errors.New("brightness should be between -100 and 100")
	}

	// Contrast
	if c.Options.Contrast < -100 || c.Options.Contrast > 100 {
		return errors.New("contrast should be between -100 and 100")
	}

	// SortPathMode
	if c.Options.SortPathMode < 0 || c.Options.SortPathMode > 2 {
		return errors.New("sort should be 0, 1 or 2")
	}

	// Color
	colorRegex := regexp.MustCompile("^[0-9A-F]{3}$")
	if !colorRegex.MatchString(c.Options.ForegroundColor) {
		return errors.New("foreground color must have color format in hexa: [0-9A-F]{3}")
	}

	if !colorRegex.MatchString(c.Options.BackgroundColor) {
		return errors.New("background color must have color format in hexa: [0-9A-F]{3}")
	}

	// Format
	if !(c.Options.Format == "jpeg" || c.Options.Format == "png") {
		return errors.New("format should be jpeg or png")
	}

	// Aspect Ratio
	if c.Options.AspectRatio < 0 && c.Options.AspectRatio != -1 {
		return errors.New("aspect ratio should be -1, 0 or > 0")
	}

	// Title Page
	if c.Options.TitlePage < 0 || c.Options.TitlePage > 2 {
		return errors.New("title page should be 0, 1 or 2")
	}

	// Grayscale Mode
	if c.Options.GrayscaleMode < 0 || c.Options.GrayscaleMode > 2 {
		return errors.New("grayscale mode should be 0, 1 or 2")
	}

	return nil
}

// Customize version of FlagSet.PrintDefaults
func (c *converter) Usage(isString bool, f *flag.Flag) string {
	var b strings.Builder
	fmt.Fprintf(&b, "  -%s", f.Name) // Two spaces before -; see next two comments.
	name, usage := flag.UnquoteUsage(f)
	if len(name) > 0 {
		b.WriteString(" ")
		b.WriteString(name)
	}
	// Print the default value only if it differs to the zero value
	// for this flag type.
	if isZero, err := c.isZeroValue(f, f.DefValue); err != nil {
		c.isZeroValueErrs = append(c.isZeroValueErrs, err)
	} else if !isZero {
		if isString {
			fmt.Fprintf(&b, " (default %q)", f.DefValue)
		} else {
			fmt.Fprintf(&b, " (default %v)", f.DefValue)
		}
	}

	// Boolean flags of one ASCII letter are so common we
	// treat them specially, putting their usage on the same line.
	if b.Len() <= 4 { // space, space, '-', 'x'.
		b.WriteString("\t")
	} else {
		// Four spaces before the tab triggers good alignment
		// for both 4- and 8-space tab stops.
		b.WriteString("\n    \t")
	}
	b.WriteString(strings.ReplaceAll(usage, "\n", "\n    \t"))

	return b.String()
}

// Helper to show usage, err and exit 1
func (c *converter) Fatal(err error) {
	c.Cmd.Usage()
	fmt.Fprintf(os.Stderr, "\nError: %s\n", err)
	os.Exit(1)
}

func (c *converter) Stats() {
	// Display elapse time and memory usage
	elapse := time.Since(c.startAt).Round(time.Millisecond)
	var mem runtime.MemStats
	runtime.ReadMemStats(&mem)

	if c.Options.Json {
		json.NewEncoder(os.Stdout).Encode(map[string]any{
			"type": "stats",
			"data": map[string]any{
				"elapse_ms":       elapse.Milliseconds(),
				"memory_usage_mb": mem.Sys / 1024 / 1024,
			},
		})
	} else {
		fmt.Fprintf(
			os.Stderr,
			"Completed in %s, Memory usage %d Mb\n",
			elapse,
			mem.Sys/1024/1024,
		)
	}
}
