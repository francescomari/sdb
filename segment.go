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

// PrintSegment prints the content of the segment with identifier 'segmentID'
// from the TAR file at 'path' to 'writer'. The desired output format is
// specified by 'output'.
func PrintSegment(path string, segmentID string, writer io.Writer, output OutputType) error {
	file, err := os.Open(path)

	if err != nil {
		return err
	}

	defer file.Close()

	requested := normalizeSegmentID(segmentID)

	reader := tar.NewReader(file)

	for {
		header, err := reader.Next()

		if header == nil {
			break
		}

		if err != nil {
			return err
		}

		if !segmentEntryRegexp.MatchString(header.Name) {
			continue
		}

		current := normalizeSegmentID(entryNameToSegmentID(header.Name))

		if current != requested {
			continue
		}

		if output == OutputHex {
			return printHex(reader, writer)
		}

		return fmt.Errorf("Invalid output type")
	}

	return nil
}

func isBulkSegmentID(id string) bool {
	return id[16] == 'b'
}

func normalizeSegmentID(id string) string {
	return strings.ToLower(strings.TrimSpace(strings.Replace(id, "-", "", -1)))
}
