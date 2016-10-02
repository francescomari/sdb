package sdb

import (
	"archive/tar"
	"fmt"
	"io"
	"os"
	"strings"

	"github.com/francescomari/sdb/binaries"
)

// PrintBinaries prints the content of the binary references index from the TAR
// file at 'path' to 'writer'. The desired output format is specified by
// 'output'.
func PrintBinaries(path string, writer io.Writer, output OutputType) error {
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

		if !strings.HasSuffix(header.Name, ".brf") {
			continue
		}

		if output == OutputHex {
			return printHex(reader, writer)
		}

		if output == OutputText {
			return printBinariesText(reader, writer)
		}

		return fmt.Errorf("Invalid output type")
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
