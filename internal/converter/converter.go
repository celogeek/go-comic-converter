/*
Converter Helper to parse and prepare options for go-comic-converter.

It use goflag with additional feature:
  - Keep original order
  - Support section
*/
package converter

import (
	"errors"
	"flag"
	"fmt"
	"os"
	"path/filepath"
	"reflect"
	"regexp"
	"runtime"
	"strings"
	"time"

	"github.com/celogeek/go-comic-converter/v2/internal/converter/options"
)

type Converter struct {
	Options *options.Options
	Cmd     *flag.FlagSet

	order           []converterOrder
	isZeroValueErrs []error
	startAt         time.Time
}

// Create a new parser
func New() *Converter {
	options := options.New()
	cmd := flag.NewFlagSet("go-comic-converter", flag.ExitOnError)
	conv := &Converter{
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
func (c *Converter) LoadConfig() error {
	return c.Options.LoadConfig()
}

// Create a new section of config
func (c *Converter) AddSection(section string) {
	c.order = append(c.order, converterOrderSection{value: section})
}

// Add a string parameter
func (c *Converter) AddStringParam(p *string, name string, value string, usage string) {
	c.Cmd.StringVar(p, name, value, usage)
	c.order = append(c.order, converterOrderName{value: name, isString: true})
}

// Add an integer parameter
func (c *Converter) AddIntParam(p *int, name string, value int, usage string) {
	c.Cmd.IntVar(p, name, value, usage)
	c.order = append(c.order, converterOrderName{value: name})
}

// Add an float parameter
func (c *Converter) AddFloatParam(p *float64, name string, value float64, usage string) {
	c.Cmd.Float64Var(p, name, value, usage)
	c.order = append(c.order, converterOrderName{value: name})
}

// Add a boolean parameter
func (c *Converter) AddBoolParam(p *bool, name string, value bool, usage string) {
	c.Cmd.BoolVar(p, name, value, usage)
	c.order = append(c.order, converterOrderName{value: name})
}

// Initialize the parser with all section and parameter.
func (c *Converter) InitParse() {
	c.AddSection("Output")
	c.AddStringParam(&c.Options.Input, "input", "", "Source of comic to convert: directory, cbz, zip, cbr, rar, pdf")
	c.AddStringParam(&c.Options.Output, "output", "", "Output of the EPUB (directory or EPUB): (default [INPUT].epub)")
	c.AddStringParam(&c.Options.Author, "author", "GO Comic Converter", "Author of the EPUB")
	c.AddStringParam(&c.Options.Title, "title", "", "Title of the EPUB")

	c.AddSection("Config")
	c.AddStringParam(&c.Options.Profile, "profile", c.Options.Profile, fmt.Sprintf("Profile to use: \n%s", c.Options.AvailableProfiles()))
	c.AddIntParam(&c.Options.Quality, "quality", c.Options.Quality, "Quality of the image")
	c.AddBoolParam(&c.Options.Grayscale, "grayscale", c.Options.Grayscale, "Grayscale image. Ideal for eInk devices.")
	c.AddIntParam(&c.Options.GrayscaleMode, "grayscale-mode", c.Options.GrayscaleMode, "Grayscale Mode\n0 = normal\n1 = average\n2 = luminance")
	c.AddBoolParam(&c.Options.Crop, "crop", c.Options.Crop, "Crop images")
	c.AddIntParam(&c.Options.CropRatioLeft, "crop-ratio-left", c.Options.CropRatioLeft, "Crop ratio left: ratio of pixels allow to be non blank while cutting on the left.")
	c.AddIntParam(&c.Options.CropRatioUp, "crop-ratio-up", c.Options.CropRatioUp, "Crop ratio up: ratio of pixels allow to be non blank while cutting on the top.")
	c.AddIntParam(&c.Options.CropRatioRight, "crop-ratio-right", c.Options.CropRatioRight, "Crop ratio right: ratio of pixels allow to be non blank while cutting on the right.")
	c.AddIntParam(&c.Options.CropRatioBottom, "crop-ratio-bottom", c.Options.CropRatioBottom, "Crop ratio bottom: ratio of pixels allow to be non blank while cutting on the bottom.")
	c.AddIntParam(&c.Options.Brightness, "brightness", c.Options.Brightness, "Brightness readjustement: between -100 and 100, > 0 lighter, < 0 darker")
	c.AddIntParam(&c.Options.Contrast, "contrast", c.Options.Contrast, "Contrast readjustement: between -100 and 100, > 0 more contrast, < 0 less contrast")
	c.AddBoolParam(&c.Options.AutoContrast, "autocontrast", c.Options.AutoContrast, "Improve contrast automatically")
	c.AddBoolParam(&c.Options.AutoRotate, "autorotate", c.Options.AutoRotate, "Auto Rotate page when width > height")
	c.AddBoolParam(&c.Options.AutoSplitDoublePage, "autosplitdoublepage", c.Options.AutoSplitDoublePage, "Auto Split double page when width > height")
	c.AddBoolParam(&c.Options.KeepDoublePageIfSplitted, "keepdoublepageifsplitted", c.Options.KeepDoublePageIfSplitted, "Keep the double page if splitted")
	c.AddBoolParam(&c.Options.NoBlankImage, "noblankimage", c.Options.NoBlankImage, "Remove blank image")
	c.AddBoolParam(&c.Options.Manga, "manga", c.Options.Manga, "Manga mode (right to left)")
	c.AddBoolParam(&c.Options.HasCover, "hascover", c.Options.HasCover, "Has cover. Indicate if your comic have a cover. The first page will be used as a cover and include after the title.")
	c.AddIntParam(&c.Options.LimitMb, "limitmb", c.Options.LimitMb, "Limit size of the EPUB: Default nolimit (0), Minimum 20")
	c.AddBoolParam(&c.Options.StripFirstDirectoryFromToc, "strip", c.Options.StripFirstDirectoryFromToc, "Strip first directory from the TOC if only 1")
	c.AddIntParam(&c.Options.SortPathMode, "sort", c.Options.SortPathMode, "Sort path mode\n0 = alpha for path and file\n1 = alphanum for path and alpha for file\n2 = alphanum for path and file")
	c.AddStringParam(&c.Options.ForegroundColor, "foreground-color", c.Options.ForegroundColor, "Foreground color in hexa format RGB. Black=000, White=FFF")
	c.AddStringParam(&c.Options.BackgroundColor, "background-color", c.Options.BackgroundColor, "Background color in hexa format RGB. Black=000, White=FFF, Light Gray=DDD, Dark Gray=777")
	c.AddBoolParam(&c.Options.NoResize, "noresize", c.Options.NoResize, "Do not reduce image size if exceed device size")
	c.AddStringParam(&c.Options.Format, "format", c.Options.Format, "Format of output images: jpeg (lossy), png (lossless)")
	c.AddFloatParam(&c.Options.AspectRatio, "aspect-ratio", c.Options.AspectRatio, "Aspect ratio (height/width) of the output\n -1 = same as device\n  0 = same as source\n1.6 = amazon advice for kindle")
	c.AddBoolParam(&c.Options.PortraitOnly, "portrait-only", c.Options.PortraitOnly, "Portrait only: force orientation to portrait only.")
	c.AddIntParam(&c.Options.TitlePage, "titlepage", c.Options.TitlePage, "Title page\n0 = never\n1 = always\n2 = only if epub is splitted")

	c.AddSection("Default config")
	c.AddBoolParam(&c.Options.Show, "show", false, "Show your default parameters")
	c.AddBoolParam(&c.Options.Save, "save", false, "Save your parameters as default")
	c.AddBoolParam(&c.Options.Reset, "reset", false, "Reset your parameters to default")

	c.AddSection("Shortcut")
	c.AddBoolParam(&c.Options.Auto, "auto", false, "Activate all automatic options")
	c.AddBoolParam(&c.Options.NoFilter, "nofilter", false, "Deactivate all filters")
	c.AddBoolParam(&c.Options.MaxQuality, "maxquality", false, "Max quality: color png + noresize")
	c.AddBoolParam(&c.Options.BestQuality, "bestquality", false, "Max quality: color jpg q100 + noresize")
	c.AddBoolParam(&c.Options.GreatQuality, "greatquality", false, "Max quality: grayscale jpg q90 + noresize")
	c.AddBoolParam(&c.Options.GoodQuality, "goodquality", false, "Max quality: grayscale jpg q90")

	c.AddSection("Compatibility")
	c.AddBoolParam(&c.Options.AppleBookCompatibility, "applebookcompatibility", c.Options.AppleBookCompatibility, "Apple book compatibility")

	c.AddSection("Other")
	c.AddIntParam(&c.Options.Workers, "workers", runtime.NumCPU(), "Number of workers")
	c.AddBoolParam(&c.Options.Dry, "dry", false, "Dry run to show all options")
	c.AddBoolParam(&c.Options.DryVerbose, "dry-verbose", false, "Display also sorted files after the TOC")
	c.AddBoolParam(&c.Options.Quiet, "quiet", false, "Disable progress bar")
	c.AddBoolParam(&c.Options.Version, "version", false, "Show current and available version")
	c.AddBoolParam(&c.Options.Help, "help", false, "Show this help message")
}

// Customize version of FlagSet.PrintDefaults
func (c *Converter) Usage(isString bool, f *flag.Flag) string {
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

// Taken from flag package as it is private and needed for usage.
//
// isZeroValue determines whether the string represents the zero
// value for a flag.
func (c *Converter) isZeroValue(f *flag.Flag, value string) (ok bool, err error) {
	// Build a zero value of the flag's Value type, and see if the
	// result of calling its String method equals the value passed in.
	// This works unless the Value type is itself an interface type.
	typ := reflect.TypeOf(f.Value)
	var z reflect.Value
	if typ.Kind() == reflect.Pointer {
		z = reflect.New(typ.Elem())
	} else {
		z = reflect.Zero(typ)
	}
	// Catch panics calling the String method, which shouldn't prevent the
	// usage message from being printed, but that we should report to the
	// user so that they know to fix their code.
	defer func() {
		if e := recover(); e != nil {
			if typ.Kind() == reflect.Pointer {
				typ = typ.Elem()
			}
			err = fmt.Errorf("panic calling String method on zero %v for flag %s: %v", typ, f.Name, e)
		}
	}()
	return value == z.Interface().(flag.Value).String(), nil
}

// Parse all parameters
func (c *Converter) Parse() {
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
func (c *Converter) Validate() error {
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

// Helper to show usage, err and exit 1
func (c *Converter) Fatal(err error) {
	c.Cmd.Usage()
	fmt.Fprintf(os.Stderr, "\nError: %s\n", err)
	os.Exit(1)
}

func (c *Converter) Stats() {
	// Display elapse time and memory usage
	var mem runtime.MemStats
	runtime.ReadMemStats(&mem)
	fmt.Fprintf(
		os.Stderr,
		"Completed in %s, Memory usage %d Mb\n",
		time.Since(c.startAt).Round(time.Millisecond),
		mem.Sys/1024/1024,
	)
}
