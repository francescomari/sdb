package sdb

import (
	"archive/tar"
	"fmt"
	"io"
	"os"
	"strings"

	"github.com/francescomari/sdb/graph"
)

// GraphEntryFilter locates the graph in a TAR file.
func GraphEntryFilter(name string) bool {
	return strings.HasSuffix(name, ".gph")
}

// PrintGraph prints the content of the graph from the TAR file at 'path' to
// 'writer'.
func PrintGraph(path string, writer io.Writer) error {
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

		if GraphEntryFilter(header.Name) {
			return printGraphText(reader, writer)
		}
	}

	return nil
}

func printGraphText(reader io.Reader, writer io.Writer) error {
	var gph graph.Graph

	if _, err := gph.ReadFrom(reader); err != nil {
		return nil
	}

	for _, entry := range gph.Entries {
		fmt.Fprintf(writer, "%016x%016x\n", entry.Msb, entry.Lsb)

		for _, reference := range entry.References {
			fmt.Fprintf(writer, "    %016x%016x\n", reference.Msb, reference.Lsb)
		}
	}

	return nil
}
