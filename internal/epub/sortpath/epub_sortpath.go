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

func comparePart(a, b []part) float64 {
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

type by struct {
	filenames []string
	paths     [][]part
}

func (b by) Len() int           { return len(b.filenames) }
func (b by) Less(i, j int) bool { return comparePart(b.paths[i], b.paths[j]) < 0 }
func (b by) Swap(i, j int) {
	b.filenames[i], b.filenames[j] = b.filenames[j], b.filenames[i]
	b.paths[i], b.paths[j] = b.paths[j], b.paths[i]
}

func By(filenames []string, mode int) by {
	p := [][]part{}
	for _, filename := range filenames {
		p = append(p, parse(filename, mode))
	}
	return by{filenames, p}
}
