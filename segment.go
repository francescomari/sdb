package sdb

import (
	"fmt"
	"io"
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
// file at 'p' and prints them to 'w'. The segment IDs are printed in the same
// order as the correspodning entries appear in the TAR files.
func PrintSegments(p string, w io.Writer) error {
	return forEachMatchingEntry(p, segmentEntryRegexp.MatchString, printSegmentNameTo(w))
}

func printSegmentNameTo(w io.Writer) handler {
	return func(n string, _ io.Reader) error {
		id := normalizeSegmentID(entryNameToSegmentID(n))

		kind := "data"

		if isBulkSegmentID(id) {
			kind = "bulk"
		}

		fmt.Fprintf(w, "%s %s\n", kind, id)

		return nil
	}
}

// PrintSegment prints the content of the segment 'id' from the TAR file at
// 'path' to 'writer'.
func PrintSegment(path string, id string, writer io.Writer) error {
	return onMatchingEntry(path, SegmentEntryFilter(id), printSegmentTo(writer))
}

func printSegmentTo(w io.Writer) handler {
	return func(_ string, r io.Reader) error {
		var s segment.Segment

		if _, err := s.ReadFrom(r); err != nil {
			return err
		}

		fmt.Fprintf(w, "Version    %d\n", s.Version)
		fmt.Fprintf(w, "Generation %d\n", s.Generation)

		fmt.Fprintf(w, "References\n")

		for i, r := range s.References {
			fmt.Fprintf(w, "    %4d %016x%016x\n", i+1, r.Msb, r.Lsb)
		}

		fmt.Fprintf(w, "Records\n")

		for _, r := range s.Records {
			fmt.Fprintf(w, "    %08x %-10s %08x\n", r.Number, recordTypeString(r.Type), r.Offset)
		}

		return nil
	}
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
