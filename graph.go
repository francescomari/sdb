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
	graphMagic      = 0x0a30470a
	graphFooterSize = 16
	graphKeySize    = 20
	graphValueSize  = 16
)

const (
	graphFooterChecksumOffset = 0
	graphFooterCountOffset    = 4
	graphFooterSizeOffset     = 8
	graphFooterMagicOffset    = 12
)

const (
	graphKeyMsbOffset   = 0
	graphKeyLsbOffset   = 8
	graphKeyCountOffset = 16
)

const (
	graphValueMsbOffset = 0
	graphValueLsbOffset = 8
)

// PrintGraph prints the content of the graph from the TAR file at 'path' to
// 'writer'. The desired output format is specified by 'output'.
func PrintGraph(path string, writer io.Writer, output OutputType) error {
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

		if !strings.HasSuffix(header.Name, ".gph") {
			continue
		}

		if output == OutputHex {
			return printHex(reader, writer)
		}

		if output == OutputText {
			return printGraphText(reader, writer)
		}

		return fmt.Errorf("Invalid output type")
	}

	return nil
}

func printGraphText(reader io.Reader, writer io.Writer) error {
	data, err := ioutil.ReadAll(reader)

	if err != nil {
		return err
	}

	n := len(data)

	if n < graphFooterSize {
		return fmt.Errorf("Invalid data")
	}

	var (
		footer   = data[n-graphFooterSize:]
		checksum = int(binary.BigEndian.Uint32(footer[graphFooterChecksumOffset:]))
		count    = int(binary.BigEndian.Uint32(footer[graphFooterCountOffset:]))
		size     = int(binary.BigEndian.Uint32(footer[graphFooterSizeOffset:]))
		magic    = int(binary.BigEndian.Uint32(footer[graphFooterMagicOffset:]))
	)

	if magic != graphMagic {
		return fmt.Errorf("Invalid magic")
	}

	if size < graphFooterSize {
		return fmt.Errorf("Invalid size")
	}

	if count < 0 {
		return fmt.Errorf("Invalid count")
	}

	entries := data[n-size : n-graphFooterSize]

	if int(crc32.ChecksumIEEE(entries)) != checksum {
		return fmt.Errorf("Invalid checksum")
	}

	for i := 0; i < count; i++ {
		var (
			msb  = binary.BigEndian.Uint64(entries[graphKeyMsbOffset:])
			lsb  = binary.BigEndian.Uint64(entries[graphKeyLsbOffset:])
			more = int(binary.BigEndian.Uint32(entries[graphKeyCountOffset:]))
		)

		entries = entries[graphKeySize:]

		fmt.Fprintf(writer, "%016x%016x\n", msb, lsb)

		for j := 0; j < more; j++ {
			var (
				msb = binary.BigEndian.Uint64(entries[graphValueMsbOffset:])
				lsb = binary.BigEndian.Uint64(entries[graphValueLsbOffset:])
			)

			entries = entries[graphValueSize:]

			fmt.Fprintf(writer, "    %016x%016x\n", msb, lsb)
		}
	}

	return nil
}
