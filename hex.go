package sdb

import (
	"bufio"
	"fmt"
	"io"
)

const (
	hexLineLength = 32
	hexStopLength = 8
)

func printHex(reader io.Reader, writer io.Writer) error {
	line, data, breader := 0, make([]byte, 0), bufio.NewReader(reader)

	for {
		c, err := breader.ReadByte()

		if err == io.EOF {
			break
		}

		if err != nil {
			return err
		}

		data = append(data, c)

		if len(data) < hexLineLength {
			continue
		}

		printHexLine(line, data, writer)

		line, data = line+1, make([]byte, 0)
	}

	if len(data) == 0 {
		return nil
	}

	printHexLine(line, data, writer)

	return nil
}

func printHexLine(line int, data []byte, writer io.Writer) {
	fmt.Fprintf(writer, "%08x ", line*hexLineLength)

	for i := 0; i < hexLineLength; i++ {
		if i%8 == 0 {
			fmt.Fprintf(writer, " ")
		}

		if i < len(data) {
			fmt.Fprintf(writer, "%02x ", data[i])
		} else {
			fmt.Fprintf(writer, "   ")
		}
	}

	fmt.Fprintf(writer, " |")

	for i := 0; i < hexLineLength; i++ {
		if i < len(data) && 0x21 <= data[i] && data[i] <= 0x7e {
			fmt.Fprintf(writer, "%c", data[i])
		} else {
			fmt.Fprintf(writer, " ")
		}
	}

	fmt.Fprintf(writer, "|\n")
}
