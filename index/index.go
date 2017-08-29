package index

import (
	"bytes"
	"encoding/binary"
	"fmt"
	"hash/crc32"
	"io"
)

// Index is a catalog of every segment stored in a TAR file.
type Index struct {
	Entries []Entry
}

// Entry is a reference to a segment. An Index is composed of one or more
// entries.
type Entry struct {
	Msb            uint64
	Lsb            uint64
	Position       int
	Size           int
	Generation     int
	FullGeneration int
	Compacted      bool
}

// ReadFrom reads the index from 'r' and returns the number of bytes read and an
// error.
func (index *Index) ReadFrom(r io.Reader) (int64, error) {
	var b bytes.Buffer

	n, err := b.ReadFrom(r)

	if err != nil {
		return n, err
	}

	err = index.parse(b.Bytes())

	if err != nil {
		return n, err
	}

	return n, nil
}

const (
	v1Magic = 0x0a304b0a
	v2Magic = 0x0a314b0a
)

func (index *Index) parse(data []byte) error {
	n := len(data)

	if n < 4 {
		return fmt.Errorf("invalid data")
	}

	magic := int(binary.BigEndian.Uint32(data[n-4:]))

	if magic == v1Magic {
		return index.parseV1(data)
	}
	if magic == v2Magic {
		return index.parseV2(data)
	}

	return fmt.Errorf("unrecognized magic %08x", magic)
}

func (index *Index) parseV1(data []byte) error {
	const (
		indexMagic     = v1Magic
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

	n := len(data)

	if n < footerSize {
		return fmt.Errorf("invalid data")
	}

	var (
		footer   = data[n-footerSize:]
		checksum = int(binary.BigEndian.Uint32(footer[footerChecksumOffset:]))
		count    = int(binary.BigEndian.Uint32(footer[footerCountOffset:]))
		size     = int(binary.BigEndian.Uint32(footer[footerSizeOffset:]))
		magic    = int(binary.BigEndian.Uint32(footer[footerMagicOffset:]))
	)

	if magic != indexMagic {
		return fmt.Errorf("invalid magic %08x", magic)
	}

	if size < count*indexEntrySize+footerSize {
		return fmt.Errorf("invalid count or size")
	}

	if n < count*indexEntrySize+footerSize {
		return fmt.Errorf("invalid count or data size")
	}

	entries := data[n-footerSize-count*indexEntrySize : n-footerSize]

	if int(crc32.ChecksumIEEE(entries)) != checksum {
		return fmt.Errorf("invalid checkusm")
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

		index.Entries = append(index.Entries, Entry{
			Msb:            msb,
			Lsb:            lsb,
			Position:       position,
			Size:           size,
			Generation:     generation,
			FullGeneration: generation,
			Compacted:      true,
		})
	}

	return nil
}

func (index *Index) parseV2(data []byte) error {
	const (
		indexMagic     = v2Magic
		footerSize     = 16
		indexEntrySize = 33
	)

	const (
		footerChecksumOffset = 0
		footerCountOffset    = 4
		footerSizeOffset     = 8
		footerMagicOffset    = 12
	)

	const (
		entryMsbOffset            = 0
		entryLsbOffset            = 8
		entryPositionOffset       = 16
		entrySizeOffset           = 20
		entryGenerationOffset     = 24
		entryFullGenerationOffset = 28
		entryCompactedOffset      = 32
	)

	n := len(data)

	if n < footerSize {
		return fmt.Errorf("invalid data")
	}

	var (
		footer   = data[n-footerSize:]
		checksum = int(binary.BigEndian.Uint32(footer[footerChecksumOffset:]))
		count    = int(binary.BigEndian.Uint32(footer[footerCountOffset:]))
		size     = int(binary.BigEndian.Uint32(footer[footerSizeOffset:]))
		magic    = int(binary.BigEndian.Uint32(footer[footerMagicOffset:]))
	)

	if magic != indexMagic {
		return fmt.Errorf("invalid magic %08x", magic)
	}

	if size < count*indexEntrySize+footerSize {
		return fmt.Errorf("invalid count or size")
	}

	if n < count*indexEntrySize+footerSize {
		return fmt.Errorf("invalid count or data size")
	}

	entries := data[n-footerSize-count*indexEntrySize : n-footerSize]

	if int(crc32.ChecksumIEEE(entries)) != checksum {
		return fmt.Errorf("invalid checkusm")
	}

	index.Entries = nil

	for i := 0; i < count; i++ {
		entry := entries[i*indexEntrySize:]

		var (
			msb            = binary.BigEndian.Uint64(entry[entryMsbOffset:])
			lsb            = binary.BigEndian.Uint64(entry[entryLsbOffset:])
			position       = int(binary.BigEndian.Uint32(entry[entryPositionOffset:]))
			size           = int(binary.BigEndian.Uint32(entry[entrySizeOffset:]))
			generation     = int(binary.BigEndian.Uint32(entry[entryGenerationOffset:]))
			fullGeneration = int(binary.BigEndian.Uint32(entry[entryFullGenerationOffset:]))
			compacted      = int(entry[entryCompactedOffset]) != 0
		)

		index.Entries = append(index.Entries, Entry{
			Msb:            msb,
			Lsb:            lsb,
			Position:       position,
			Size:           size,
			Generation:     generation,
			FullGeneration: fullGeneration,
			Compacted:      compacted,
		})
	}

	return nil
}
