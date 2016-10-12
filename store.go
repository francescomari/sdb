package sdb

import (
	"fmt"
	"io/ioutil"
	"regexp"
	"sort"
	"strconv"
)

var tarFileRegexp = regexp.MustCompile("^data([0-9]{5})([a-z]).tar$")

func forEachTarFile(directory string, all bool, f func(name string)) error {
	infos, err := ioutil.ReadDir(directory)

	if err != nil {
		return fmt.Errorf("Unable to read directory '%s': %s", directory, err)
	}

	var tars []tarFile

	for _, info := range infos {
		if info.Mode().IsRegular() == false {
			continue
		}

		matches := tarFileRegexp.FindStringSubmatch(info.Name())

		if matches == nil {
			continue
		}

		number, err := strconv.ParseUint(matches[1], 10, 64)

		if err != nil {
			panic("Invalid TAR file name detection")
		}

		tars = append(tars, tarFile{info.Name(), number, matches[2][0]})
	}

	generations := make(map[uint64]uint8)

	for _, tar := range tars {
		if generation, ok := generations[tar.number]; !ok || generation < tar.generation {
			generations[tar.number] = tar.generation
		}
	}

	var youngest []tarFile

	for _, tar := range tars {
		if all || generations[tar.number] == tar.generation {
			youngest = append(youngest, tar)
		}
	}

	sort.Sort(tarFiles(youngest))

	for _, tar := range youngest {
		f(tar.name)
	}

	return nil
}

type tarFile struct {
	name       string
	number     uint64
	generation uint8
}

type tarFiles []tarFile

func (tars tarFiles) Len() int {
	return len(tars)
}

func (tars tarFiles) Less(i, j int) bool {
	return tars[i].number < tars[j].number || tars[i].number == tars[j].number && tars[i].generation < tars[j].generation
}

func (tars tarFiles) Swap(i, j int) {
	tars[i], tars[j] = tars[j], tars[i]
}
