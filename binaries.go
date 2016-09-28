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
	binariesMagic          = 0x0a30420a
	binariesFooterSize     = 16
	binariesGenerationSize = 8
	binariesSegmentSize    = 20
	binariesReferenceSize  = 4
)

const (
	binariesFooterChecksumOffset = 0
	binariesFooterCountOffset    = 4
	binariesFooterSizeOffset     = 8
	binariesFooterMagicOffset    = 12
)

const (
	binariesGenerationNumberOffset = 0
	binariesGenerationCountOffset  = 4
)

const (
	binariesSegmentMsbOffset   = 0
	binariesSegmentLsbOffset   = 8
	binariesSegmentCountOffset = 16
)

const (
	binariesReferenceSizeOffset  = 0
	binariesReferenceValueOffset = 4
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
	data, err := ioutil.ReadAll(reader)

	if err != nil {
		return err
	}

	n := len(data)

	if n < binariesFooterSize {
		return fmt.Errorf("Invalid data")
	}

	var (
		footer   = data[n-binariesFooterSize:]
		checksum = int(binary.BigEndian.Uint32(footer[binariesFooterChecksumOffset:]))
		count    = int(binary.BigEndian.Uint32(footer[binariesFooterCountOffset:]))
		size     = int(binary.BigEndian.Uint32(footer[binariesFooterSizeOffset:]))
		magic    = int(binary.BigEndian.Uint32(footer[binariesFooterMagicOffset:]))
	)

	if magic != binariesMagic {
		return fmt.Errorf("Invalid magic")
	}

	if size < binariesFooterSize {
		return fmt.Errorf("Invalid size")
	}

	if count < 0 {
		return fmt.Errorf("Invalid count")
	}

	entries := data[n-size : n-binariesFooterSize]

	if int(crc32.ChecksumIEEE(entries)) != checksum {
		return fmt.Errorf("Invalid checksum")
	}

	for i := 0; i < count; i++ {
		var (
			generation = int(binary.BigEndian.Uint32(entries[binariesGenerationNumberOffset:]))
			nSegments  = int(binary.BigEndian.Uint32(entries[binariesGenerationCountOffset:]))
		)

		entries = entries[binariesGenerationSize:]

		fmt.Fprintf(writer, "%d\n", generation)

		for j := 0; j < nSegments; j++ {
			var (
				msb         = binary.BigEndian.Uint64(entries[binariesSegmentMsbOffset:])
				lsb         = binary.BigEndian.Uint64(entries[binariesSegmentLsbOffset:])
				nReferences = int(binary.BigEndian.Uint32(entries[binariesSegmentCountOffset:]))
			)

			entries = entries[binariesSegmentSize:]

			fmt.Fprintf(writer, "    %016x%016x\n", msb, lsb)

			for k := 0; k < nReferences; k++ {
				var (
					size      = int(binary.BigEndian.Uint32(entries[binariesReferenceSizeOffset:]))
					reference = entries[binariesReferenceValueOffset : binariesReferenceValueOffset+size]
				)

				entries = entries[binariesReferenceSize+size:]

				fmt.Fprintf(writer, "        %s\n", string(reference))
			}
		}
	}

	return nil
}
