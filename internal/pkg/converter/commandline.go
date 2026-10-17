package converter

import (
	"flag"
	"regexp"
	"strconv"
	"strings"

	"github.com/celogeek/go-comic-converter/v3/internal/pkg/utils"
)

// Sections of the parameters that are part of the command line.
const (
	sectionOutput = "Output"
	sectionConfig = "Config"
	// sectionProfiles is only the table of the profiles at the end of the help, it has no parameter.
	sectionProfiles = "Profiles"
)

// commandName name of the program in the command line.
const commandName = "go-comic-converter"

// CommandLineArgs the arguments that give the current options, whatever the saved config:
// -ignore-config, the profile, and the parameters that differ from the built-in default
// (the default of the profile for the zoom).
//
// Each argument is -name=value, without any quoting.
func (c *Converter) CommandLineArgs() []string {
	// the default of the parameters, without the saved config: it is what -ignore-config gives
	builtin := New()
	builtin.InitParse()

	args := []string{"-ignore-config"}
	section := ""
	for _, o := range c.order {
		switch v := o.(type) {
		case orderSection:
			section = v.Value()
		case orderName:
			f := c.Cmd.Lookup(v.Value())
			if f == nil || (section != sectionOutput && section != sectionConfig) {
				continue
			}
			value := f.Value.String()
			if f.Name != "profile" && c.isDefault(builtin.Cmd.Lookup(f.Name), f, value) {
				continue
			}
			args = append(args, c.commandLineArg(f, value))
		}
	}
	return args
}

// isDefault tells if the value of the parameter is the one given by -ignore-config,
// where def is the parameter with its built-in default.
func (c *Converter) isDefault(def, f *flag.Flag, value string) bool {
	if f.Name == "zoom" {
		// the default is the one of the profile, an explicit zoom as the profile is useless
		return c.Options.Image.ZoomFactor() == max(c.Options.Image.Features.Zoom, 1)
	}
	return def != nil && value == def.DefValue
}

// commandLineArg the argument of a parameter: booleans are always -name=1 or -name=0.
func (c *Converter) commandLineArg(f *flag.Flag, value string) string {
	if f.Name == "zoom" {
		// the zoom of the profile is made explicit
		value = utils.FloatToString(c.Options.Image.ZoomFactor(), -1)
	}
	if b, ok := f.Value.(interface{ IsBoolFlag() bool }); ok && b.IsBoolFlag() {
		value = strconv.Itoa(utils.BoolStringToInt(value))
	}
	return "-" + f.Name + "=" + value
}

// CommandLine a command line, ready to be copied, equivalent to the current options.
//
// It ignores the saved config, and gives the profile and the options that differ from the default:
// it does the same whatever the config of the one who runs it.
func (c *Converter) CommandLine() string {
	args := c.CommandLineArgs()
	lines := make([]string, 0, len(args)+1)
	lines = append(lines, commandName+" "+shellQuoteArg(args[0]))
	for _, a := range args[1:] {
		lines = append(lines, "  "+shellQuoteArg(a))
	}
	return strings.Join(lines, " \\\n")
}

var shellSafe = regexp.MustCompile(`^[A-Za-z0-9_@%+=:,./-]+$`)

// shellQuoteArg quote the value of a -name=value argument for a shell.
func shellQuoteArg(arg string) string {
	name, value, ok := strings.Cut(arg, "=")
	if !ok || shellSafe.MatchString(value) {
		return arg
	}
	return name + "=" + "'" + strings.ReplaceAll(value, "'", `'\''`) + "'"
}
