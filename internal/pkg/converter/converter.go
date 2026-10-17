// Package converter Helper to parse and prepare options for go-comic-converter.
//
// It uses goflag with additional feature:
//   - Keep original order
//   - Support section
package converter

import (
	"encoding/json"
	"errors"
	"flag"
	"fmt"
	"os"
	"path/filepath"
	"reflect"
	"regexp"
	"runtime"
	"slices"
	"strconv"
	"strings"
	"time"

	"github.com/gen2brain/jxl"

	"github.com/celogeek/go-comic-converter/v3/internal/pkg/utils"
	"github.com/celogeek/go-comic-converter/v3/pkg/epuboptions"
)

type Converter struct {
	Options *Options
	Cmd     *flag.FlagSet

	order           []order
	reversibleBool  map[string]struct{}
	isZeroValueErrs []error
	startAt         time.Time
}

// New Create a new parser
func New() *Converter {
	o := NewOptions()
	cmd := flag.NewFlagSet("go-comic-converter", flag.ExitOnError)
	conv := &Converter{
		Options:        o,
		Cmd:            cmd,
		order:          make([]order, 0),
		reversibleBool: make(map[string]struct{}),
		startAt:        time.Now(),
	}

	var cmdOutput strings.Builder
	cmd.SetOutput(&cmdOutput)
	cmd.Usage = func() {
		utils.Printf("Usage of %s:\n", filepath.Base(os.Args[0]))
		for _, o := range conv.order {
			switch v := o.(type) {
			case orderSection:
				utils.Printf("\n%s:\n", o.Value())
			case orderName:
				utils.Println(conv.Usage(v.isString, cmd.Lookup(v.Value())))
			}
		}
		if cmdOutput.Len() > 0 {
			utils.Printf("\nError: %s", cmdOutput.String())
		}
	}

	return conv
}

// LoadConfig Load default options (config + default)
//
// The saved config is skipped with -ignore-config: the flags take their default from the options,
// so it has to be known before they are declared (and the help shows the default values).
func (c *Converter) LoadConfig() error {
	if ignoreConfig(os.Args[1:]) {
		return nil
	}
	return c.Options.LoadConfig()
}

// ignoreConfig tells if the arguments ask to ignore the saved config (-ignore-config, -ignore-config=0|1, the last one wins).
func ignoreConfig(args []string) (ignore bool) {
	for _, a := range args {
		if a == "--" {
			break
		}
		name, value, hasValue := strings.Cut(strings.TrimLeft(a, "-"), "=")
		if name != "ignore-config" || !strings.HasPrefix(a, "-") {
			continue
		}
		ignore = true
		if hasValue {
			if v, err := strconv.ParseBool(value); err == nil {
				ignore = v
			}
		}
	}
	return ignore
}

// AddSection Create a new section of config
func (c *Converter) AddSection(section string) {
	c.order = append(c.order, orderSection{value: section})
}

// AddStringParam Add a string parameter
func (c *Converter) AddStringParam(p *string, name string, value string, usage string) {
	c.Cmd.StringVar(p, name, value, usage)
	c.order = append(c.order, orderName{value: name, isString: true})
}

// AddIntParam Add an integer parameter
func (c *Converter) AddIntParam(p *int, name string, value int, usage string) {
	c.Cmd.IntVar(p, name, value, usage)
	c.order = append(c.order, orderName{value: name})
}

// AddFloatParam Add an float parameter
func (c *Converter) AddFloatParam(p *float64, name string, value float64, usage string) {
	c.Cmd.Float64Var(p, name, value, usage)
	c.order = append(c.order, orderName{value: name})
}

// AddBoolParam Add a boolean parameter
func (c *Converter) AddBoolParam(p *bool, name string, value bool, usage string) {
	c.Cmd.BoolVar(p, name, value, usage)
	c.order = append(c.order, orderName{value: name})
}

// AddReversibleBoolParam Add a boolean parameter
func (c *Converter) AddReversibleBoolParam(p *bool, name string, value bool, usage string) {
	c.AddBoolParam(p, name, value, usage)
	c.reversibleBool[name] = struct{}{}
}

