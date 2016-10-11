package sdb

import (
	"fmt"
	"io"
	"strings"

	"github.com/francescomari/sdb/binaries"
)

// BinariesEntryFilter locates the binary references index in a TAR file.
func BinariesEntryFilter(name string) bool {
	return strings.HasSuffix(name, ".brf")
}

// PrintBinaries prints the content of the binary references index from the TAR
// file at 'p' to 'w'.
func PrintBinaries(p string, w io.Writer) error {
	return onMatchingEntry(p, BinariesEntryFilter, printBinariesTo(w))
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
