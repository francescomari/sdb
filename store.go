package sdb

import (
	"fmt"
	"io"
	"io/ioutil"
	"os"
	"regexp"
	"sort"
	"strconv"
)

var tarFileNameRegexp = regexp.MustCompile("^data([0-9]{5})([a-z]).tar$")

// PrintTars prints the names of the active TAR files in 'directory to
// 'writer''. If the 'all' parameter is 'true', it prints the names of both
// active and non-active TAR files in 'directory' instead.
func PrintTars(directory string, all bool, writer io.Writer) error {
	infos, err := ioutil.ReadDir(directory)

	if err != nil {
		return fmt.Errorf("Unable to read directory '%s': %s", directory, err)
	}

	files := fileInfos(infos).regulars().names().tars()

	if !all {
		files = files.youngest()
	}

	for _, name := range files.sorted().names() {
		fmt.Fprintln(writer, name)
	}

	return nil
}

type fileInfos []os.FileInfo

func (infos fileInfos) regulars() fileInfos {
	var regulars fileInfos

	for _, info := range infos {
		if info.Mode().IsRegular() {
			regulars = append(regulars, info)
		}
	}

	return regulars
}

func (infos fileInfos) names() fileNames {
	var names fileNames

	for _, info := range infos {
		names = append(names, info.Name())
	}

	return names
}

type fileNames []string

func (names fileNames) tars() tarFileNames {
	var tars tarFileNames

	for _, name := range names {
		matches := tarFileNameRegexp.FindStringSubmatch(name)

		if matches == nil {
			continue
		}

		number, err := strconv.ParseUint(matches[1], 10, 64)

		if err != nil {
			panic("Invalid TAR file name detection")
		}

		tars = append(tars, tarFileName{name, number, matches[2][0]})
	}

	return tars
}

type tarFileName struct {
	name       string
	number     uint64
	generation uint8
}

type tarFileNames []tarFileName

func (tars tarFileNames) Len() int {
	return len(tars)
}

func (tars tarFileNames) Less(i, j int) bool {
	return tars[i].number < tars[j].number || tars[i].number == tars[j].number && tars[i].generation < tars[j].generation
}

func (tars tarFileNames) Swap(i, j int) {
	tars[i], tars[j] = tars[j], tars[i]
}

func (tars tarFileNames) names() fileNames {
	var names fileNames

	for _, tar := range tars {
		names = append(names, tar.name)
	}

	return names
}

func (tars tarFileNames) sorted() tarFileNames {
	var result tarFileNames

	for _, tar := range tars {
		result = append(result, tar)
	}

	sort.Sort(result)

	return result
}

func (tars tarFileNames) youngest() tarFileNames {
	generations := make(map[uint64]uint8)

	for _, tar := range tars {
		if generation, ok := generations[tar.number]; !ok || generation < tar.generation {
			generations[tar.number] = tar.generation
		}
	}

	var youngest []tarFileName

	for _, tar := range tars {
		if generations[tar.number] == tar.generation {
			youngest = append(youngest, tar)
		}
	}

	return youngest
}
