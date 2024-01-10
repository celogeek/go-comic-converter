package converter

import (
	"flag"
	"fmt"
	"reflect"
)

// Create a new section of config
func (c *converter) addSection(section string) {
	c.order = append(c.order, converterOrderSection{value: section})
}

// Add a string parameter
func (c *converter) addStringParam(p *string, name string, value string, usage string) {
	c.Cmd.StringVar(p, name, value, usage)
	c.order = append(c.order, converterOrderName{value: name, isString: true})
}

// Add an integer parameter
func (c *converter) addIntParam(p *int, name string, value int, usage string) {
	c.Cmd.IntVar(p, name, value, usage)
	c.order = append(c.order, converterOrderName{value: name})
}

// Add an float parameter
func (c *converter) addFloatParam(p *float64, name string, value float64, usage string) {
	c.Cmd.Float64Var(p, name, value, usage)
	c.order = append(c.order, converterOrderName{value: name})
}

// Add a boolean parameter
func (c *converter) addBoolParam(p *bool, name string, value bool, usage string) {
	c.Cmd.BoolVar(p, name, value, usage)
	c.order = append(c.order, converterOrderName{value: name})
}

// Taken from flag package as it is private and needed for usage.
//
// isZeroValue determines whether the string represents the zero
// value for a flag.
func (c *converter) isZeroValue(f *flag.Flag, value string) (ok bool, err error) {
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