// InitParse Initialize the parser with all section and parameter.
func (c *Converter) InitParse() {
	c.AddSection(sectionOutput)
	c.AddStringParam(&c.Options.Input, "input", "", "Source of comic to convert: directory, cbz, zip, cbr, rar, pdf")
	c.AddStringParam(&c.Options.Output, "output", "", "Output of the EPUB (directory or EPUB): (default [INPUT].epub)")
	c.AddStringParam(&c.Options.Author, "author", "GO Comic Converter", "Author of the EPUB")
	c.AddStringParam(&c.Options.Title, "title", "", "Title of the EPUB")

	c.AddSection(sectionConfig)
	c.AddStringParam(&c.Options.Profile, "profile", c.Options.Profile, "Profile to use: \n"+c.Options.AvailableProfiles())
	c.AddIntParam(&c.Options.Image.Quality, "quality", c.Options.Image.Quality, "Quality of the image")
	c.AddIntParam(&c.Options.Image.JXLEffort, "jxl-effort", c.Options.Image.JXLEffort, "Effort of the jxl encoder, between 1 (fast) and 7 (smaller files)")
	c.AddReversibleBoolParam(&c.Options.Image.GrayScale, "grayscale", c.Options.Image.GrayScale, "Grayscale image. Ideal for eInk devices.")
	c.AddIntParam(&c.Options.Image.GrayScaleMode, "grayscale-mode", c.Options.Image.GrayScaleMode, "Grayscale Mode\n0 = normal\n1 = average\n2 = luminance")
	c.AddReversibleBoolParam(&c.Options.Image.Crop.Enabled, "crop", c.Options.Image.Crop.Enabled, "Crop images")
	c.AddIntParam(&c.Options.Image.Crop.Left, "crop-ratio-left", c.Options.Image.Crop.Left, "Crop ratio left: ratio of pixels allow to be non blank while cutting on the left.")
	c.AddIntParam(&c.Options.Image.Crop.Up, "crop-ratio-up", c.Options.Image.Crop.Up, "Crop ratio up: ratio of pixels allow to be non blank while cutting on the top.")
	c.AddIntParam(&c.Options.Image.Crop.Right, "crop-ratio-right", c.Options.Image.Crop.Right, "Crop ratio right: ratio of pixels allow to be non blank while cutting on the right.")
	c.AddIntParam(&c.Options.Image.Crop.Bottom, "crop-ratio-bottom", c.Options.Image.Crop.Bottom, "Crop ratio bottom: ratio of pixels allow to be non blank while cutting on the bottom.")
	c.AddIntParam(&c.Options.Image.Crop.Limit, "crop-limit", c.Options.Image.Crop.Limit, "Crop limit: maximum number of cropping in percentage allowed. 0 mean unlimited.")
	c.AddReversibleBoolParam(&c.Options.Image.Crop.SkipIfLimitReached, "crop-skip-if-limit-reached", c.Options.Image.Crop.SkipIfLimitReached, "Crop skip if limit reached.")
	c.AddIntParam(&c.Options.Image.Brightness, "brightness", c.Options.Image.Brightness, "Brightness readjustment: between -100 and 100, > 0 lighter, < 0 darker")
	c.AddIntParam(&c.Options.Image.Contrast, "contrast", c.Options.Image.Contrast, "Contrast readjustment: between -100 and 100, > 0 more contrast, < 0 less contrast")
	c.AddReversibleBoolParam(&c.Options.Image.AutoContrast, "autocontrast", c.Options.Image.AutoContrast, "Improve contrast automatically")
	c.AddIntParam(&c.Options.Image.DoublePageMode, "doublepagemode", c.Options.Image.DoublePageMode, "Double page mode (width > height)\n0 = split\n1 = split + rotate\n2 = rotate")
	c.AddReversibleBoolParam(&c.Options.Image.KeepSplitDoublePageAspect, "keepsplitdoublepageaspect", c.Options.Image.KeepSplitDoublePageAspect, "Keep aspect of split part of a double page (best for landscape rendering)")
	c.AddReversibleBoolParam(&c.Options.Image.CombineSplitDoublePage, "combinesplitdoublepage", c.Options.Image.CombineSplitDoublePage, "Detect a double page split in 2 consecutive pages, and combine them (the facing edges must match)")
	c.AddReversibleBoolParam(&c.Options.Image.NoBlankImage, "noblankimage", c.Options.Image.NoBlankImage, "Remove blank image")
	c.AddIntParam(&c.Options.Image.Mode, "mode", c.Options.Image.Mode, "Reading mode\n0 = comic (left to right)\n1 = manga (right to left), not supported by all the devices (the reMarkable for example)\n2 = webtoon: the images of each directory are joined, and cut into pages of the size of the view, in the gaps between the panels\n    the options about the double pages, the crop and the blank images are ignored")
	c.AddReversibleBoolParam(&c.Options.Image.Landscape, "landscape", c.Options.Image.Landscape, "Landscape only mode, Kindle Scribe only (an error with the other profiles): the pages are spread left and right, with a blank page to keep the double pages facing each other")
	c.AddIntParam(&c.Options.LimitMb, "limitmb", c.Options.LimitMb, "Limit size of the EPUB: Default limit (200), Minimum 20, 0 for unlimited")
	c.AddReversibleBoolParam(&c.Options.StripFirstDirectoryFromToc, "strip", c.Options.StripFirstDirectoryFromToc, "Strip first directory from the TOC if only 1")
	c.AddIntParam(&c.Options.SortPathMode, "sort", c.Options.SortPathMode, "Sort path mode\n0 = alpha for path and file\n1 = alphanumeric for path and alpha for file\n2 = alphanumeric for path and file")
	c.AddStringParam(&c.Options.Image.View.Color.Foreground, "foreground-color", c.Options.Image.View.Color.Foreground, "Foreground color in hexadecimal format RGB. Black=000000, White=FFFFFF")
	c.AddStringParam(&c.Options.Image.View.Color.Background, "background-color", c.Options.Image.View.Color.Background, "Background color in hexadecimal format RGB. Black=000000, White=FFFFFF, Light Gray=DDDDDD, Dark Gray=777777")
	c.AddReversibleBoolParam(&c.Options.Image.Resize, "resize", c.Options.Image.Resize, "Reduce image size if exceed device size")
	c.AddFloatParam(&c.Options.Image.Zoom, "zoom", c.Options.Image.Zoom, "Zoom of the images, between 1 and 3: they are stored bigger than the device for a better quality when you zoom\n  0 = zoom of the profile (1.5 for Apple Books, 1 for the others)\n  1 = size of the device\nThe size of the EPUB grows with the square of the zoom")
	c.AddStringParam(&c.Options.Image.Format, "format", c.Options.Image.Format, "Format of output images: jpeg (lossy), png (lossless), jxl (lossy, lossless with quality 100), copy (no processing)")
	c.AddFloatParam(&c.Options.Image.View.AspectRatio, "aspect-ratio", c.Options.Image.View.AspectRatio, "Aspect ratio (height/width) of the output, between 0.5 and 3\n -1 = same as device\n  0 = same as source, not supported by the webtoon mode\n1.6 = amazon advice for kindle")
	c.AddIntParam(&c.Options.TitlePage, "titlepage", c.Options.TitlePage, "Title page\n0 = never\n1 = always\n2 = only if epub is split")
	c.AddReversibleBoolParam(&c.Options.CoverCaption, "covercaption", c.Options.CoverCaption, "Set cover caption when the epub is split")
	c.AddStringParam(&c.Options.Font, "font", c.Options.Font, "Path to a TrueType font (.ttf) for the cover caption and title page. Default: embedded Literata")

	c.AddSection("Default config")
	c.AddBoolParam(&c.Options.Show, "show", false, "Show your default parameters")
	c.AddBoolParam(&c.Options.Save, "save", false, "Save your parameters as default")
	c.AddBoolParam(&c.Options.Reset, "reset", false, "Reset your parameters to default")
	c.AddBoolParam(&c.Options.ShowCommand, "show-command", false, "Show the command line equivalent to the current options: -ignore-config, the profile, and the options that differ from the default")
	c.AddBoolParam(&c.Options.IgnoreConfig, "ignore-config", false, "Ignore your saved parameters, and use the default ones (with -help to show them)")

	c.AddSection("Shortcut")
	c.AddBoolParam(&c.Options.NoFilter, "nofilter", false, "Deactivate all filters")
	c.AddBoolParam(&c.Options.MaxQuality, "maxquality", false, "Max quality: color png + noresize")
	c.AddBoolParam(&c.Options.BestQuality, "bestquality", false, "Max quality: color jpg q100 + noresize")
	c.AddBoolParam(&c.Options.GreatQuality, "greatquality", false, "Max quality: grayscale jpg q90 + noresize")
	c.AddBoolParam(&c.Options.GoodQuality, "goodquality", false, "Max quality: grayscale jpg q90")

	c.AddSection("Compatibility")

	c.AddSection("Other")
	c.AddIntParam(&c.Options.Workers, "workers", runtime.NumCPU(), "Number of workers")
	c.AddBoolParam(&c.Options.Dry, "dry", false, "Dry run to show all options")
	c.AddBoolParam(&c.Options.Quiet, "quiet", false, "Disable progress bar")
	c.AddBoolParam(&c.Options.Json, "json", false, "Output progression and information in Json format")
	c.AddBoolParam(&c.Options.Version, "version", false, "Show current and available version")
	c.AddBoolParam(&c.Options.Help, "help", false, "Show this help message")
}

