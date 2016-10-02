package graph

import (
	"bytes"
	"encoding/binary"
	"fmt"
	"hash/crc32"
	"io"
)

const (
	graphMagic = 0x0a30470a
	footerSize = 16
	keySize    = 20
	valueSize  = 16
)

const (
	footerChecksumOffset = 0
	footerCountOffset    = 4
	footerSizeOffset     = 8
	footerMagicOffset    = 12
)

const (
	entryMsbOffset   = 0
	entryLsbOffset   = 8
	entryCountOffset = 16
)

const (
	referenceMsbOffset = 0
	referenceLsbOffset = 8
)

// Graph is a graph of references beetween segments.
type Graph struct {
	Entries []Entry
}

// Entry is a collection of a segment and its references towards other segments.
type Entry struct {
	Msb        uint64
	Lsb        uint64
	References []Reference
}

// Reference is the destinatin of a reference from another segment.
type Reference struct {
	Msb uint64
	Lsb uint64
}

// ReadFrom reads the content of the graph from the provided reader. It returns
// the number of bytes read and an optional error.
func (graph *Graph) ReadFrom(r io.Reader) (int64, error) {
	var b bytes.Buffer

	n, err := b.ReadFrom(r)

	if err != nil {
		return n, err
	}

	err = graph.parseFrom(b.Bytes())

	if err != nil {
		return n, err
	}

	return n, nil
}

func (graph *Graph) parseFrom(data []byte) error {
	n := len(data)

	if n < footerSize {
		return fmt.Errorf("Invalid data")
	}

	var (
		footer   = data[n-footerSize:]
		checksum = int(binary.BigEndian.Uint32(footer[footerChecksumOffset:]))
		nEntries = int(binary.BigEndian.Uint32(footer[footerCountOffset:]))
		size     = int(binary.BigEndian.Uint32(footer[footerSizeOffset:]))
		magic    = int(binary.BigEndian.Uint32(footer[footerMagicOffset:]))
	)

	if magic != graphMagic {
		return fmt.Errorf("Invalid magic")
	}

	if size < footerSize {
		return fmt.Errorf("Invalid size")
	}

	if nEntries < 0 {
		return fmt.Errorf("Invalid entries")
	}

	entries := data[n-size : n-footerSize]

	if int(crc32.ChecksumIEEE(entries)) != checksum {
		return fmt.Errorf("Invalid checksum")
	}

	buffer := bytes.NewBuffer(entries)

	for i := 0; i < nEntries; i++ {
		var entry Entry
		entry.parseFrom(buffer)
		graph.Entries = append(graph.Entries, entry)
	}

	return nil
}

func (entry *Entry) parseFrom(b *bytes.Buffer) {
	data := b.Next(keySize)

	entry.Msb = binary.BigEndian.Uint64(data[entryMsbOffset:])
	entry.Lsb = binary.BigEndian.Uint64(data[entryLsbOffset:])

	n := int(binary.BigEndian.Uint32(data[entryCountOffset:]))

	for i := 0; i < n; i++ {
		var reference Reference
		reference.parseFrom(b)
		entry.References = append(entry.References, reference)
	}
}

func (reference *Reference) parseFrom(b *bytes.Buffer) {
	data := b.Next(valueSize)

	reference.Msb = binary.BigEndian.Uint64(data[referenceMsbOffset:])
	reference.Lsb = binary.BigEndian.Uint64(data[referenceLsbOffset:])
}
