package sdb

import (
	"fmt"
	"io"
	"strings"

	"github.com/francescomari/sdb/graph"
)

// GraphEntryFilter locates the graph in a TAR file.
func GraphEntryFilter(name string) bool {
	return strings.HasSuffix(name, ".gph")
}

// PrintGraph prints the content of the graph from the TAR file at 'p' to 'w'.
func PrintGraph(path string, writer io.Writer) error {
	return onMatchingEntry(path, GraphEntryFilter, printGraphTo(writer))
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
