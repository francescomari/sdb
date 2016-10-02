package sdb

import (
	"archive/tar"
	"fmt"
	"io"
	"os"
	"strings"

	"github.com/francescomari/sdb/index"
)

const (
	indexMagic      = 0x0a304b0a
	indexFooterSize = 16
	indexEntrySize  = 28
)

const (
	indexFooterChecksumOffset = 0
	indexFooterCountOffset    = 4
	indexFooterSizeOffset     = 8
	indexFooterMagicOffset    = 12
)

const (
	indexEntryMsbOffset        = 0
	indexEntryLsbOffset        = 8
	indexEntryPositionOffset   = 16
	indexEntrySizeOffset       = 20
	indexEntryGenerationOffset = 24
)

// PrintIndex prints the content of the index from the TAR file at 'path' to
// 'writer'. The desired output format is specified by 'output'.
func PrintIndex(path string, writer io.Writer, output OutputType) error {
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

		if !strings.HasSuffix(header.Name, ".idx") {
			continue
		}

		if output == OutputHex {
			return printHex(reader, writer)
		}

		if output == OutputText {
			return printIndexText(reader, writer)
		}

		return fmt.Errorf("Invalid output type")
	}

	return nil
}

func printIndexText(reader io.Reader, writer io.Writer) error {
	var idx index.Index

	if _, err := idx.ReadFrom(reader); err != nil {
		return err
	}

	for _, e := range idx.Entries {
		id := fmt.Sprintf("%016x%016x", e.Msb, e.Lsb)

		kind := "data"

		if isBulkSegmentID(id) {
			kind = "bulk"
		}

		fmt.Fprintf(writer, "%s %s %8x %6d %6d\n", kind, id, e.Position, e.Size, e.Generation)
	}

	return nil
}
