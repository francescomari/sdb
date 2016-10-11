package sdb

import (
	"archive/tar"
	"fmt"
	"io"
	"os"
	"regexp"
	"strings"

	"github.com/francescomari/sdb/segment"
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

// PrintSegment prints the content of the segment 'id' from the TAR file at
// 'path' to 'writer'.
func PrintSegment(path string, id string, writer io.Writer) error {
	id = normalizeSegmentID(id)

	if isBulkSegmentID(id) {
		return fmt.Errorf("The ID refers to a bulk segment")
	}

	file, err := os.Open(path)

	if err != nil {
		return err
	}

	defer file.Close()

	reader := tar.NewReader(file)
	filter := SegmentEntryFilter(id)

	for {
		header, err := reader.Next()

		if header == nil {
			break
		}

		if err != nil {
			return err
		}

		if filter(header.Name) {
			return printSegmentText(reader, writer)
		}
	}

	return nil
}

func printSegmentText(reader io.Reader, writer io.Writer) error {
	var s segment.Segment

	if _, err := s.ReadFrom(reader); err != nil {
		return err
	}

	fmt.Fprintf(writer, "Version    %d\n", s.Version)
	fmt.Fprintf(writer, "Generation %d\n", s.Generation)

	fmt.Fprintf(writer, "References\n")

	for i, r := range s.References {
		fmt.Fprintf(writer, "    %4d %016x%016x\n", i+1, r.Msb, r.Lsb)
	}

	fmt.Fprintf(writer, "Records\n")

	for _, r := range s.Records {
		fmt.Fprintf(writer, "    %08x %-10s %08x\n", r.Number, recordTypeString(r.Type), r.Offset)
	}

	return nil
}

func isBulkSegmentID(id string) bool {
	return id[16] == 'b'
}

func normalizeSegmentID(id string) string {
	return strings.ToLower(strings.TrimSpace(strings.Replace(id, "-", "", -1)))
}

func recordTypeString(t segment.RecordType) string {
	switch t {
	case segment.RecordTypeBlock:
		return "block"
	case segment.RecordTypeList:
		return "list"
	case segment.RecordTypeListBucket:
		return "bucket"
	case segment.RecordTypeMapBranch:
		return "branch"
	case segment.RecordTypeMapLeaf:
		return "leaf"
	case segment.RecordTypeNode:
		return "node"
	case segment.RecordTypeTemplate:
		return "template"
	case segment.RecordTypeValue:
		return "value"
	default:
		return "unknown"
	}
}
