package sortpath

import (
	"path/filepath"
	"regexp"
	"strconv"
	"strings"
)

// Strings follow with numbers like: s1, s1.2, s2-3, ...
var split_path_regex = regexp.MustCompile(`^(.*?)(\d+(?:\.\d+)?)(?:-(\d+(?:\.\d+)?))?$`)

type part struct {
	fullname string
	name     string
	number   float64
}

// compare part, first check if both include a number,
// if so, compare string part then number part, else compare there as string.
func (a part) compare(b part) float64 {
	if a.number == 0 || b.number == 0 {
		return float64(strings.Compare(a.fullname, b.fullname))
	}
	if a.name == b.name {
		return a.number - b.number
	} else {
		return float64(strings.Compare(a.name, b.name))
	}
}

// separate from the string the number part.
func parsePart(p string) part {
	r := split_path_regex.FindStringSubmatch(p)
	if len(r) == 0 {
		return part{p, p, 0}
	}
	n, err := strconv.ParseFloat(r[2], 64)
	if err != nil {
		return part{p, p, 0}
	}
	return part{p, r[1], n}
}

// mode=0 alpha for path and file
// mode=1 alphanum for path and alpha for file
// mode=2 alphanum for path and file
func parse(filename string, mode int) []part {
	pathname, name := filepath.Split(strings.ToLower(filename))
	pathname = strings.TrimSuffix(pathname, string(filepath.Separator))
	ext := filepath.Ext(name)
	name = name[0 : len(name)-len(ext)]

	f := []part{}
	for _, p := range strings.Split(pathname, string(filepath.Separator)) {
		if mode > 0 { // alphanum for path
			f = append(f, parsePart(p))
		} else {
			f = append(f, part{p, p, 0})
		}
	}
	if mode == 2 { // alphanum for file
		f = append(f, parsePart(name))
	} else {
		f = append(f, part{name, name, 0})
	}
	return f
}

// compare 2 fullpath splitted into parts
func compareParts(a, b []part) float64 {
	m := len(a)
	if m > len(b) {
		m = len(b)
	}
	for i := 0; i < m; i++ {
		c := a[i].compare(b[i])
		if c != 0 {
			return c
		}
	}
	return float64(len(a) - len(b))
}
