package sdb

import (
	"fmt"
	"io"
)

// PrintTars prints the names of the active TAR files in 'd' to 'w''. If the
// 'all' parameter is 'true', it prints the names of both active and non-active
// TAR files in 'directory' instead.
func PrintTars(d string, all bool, w io.Writer) error {
	return forEachTarFile(d, all, func(n string) {
		fmt.Fprintln(w, n)
	})
}

// PrintBinaries prints the content of the binary references index from the TAR
// file at 'p' to 'w'.
func PrintBinaries(p string, f Format, w io.Writer) error {
	return onMatchingEntry(p, isBinary, printBinaries(f, w))
}

// PrintGraph prints the content of the graph from the TAR file at 'p' to 'w'.
func PrintGraph(p string, f Format, w io.Writer) error {
	return onMatchingEntry(p, isGraph, printGraph(f, w))
}

// PrintIndex prints the content of the index from the TAR file at 'p' to 'w'
func PrintIndex(p string, f Format, w io.Writer) error {
	return onMatchingEntry(p, isIndex, printIndex(f, w))
}

// PrintSegments lists the identifiers of every segment contained in the TAR
// file at 'p' and prints them to 'w'. The segment IDs are printed in the same
// order as the correspodning entries appear in the TAR files.
func PrintSegments(p string, w io.Writer) error {
	return forEachMatchingEntry(p, isAnySegment, printSegmentNameTo(w))
}

// PrintSegment prints the content of the segment 'id' from the TAR file at
// 'p' to 'w'.
func PrintSegment(p string, id string, f Format, w io.Writer) error {
	return onMatchingEntry(p, isSegment(id), printSegment(f, w))
}

// PrintEntries prints the name of the entries from the TAR file at 'p' to 'w'.
// The entries are printed in the same order as they are stored in the TAR file.
func PrintEntries(p string, w io.Writer) error {
	return forEachEntry(p, printNameTo(w))
}
