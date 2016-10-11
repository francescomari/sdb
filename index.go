package sdb

import (
	"fmt"
	"io"
	"strings"

	"github.com/francescomari/sdb/index"
)

// IndexEntryFilter locates the index in a TAR file.
func IndexEntryFilter(name string) bool {
	return strings.HasSuffix(name, ".idx")
}

// PrintIndex prints the content of the index from the TAR file at 'p' to 'w'
func PrintIndex(p string, w io.Writer) error {
	return onMatchingEntry(p, IndexEntryFilter, printIndexTo(w))
}

func printIndexTo(w io.Writer) handler {
	return func(_ string, r io.Reader) error {
		var idx index.Index

		if _, err := idx.ReadFrom(r); err != nil {
			return err
		}

		for _, e := range idx.Entries {
			id := fmt.Sprintf("%016x%016x", e.Msb, e.Lsb)

			kind := "data"

			if isBulkSegmentID(id) {
				kind = "bulk"
			}

			fmt.Fprintf(w, "%s %s %8x %6d %6d\n", kind, id, e.Position, e.Size, e.Generation)
		}

		return nil
	}
}
