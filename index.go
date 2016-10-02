package sdb

import (
	"archive/tar"
	"fmt"
	"io"
	"os"
	"strings"

	"github.com/francescomari/sdb/index"
)

// IndexEntryFilter locates the index in a TAR file.
func IndexEntryFilter(name string) bool {
	return strings.HasSuffix(name, ".idx")
}

// PrintIndex prints the content of the index from the TAR file at 'path' to
// 'writer'
func PrintIndex(path string, writer io.Writer) error {
	file, err := os.Open(path)

	if err != nil {
		return err
	}

	defer file.Close()

	reader := tar.NewReader(file)

	for {
		header, err := reader.Next()

		if header == nil {
			break
		}

		if err != nil {
			return err
		}

		if IndexEntryFilter(header.Name) {
			return printIndexText(reader, writer)
		}
	}

	return nil
}

func printIndexText(reader io.Reader, writer io.Writer) error {
	var idx index.Index

	if _, err := idx.ReadFrom(reader); err != nil {
		return err
	}

	for _, e := range idx.Entries {
		id := fmt.Sprintf("%016x%016x", e.Msb, e.Lsb)

		kind := "data"

		if isBulkSegmentID(id) {
			kind = "bulk"
		}

		fmt.Fprintf(writer, "%s %s %8x %6d %6d\n", kind, id, e.Position, e.Size, e.Generation)
	}

	return nil
}
