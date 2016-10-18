package main

import (
	"fmt"
	"io/ioutil"
	"os"
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
	tars := readTarFiles(infos)
	if !all {
		generations := youngestGenerationByNumber(tars)
		tars = youngestTarFiles(tars, generations)
	}
	sort.Sort(tars)
	for _, tar := range tars {
		f(tar.name)
	}
	return nil
}

func readTarFiles(infos []os.FileInfo) tarFiles {
	var tars tarFiles
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
	return tars
}

func youngestGenerationByNumber(tars tarFiles) map[uint64]uint8 {
	generations := make(map[uint64]uint8)
	for _, tar := range tars {
		if generation, ok := generations[tar.number]; !ok || generation < tar.generation {
			generations[tar.number] = tar.generation
		}
	}
	return generations
}

func youngestTarFiles(tars tarFiles, generations map[uint64]uint8) tarFiles {
	var youngest tarFiles
	for _, tar := range tars {
		if generations[tar.number] == tar.generation {
			youngest = append(youngest, tar)
		}
	}
	return youngest
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
