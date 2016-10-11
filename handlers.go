package sdb

import (
	"encoding/hex"
	"fmt"
	"io"
	"strings"

	"github.com/francescomari/sdb/binaries"
	"github.com/francescomari/sdb/graph"
	"github.com/francescomari/sdb/index"
	"github.com/francescomari/sdb/segment"
)

func invalidFormat() handler {
	return func(_ string, _ io.Reader) error {
		return ErrInvalidFormat
	}
}

func printBinaries(f Format, w io.Writer) handler {
	switch f {
	case FormatHex:
		return printHexTo(w)
	case FormatText:
		return printBinariesTo(w)
	default:
		return invalidFormat()
	}
}

func printBinariesTo(w io.Writer) handler {
	return func(_ string, r io.Reader) error {
		var bns binaries.Binaries

		if _, err := bns.ReadFrom(r); err != nil {
			return err
		}

		for _, generation := range bns.Generations {
			fmt.Fprintf(w, "%d\n", generation.Generation)

			for _, segment := range generation.Segments {
				fmt.Fprintf(w, "    %016x%016x\n", segment.Msb, segment.Lsb)

				for _, reference := range segment.References {
					fmt.Fprintf(w, "        %s\n", reference)
				}
			}
		}

		return nil
	}
}

func printGraph(f Format, w io.Writer) handler {
	switch f {
	case FormatHex:
		return printHexTo(w)
	case FormatText:
		return printGraphTo(w)
	default:
		return invalidFormat()
	}
}

func printGraphTo(w io.Writer) handler {
	return func(_ string, r io.Reader) error {
		var gph graph.Graph

		if _, err := gph.ReadFrom(r); err != nil {
			return nil
		}

		for _, entry := range gph.Entries {
			fmt.Fprintf(w, "%016x%016x\n", entry.Msb, entry.Lsb)

			for _, reference := range entry.References {
				fmt.Fprintf(w, "    %016x%016x\n", reference.Msb, reference.Lsb)
			}
		}

		return nil
	}
}

func printIndex(f Format, w io.Writer) handler {
	switch f {
	case FormatHex:
		return printHexTo(w)
	case FormatText:
		return printIndexTo(w)
	default:
		return invalidFormat()
	}
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

func printSegment(f Format, w io.Writer) handler {
	switch f {
	case FormatHex:
		return printHexTo(w)
	case FormatText:
		return printSegmentTo(w)
	default:
		return invalidFormat()
	}
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

func printNameTo(w io.Writer) handler {
	return func(n string, _ io.Reader) error {
		fmt.Fprintln(w, n)
		return nil
	}
}

func printHexTo(w io.Writer) handler {
	return func(_ string, r io.Reader) (err error) {
		d := hex.Dumper(w)
		defer d.Close()
		_, err = io.Copy(d, r)
		return
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
