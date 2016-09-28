package sdb

import (
	"archive/tar"
	"encoding/binary"
	"fmt"
	"hash/crc32"
	"io"
	"io/ioutil"
	"os"
	"strings"
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
	data, err := ioutil.ReadAll(reader)

	if err != nil {
		return err
	}

	n := len(data)

	if n < indexFooterSize {
		return fmt.Errorf("Invalid data")
	}

	var (
		footer   = data[n-indexFooterSize:]
		checksum = int(binary.BigEndian.Uint32(footer[indexFooterChecksumOffset:]))
		count    = int(binary.BigEndian.Uint32(footer[indexFooterCountOffset:]))
		size     = int(binary.BigEndian.Uint32(footer[indexFooterSizeOffset:]))
		magic    = int(binary.BigEndian.Uint32(footer[indexFooterMagicOffset:]))
	)

	if magic != indexMagic {
		return fmt.Errorf("Invalid magic %08x", magic)
	}

	if size < count*indexEntrySize+indexFooterSize {
		return fmt.Errorf("Invalid count or size")
	}

	if n < count*indexEntrySize+indexFooterSize {
		return fmt.Errorf("Invalid count or data size")
	}

	entries := data[n-indexFooterSize-count*indexEntrySize : n-indexFooterSize]

	if int(crc32.ChecksumIEEE(entries)) != checksum {
		return fmt.Errorf("Invalid checkusm")
	}

	for i := 0; i < count; i++ {
		entry := entries[i*indexEntrySize:]

		var (
			msb        = binary.BigEndian.Uint64(entry[indexEntryMsbOffset:])
			lsb        = binary.BigEndian.Uint64(entry[indexEntryLsbOffset:])
			position   = int(binary.BigEndian.Uint32(entry[indexEntryPositionOffset:]))
			size       = int(binary.BigEndian.Uint32(entry[indexEntrySizeOffset:]))
			generation = int(binary.BigEndian.Uint32(entry[indexEntryGenerationOffset:]))
		)

		id := fmt.Sprintf("%016x%016x", msb, lsb)

		kind := "data"

		if isBulkSegmentID(id) {
			kind = "bulk"
		}

		fmt.Fprintf(writer, "%s %s %8x %6d %6d\n", kind, id, position, size, generation)
	}

	return nil
}
