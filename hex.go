package sdb

import (
	"encoding/hex"
	"io"
)

func printHex(r io.Reader, w io.Writer) (err error) {
	d := hex.Dumper(w)
	defer d.Close()
	_, err = io.Copy(d, r)
	return
}