// Usage Customize version of FlagSet.PrintDefaults
func (c *Converter) Usage(isString bool, f *flag.Flag) string {
	var b strings.Builder
	b.WriteString("  -")
	b.WriteString(f.Name)
	name, usage := flag.UnquoteUsage(f)
	if len(name) > 0 {
		b.WriteString("=")
		b.WriteString(name)
	} else {
		if _, ok := c.reversibleBool[f.Name]; ok {
			b.WriteString(" or -")
			b.WriteString(f.Name)
			b.WriteString("=1 (to enabled), -")
			b.WriteString(f.Name)
			b.WriteString("=0 (to disabled)")
		}
	}
	// Print the default value only if it differs to the zero value
	// for this flag type.
	if isZero, err := c.isZeroValue(f, f.DefValue); err != nil {
		c.isZeroValueErrs = append(c.isZeroValueErrs, err)
	} else if !isZero {
		if isString {
			b.WriteString(fmt.Sprintf(" (default %q)", f.DefValue))
		} else {
			if name == "" {
				b.WriteString(fmt.Sprintf(" (default %d)", utils.BoolStringToInt(f.DefValue)))
			} else {
				b.WriteString(fmt.Sprintf(" (default %v)", f.DefValue))
			}
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
	if err := c.Cmd.Parse(os.Args[1:]); err != nil {
		utils.Fatalf("cannot parse command line options: %v", err)
	}

	if len(c.Cmd.Args()) > 0 {
		arg := c.Cmd.Args()[0]
		fmt.Printf("After parsing command-line options, some of them are left unknown: %q\n", arg)
		fmt.Println("")
		if arg == "0" || arg == "1" || arg == "true" || arg == "false" {
			fmt.Println("It looks like you're trying to set a boolean flag. The right syntax is:")
			fmt.Println("  -boolean_option=0    (to disable)")
			fmt.Println("  -boolean_option=1    (to enable)")
			fmt.Println("")
			fmt.Println("Do NOT use: -boolean_option 0")
			fmt.Println("")
		}
		os.Exit(1)
	}

	if c.Options.Help {
		c.Cmd.Usage()
		os.Exit(0)
	}

	// The device and its features come only from the profile.
	if p := c.Options.GetProfile(); p != nil {
		c.Options.Image.Device = p.Type
		c.Options.Image.Features = p.Features
	}

	if c.Options.MaxQuality {
		c.Options.Image.Format = "png"
		c.Options.Image.GrayScale = false
		c.Options.Image.Resize = false
	} else if c.Options.BestQuality {
		c.Options.Image.Format = "jpeg"
		c.Options.Image.Quality = 100
		c.Options.Image.GrayScale = false
		c.Options.Image.Resize = false
	} else if c.Options.GreatQuality {
		c.Options.Image.Format = "jpeg"
		c.Options.Image.Quality = 90
		c.Options.Image.GrayScale = true
		c.Options.Image.Resize = false
	} else if c.Options.GoodQuality {
		c.Options.Image.Format = "jpeg"
		c.Options.Image.Quality = 90
		c.Options.Image.GrayScale = true
		c.Options.Image.Resize = true
	}

	if c.Options.NoFilter {
		c.Options.Image.Crop.Enabled = false
		c.Options.Image.Brightness = 0
		c.Options.Image.Contrast = 0
		c.Options.Image.AutoContrast = false
		c.Options.Image.NoBlankImage = false
		c.Options.Image.Resize = false
	}

	if c.Options.Image.Device == epuboptions.Apple {
		c.Options.Image.DoublePageMode = epuboptions.DoublePageSplit
		c.Options.Image.KeepSplitDoublePageAspect = true
	}
}

// Validate Check parameters
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
		// "." and ".." have no name, use the real one
		if b := filepath.Base(inputBase); b == "." || b == ".." {
			if inputBase, err = filepath.Abs(inputBase); err != nil {
				return err
			}
		}
		defaultOutput = inputBase + ".epub"
	} else {
		ext := filepath.Ext(inputBase)
		defaultOutput = inputBase[0:len(inputBase)-len(ext)] + ".epub"
	}

	if c.Options.Output == "" {
		c.Options.Output = defaultOutput
	}

	c.Options.Output = filepath.Clean(c.Options.Output)
	if strings.EqualFold(filepath.Ext(c.Options.Output), ".epub") {
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
		return ErrProfileMissing
	}

	p := c.Options.GetProfile()
	if p == nil {
		return &UnknownProfileError{Profile: c.Options.Profile}
	}

	// Grayscale
	if !p.Features.Color && !c.Options.Image.GrayScale {
		return fmt.Errorf("profile %q doesn't support color, grayscale can't be turned off", c.Options.Profile)
	}

	// Mode
	if c.Options.Image.Mode < epuboptions.ModeComic || c.Options.Image.Mode > epuboptions.ModeWebtoon {
		return errors.New("mode should be 0, 1 or 2")
	}
	if c.Options.Image.Mode == epuboptions.ModeManga && !p.Features.Manga {
		return fmt.Errorf("profile %q doesn't support the manga mode", c.Options.Profile)
	}

	// Landscape
	if c.Options.Image.Landscape && !p.Features.Landscape {
		return fmt.Errorf("profile %q doesn't support the landscape only mode", c.Options.Profile)
	}
	if c.Options.Image.Landscape && c.Options.Image.DoublePageMode != epuboptions.DoublePageSplit {
		return errors.New("the landscape only mode doesn't support the rotation of the double pages: use doublepagemode=0")
	}

	// LimitMb
	if c.Options.LimitMb < 20 && c.Options.LimitMb != 0 {
		return errors.New("limitmb should be 0 or >= 20")
	}

	// Quality
	if c.Options.Image.Quality < 1 || c.Options.Image.Quality > 100 {
		return errors.New("quality should be between 1 and 100")
	}

	// JXL effort
	if c.Options.Image.JXLEffort < 1 || c.Options.Image.JXLEffort > jxl.MaxEffort {
		return fmt.Errorf("jxl effort should be between 1 and %d", jxl.MaxEffort)
	}

	// Workers
	if c.Options.Workers < 1 {
		return errors.New("workers should be at least 1")
	}

	// Brightness
	if c.Options.Image.Brightness < -100 || c.Options.Image.Brightness > 100 {
		return errors.New("brightness should be between -100 and 100")
	}

	// Contrast
	if c.Options.Image.Contrast < -100 || c.Options.Image.Contrast > 100 {
		return errors.New("contrast should be between -100 and 100")
	}

	// SortPathMode
	if c.Options.SortPathMode < 0 || c.Options.SortPathMode > 2 {
		return errors.New("sort should be 0, 1 or 2")
	}

	// Color
	colorRegex := regexp.MustCompile("^([0-9a-fA-F]{3}){1,2}$")
	if !colorRegex.MatchString(c.Options.Image.View.Color.Foreground) {
		return errors.New("foreground color must have color format in hexadecimal on 3 or 6 characters: example: 000, FFF, 000000, FFFFFF, A000FF")
	}

	if !colorRegex.MatchString(c.Options.Image.View.Color.Background) {
		return errors.New("background color must have color format in hexadecimal on 3 or 6 characters: example: 000, FFF, 000000, FFFFFF, A000FF")
	}

	// Format
	if !slices.Contains([]string{"jpeg", "png", "jxl", "copy"}, c.Options.Image.Format) {
		return errors.New("format should be jpeg, png, jxl or copy")
	}

	// the webtoon is cut into pages: the images must be processed
	if c.Options.Image.WebtoonMode() && c.Options.Image.Format == "copy" {
		return errors.New("the webtoon mode doesn't support the copy format: use jpeg, png or jxl")
	}

	// the source of a webtoon is a long strip, its aspect ratio is not the one of a page
	if c.Options.Image.WebtoonMode() && c.Options.Image.View.AspectRatio == 0 {
		return fmt.Errorf("the webtoon mode doesn't support the aspect ratio of the source (0): use -1 for the device, or a ratio between %v and %v",
			epuboptions.MinAspectRatio, epuboptions.MaxAspectRatio)
	}

	// Aspect Ratio
	if ar := c.Options.Image.View.AspectRatio; ar != -1 && ar != 0 &&
		(ar < epuboptions.MinAspectRatio || ar > epuboptions.MaxAspectRatio) {
		return fmt.Errorf(
			"aspect ratio should be -1, 0 or between %v and %v",
			epuboptions.MinAspectRatio, epuboptions.MaxAspectRatio,
		)
	}

	// Zoom
	if z := c.Options.Image.Zoom; z != 0 && (z < epuboptions.MinZoom || z > epuboptions.MaxZoom) {
		return fmt.Errorf(
			"zoom should be 0 or between %v and %v",
			epuboptions.MinZoom, epuboptions.MaxZoom,
		)
	}

	// Title Page
	if c.Options.TitlePage < 0 || c.Options.TitlePage > 2 {
		return errors.New("title page should be 0, 1 or 2")
	}

	// Double Page Mode
	if c.Options.Image.DoublePageMode < 0 || c.Options.Image.DoublePageMode > 2 {
		return errors.New("double page mode should be 0, 1 or 2")
	}

	// Grayscale Mode
	if c.Options.Image.GrayScaleMode < 0 || c.Options.Image.GrayScaleMode > 2 {
		return errors.New("grayscale mode should be 0, 1 or 2")
	}

	// crop
	if c.Options.Image.Crop.Limit < 0 || c.Options.Image.Crop.Limit > 100 {
		return errors.New("crop limit should be between 0 and 100")
	}

	return nil
}

