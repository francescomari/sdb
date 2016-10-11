package sdb

import (
	"archive/tar"
	"errors"
	"fmt"
	"io"
	"os"
	"strings"
)

// PrintEntries prints the name of the entries from the TAR file at 'p' to 'w'.
// The entries are printed in the same order as they are stored in the TAR file.
func PrintEntries(p string, w io.Writer) error {
	return forEachEntry(p, printNameTo(w))
}

// DumpEntry prints to 'w' the hexdump of the first entry in the TAR file at 'p'
// matching the criteria 'm'.
func DumpEntry(p string, m func(string) bool, w io.Writer) error {
	return onMatchingEntry(p, m, printHexTo(w))
}

func printNameTo(w io.Writer) handler {
	return func(n string, _ io.Reader) error {
		fmt.Fprintln(w, n)
		return nil
	}
}

func printHexTo(w io.Writer) handler {
	return func(_ string, r io.Reader) error {
		return printHex(r, w)
	}
}

type handler func(n string, r io.Reader) error

type matcher func(string) bool

var errStop = errors.New("stop")

func any(_ string) bool {
	return true
}

func failWith(h handler, f error) handler {
	return func(n string, r io.Reader) error {
		if err := h(n, r); err != nil {
			return err
		}
		return f
	}
}

func forEachMatchingEntry(p string, m matcher, h handler) error {
	f, err := os.Open(p)

	if err != nil {
		return err
	}

	defer f.Close()

	r := tar.NewReader(f)

	for {
		hdr, err := r.Next()

		if hdr == nil {
			break
		}

		if err != nil {
			return err
		}

		if m(hdr.Name) {
			if err := h(hdr.Name, r); err != nil {
				return err
			}
		}
	}

	return nil
}

func forEachEntry(p string, h handler) error {
	return forEachMatchingEntry(p, any, h)
}

func onMatchingEntry(p string, m matcher, h handler) error {
	err := forEachMatchingEntry(p, m, failWith(h, errStop))

	if err == errStop {
		return nil
	}

	return err
}

func entryNameToSegmentID(header string) string {
	return header[:strings.Index(header, ".")]
}
