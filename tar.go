package sdb

import (
	"archive/tar"
	"fmt"
	"io"
	"os"
	"strings"
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

// DumpEntry prints to 'writer' the hexdump of the first entry in the TAR file
// at 'path' matching the criteria 'filter'.
func DumpEntry(path string, filter func(string) bool, writer io.Writer) error {
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

		if filter(header.Name) {
			return printHex(reader, writer)
		}
	}

	return nil
}

func entryNameToSegmentID(header string) string {
	return header[:strings.Index(header, ".")]
}