// ErrProfileMissing no profile selected.
var ErrProfileMissing = errors.New("profile missing")

// UnknownProfileError the selected profile doesn't exist.
type UnknownProfileError struct {
	Profile string
}

func (e *UnknownProfileError) Error() string {
	return fmt.Sprintf("profile %q doesn't exists", e.Profile)
}

// Fatal Helper to show usage, err and exit 1.
// A wrong profile only shows the available profiles instead of the whole usage.
func (c *Converter) Fatal(err error) {
	var unknown *UnknownProfileError
	if errors.Is(err, ErrProfileMissing) || errors.As(err, &unknown) {
		utils.Fatalf("Error: %s\n\nAvailable profiles:\n%s\n", err, c.Options.AvailableProfiles())
	}
	c.Cmd.Usage()
	utils.Fatalf("\nError: %s\n", err)
}

func (c *Converter) Stats() {
	// Display elapse time and memory usage
	elapse := time.Since(c.startAt).Round(time.Millisecond)
	var mem runtime.MemStats
	runtime.ReadMemStats(&mem)

	if c.Options.Json {
		_ = json.NewEncoder(os.Stdout).Encode(map[string]any{
			"type": "stats",
			"data": map[string]any{
				"elapse_ms":       elapse.Milliseconds(),
				"memory_usage_mb": mem.Sys / 1024 / 1024,
			},
		})
	} else {
		utils.Printf(
			"Completed in %s, Memory usage %d Mb\n",
			elapse,
			mem.Sys/1024/1024,
		)
	}
}
