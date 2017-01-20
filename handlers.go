package main

import (
	"encoding/hex"
	"errors"
	"fmt"
	"io"
	"strings"

	"github.com/francescomari/sdb/binaries"
	"github.com/francescomari/sdb/graph"
	"github.com/francescomari/sdb/index"
	"github.com/francescomari/sdb/segment"
)

var errInvalidFormat = errors.New("Invalid format")

func invalidFormat() handler {
	return func(_ string, _ io.Reader) error {
		return errInvalidFormat
	}
}

func doPrintTo(w io.Writer) func(n string) {
	return func(n string) {
		fmt.Fprintf(w, n)
	}
}

func doPrintBinaries(f format, w io.Writer) handler {
	switch f {
	case formatHex:
		return doPrintHexTo(w)
	case formatText:
		return doPrintBinariesTo(w)
	default:
		return invalidFormat()
	}
}

func doPrintBinariesTo(w io.Writer) handler {
	return func(_ string, r io.Reader) error {
		var bns binaries.Binaries
		if _, err := bns.ReadFrom(r); err != nil {
			return err
		}
		for _, g := range bns.Generations {
			for _, s := range g.Segments {
				for _, r := range s.References {
					fmt.Fprintf(w, "%d %s %s\n", g.Generation, segmentID(s.Msb, s.Lsb), r)
				}
			}
		}
		return nil
	}
}

func doPrintGraph(f format, w io.Writer) handler {
	switch f {
	case formatHex:
		return doPrintHexTo(w)
	case formatText:
		return doPrintGraphTo(w)
	default:
		return invalidFormat()
	}
}

func doPrintGraphTo(w io.Writer) handler {
	return func(_ string, r io.Reader) error {
		var gph graph.Graph
		if _, err := gph.ReadFrom(r); err != nil {
			return nil
		}
		for _, e := range gph.Entries {
			for _, r := range e.References {
				fmt.Fprintf(w, "%s %s\n", segmentID(e.Msb, e.Lsb), segmentID(r.Msb, r.Lsb))
			}
		}
		return nil
	}
}

func doPrintIndex(f format, w io.Writer) handler {
	switch f {
	case formatHex:
		return doPrintHexTo(w)
	case formatText:
		return doPrintIndexTo(w)
	default:
		return invalidFormat()
	}
}

func doPrintIndexTo(w io.Writer) handler {
	return func(_ string, r io.Reader) error {
		var idx index.Index
		if _, err := idx.ReadFrom(r); err != nil {
			return err
		}
		for _, e := range idx.Entries {
			id := segmentID(e.Msb, e.Lsb)
			fmt.Fprintf(w, "%s %s %x %d %d\n", segmentType(id), id, e.Position, e.Size, e.Generation)
		}
		return nil
	}
}

func doPrintSegmentNameTo(w io.Writer) handler {
	return func(n string, _ io.Reader) error {
		id := normalizeSegmentID(entryNameToSegmentID(n))
		fmt.Fprintf(w, "%s %s\n", segmentType(id), id)
		return nil
	}
}

func doPrintSegment(f format, w io.Writer) handler {
	switch f {
	case formatHex:
		return doPrintHexTo(w)
	case formatText:
		return doPrintSegmentTo(w)
	default:
		return invalidFormat()
	}
}

func doPrintSegmentTo(w io.Writer) handler {
	return func(_ string, r io.Reader) error {
		var s segment.Segment
		if _, err := s.ReadFrom(r); err != nil {
			return err
		}
		fmt.Fprintf(w, "version %d\n", s.Version)
		fmt.Fprintf(w, "generation %d\n", s.Generation)
		for i, r := range s.References {
			fmt.Fprintf(w, "reference %d %s\n", i+1, segmentID(r.Msb, r.Lsb))
		}
		for _, r := range s.Records {
			fmt.Fprintf(w, "record %x %s %x\n", r.Number, recordType(r.Type), r.Offset)
		}
		return nil
	}
}

func doPrintNameTo(w io.Writer) handler {
	return func(n string, _ io.Reader) error {
		fmt.Fprintln(w, n)
		return nil
	}
}

func doPrintHexTo(w io.Writer) handler {
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

func recordType(t segment.RecordType) string {
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
	case segment.RecordTypeBlobID:
		return "binary"
	default:
		return "unknown"
	}
}

func segmentType(id string) string {
	if isBulkSegmentID(id) {
		return "bulk"
	}
	return "data"
}

func segmentID(msb, lsb uint64) string {
	return fmt.Sprintf("%016x%016x", msb, lsb)
}
