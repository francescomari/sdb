package main

import (
	"regexp"
	"strings"
)

var segmentEntryRegexp = regexp.MustCompile("^[0-9a-f]{8}-[0-9a-f]{4}-[0-9a-f]{4}-[0-9a-f]{4}-[0-9a-f]{12}\\.[0-9a-f]{8}$")

func any(_ string) bool {
	return true
}

func isBinary(name string) bool {
	return strings.HasSuffix(name, ".brf")
}

func isGraph(name string) bool {
	return strings.HasSuffix(name, ".gph")
}

func isIndex(name string) bool {
	return strings.HasSuffix(name, ".idx")
}

func isAnySegment(n string) bool {
	return segmentEntryRegexp.MatchString(n)
}

func isSegment(id string) matcher {
	return func(name string) bool {
		return normalizeSegmentID(id) == normalizeSegmentID(entryNameToSegmentID(name))
	}
}

func entryNameToSegmentID(header string) string {
	return header[:strings.Index(header, ".")]
}
