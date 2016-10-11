package sdb

import (
	"archive/tar"
	"errors"
	"io"
	"os"
)

type handler func(n string, r io.Reader) error

type matcher func(string) bool

var errStop = errors.New("stop")

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
			if err := h(hdr.Name, r); err == errStop {
				return nil
			} else if err != nil {
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
	return forEachMatchingEntry(p, m, func(n string, r io.Reader) error {
		if err := h(n, r); err != nil {
			return err
		}
		return errStop
	})
}
