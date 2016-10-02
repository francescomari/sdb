package sdb

import (
	"archive/tar"
	"fmt"
	"io"
	"os"
	"regexp"
	"strings"
)

var segmentEntryRegexp = regexp.MustCompile("^[0-9a-f]{8}-[0-9a-f]{4}-[0-9a-f]{4}-[0-9a-f]{4}-[0-9a-f]{12}\\.[0-9a-f]{8}$")

// SegmentEntryFilter builds a filter to locate a specific segment in a TAR file.
func SegmentEntryFilter(id string) func(string) bool {
	sid := normalizeSegmentID(id)
	return func(name string) bool {
		return sid == normalizeSegmentID(entryNameToSegmentID(name))
	}
}

// PrintSegments lists the identifiers of every segment contained in the TAR
// file at 'path' and prints them to 'writer'. The segment IDs are printed in
// the same order as the correspodning entries appear in the TAR files.
func PrintSegments(path string, writer io.Writer) error {
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

		if segmentEntryRegexp.MatchString(header.Name) {
			id := normalizeSegmentID(entryNameToSegmentID(header.Name))

			kind := "data"

			if isBulkSegmentID(id) {
				kind = "bulk"
			}

			fmt.Fprintf(writer, "%s %s\n", kind, id)
		}
	}

	return nil
}

func isBulkSegmentID(id string) bool {
	return id[16] == 'b'
}

func normalizeSegmentID(id string) string {
	return strings.ToLower(strings.TrimSpace(strings.Replace(id, "-", "", -1)))
}
