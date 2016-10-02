package index

import (
	"bytes"
	"encoding/binary"
	"fmt"
	"hash/crc32"
	"io"
)

const (
	indexMagic     = 0x0a304b0a
	footerSize     = 16
	indexEntrySize = 28
)

const (
	footerChecksumOffset = 0
	footerCountOffset    = 4
	footerSizeOffset     = 8
	footerMagicOffset    = 12
)

const (
	entryMsbOffset        = 0
	entryLsbOffset        = 8
	entryPositionOffset   = 16
	entrySizeOffset       = 20
	entryGenerationOffset = 24
)

// Index is a catalog of every segment stored in a TAR file.
type Index struct {
	Entries []Entry
}

// Entry is a reference to a segment. An Index is composed of one or more
// entries.
type Entry struct {
	Msb        uint64
	Lsb        uint64
	Position   int
	Size       int
	Generation int
}

// ReadFrom reads the index from 'r' and returns the number of bytes read and an
// error.
func (index *Index) ReadFrom(r io.Reader) (int64, error) {
	var b bytes.Buffer

	n, err := b.ReadFrom(r)

	if err != nil {
		return n, err
	}

	err = index.parseFrom(b.Bytes())

	if err != nil {
		return n, err
	}

	return n, nil
}

func (index *Index) parseFrom(data []byte) error {
	n := len(data)

	if n < footerSize {
		return fmt.Errorf("Invalid data")
	}

	var (
		footer   = data[n-footerSize:]
		checksum = int(binary.BigEndian.Uint32(footer[footerChecksumOffset:]))
		count    = int(binary.BigEndian.Uint32(footer[footerCountOffset:]))
		size     = int(binary.BigEndian.Uint32(footer[footerSizeOffset:]))
		magic    = int(binary.BigEndian.Uint32(footer[footerMagicOffset:]))
	)

	if magic != indexMagic {
		return fmt.Errorf("Invalid magic %08x", magic)
	}

	if size < count*indexEntrySize+footerSize {
		return fmt.Errorf("Invalid count or size")
	}

	if n < count*indexEntrySize+footerSize {
		return fmt.Errorf("Invalid count or data size")
	}

	entries := data[n-footerSize-count*indexEntrySize : n-footerSize]

	if int(crc32.ChecksumIEEE(entries)) != checksum {
		return fmt.Errorf("Invalid checkusm")
	}

	index.Entries = nil

	for i := 0; i < count; i++ {
		entry := entries[i*indexEntrySize:]

		var (
			msb        = binary.BigEndian.Uint64(entry[entryMsbOffset:])
			lsb        = binary.BigEndian.Uint64(entry[entryLsbOffset:])
			position   = int(binary.BigEndian.Uint32(entry[entryPositionOffset:]))
			size       = int(binary.BigEndian.Uint32(entry[entrySizeOffset:]))
			generation = int(binary.BigEndian.Uint32(entry[entryGenerationOffset:]))
		)

		index.Entries = append(index.Entries, Entry{msb, lsb, position, size, generation})
	}

	return nil
}
