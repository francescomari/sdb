package sdb

import (
	"archive/tar"
	"fmt"
	"io"
	"os"
	"strings"
)

// OutputType represents a particular kind of output format.
type OutputType int

const (
	// OutputHex prints the output in a hexadecimal dump.
	OutputHex OutputType = iota
	// OutputText prints the output in a plain, human readable form.
	OutputText
)

// PrintEntries prints the name of the entries from the TAR file at 'path' to
// 'writer'. The entries are printed in the same order as they are stored in the
// TAR file.
func PrintEntries(path string, writer io.Writer) error {
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

		fmt.Fprintln(writer, header.Name)
	}

	return nil
}

func entryNameToSegmentID(header string) string {
	return header[:strings.Index(header, ".")]
}
