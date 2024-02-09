/*
Package sortpath support sorting of path that may include number.

A series of path can look like:
  - Tome1/Chap1/Image1.jpg
  - Tome1/Chap2/Image1.jpg
  - Tome1/Chap10/Image2.jpg

The module will split the string by path,
and compare them by decomposing the string and number part.

The module support 3 mode:
  - mode=0 alpha for path and file
  - mode=1 alphanumeric for path and alpha for file
  - mode=2 alphanumeric for path and file

Example:

	files := []string{
		'T1/C1/Img1.jpg',
		'T1/C2/Img1.jpg',
		'T1/C10/Img1.jpg',
	}

	sort.Sort(sortpath.By(files, 1))
*/
package sortpath

import "sort"

// struct that implement interface for sort.Sort
type by struct {
	filenames []string
	paths     [][]part
}

func (b by) Len() int           { return len(b.filenames) }
func (b by) Less(i, j int) bool { return compareParts(b.paths[i], b.paths[j]) < 0 }
func (b by) Swap(i, j int) {
	b.filenames[i], b.filenames[j] = b.filenames[j], b.filenames[i]
	b.paths[i], b.paths[j] = b.paths[j], b.paths[i]
}

// By use sortpath.By with sort.Sort
func By(filenames []string, mode int) sort.Interface {
	var p [][]part
	for _, filename := range filenames {
		p = append(p, parse(filename, mode))
	}
	return by{filenames, p}
}
