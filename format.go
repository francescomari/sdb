package sdb

// Format represents an output format that can be used to represent data.
type Format int

const (
	// FormatHex can be used to print a hex dump of a piece of data.
	FormatHex Format = iota
	// FormatText can be used to print a plain-text, human-readable version of a
	// piece of data.
	FormatText
)
