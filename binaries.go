package sdb

import (
	"archive/tar"
	"fmt"
	"io"
	"os"
	"strings"

	"github.com/francescomari/sdb/binaries"
)

// BinariesEntryFilter locates the binary references index in a TAR file.
func BinariesEntryFilter(name string) bool {
	return strings.HasSuffix(name, ".brf")
}

// PrintBinaries prints the content of the binary references index from the TAR
// file at 'path' to 'writer'.
func PrintBinaries(path string, writer io.Writer) error {
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

		if BinariesEntryFilter(header.Name) {
			return printBinariesText(reader, writer)
		}
	}

	return nil
}

func printBinariesText(reader io.Reader, writer io.Writer) error {
	var bns binaries.Binaries

	if _, err := bns.ReadFrom(reader); err != nil {
		return err
	}

	for _, generation := range bns.Generations {
		fmt.Fprintf(writer, "%d\n", generation.Generation)

		for _, segment := range generation.Segments {
			fmt.Fprintf(writer, "    %016x%016x\n", segment.Msb, segment.Lsb)

			for _, reference := range segment.References {
				fmt.Fprintf(writer, "        %s\n", reference)
			}
		}
	}

	return nil
}
