package binaries

import (
	"bytes"
	"encoding/binary"
	"fmt"
	"hash/crc32"
	"io"
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

type Binaries struct {
	Generations []Generation
}

type Generation struct {
	Generation int
	Segments   []Segment
}

type Segment struct {
	Msb        uint64
	Lsb        uint64
	References []string
}

func (binaries *Binaries) ReadFrom(r io.Reader) (int64, error) {
	var b bytes.Buffer

	n, err := b.ReadFrom(r)

	if err != nil {
		return n, err
	}

	err = binaries.parseFrom(b.Bytes())

	if err != nil {
		return n, err
	}

	return n, nil
}

func (binaries *Binaries) parseFrom(data []byte) error {
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

	buffer := bytes.NewBuffer(entries)

	for i := 0; i < count; i++ {
		var generation Generation
		generation.parseFrom(buffer)
		binaries.Generations = append(binaries.Generations, generation)
	}

	return nil
}

func (generation *Generation) parseFrom(b *bytes.Buffer) {
	data := b.Next(binariesGenerationSize)

	generation.Generation = int(binary.BigEndian.Uint32(data[binariesGenerationNumberOffset:]))
	nSegments := int(binary.BigEndian.Uint32(data[binariesGenerationCountOffset:]))

	for i := 0; i < nSegments; i++ {
		var segment Segment
		segment.parseFrom(b)
		generation.Segments = append(generation.Segments, segment)
	}
}

func (segment *Segment) parseFrom(b *bytes.Buffer) {
	data := b.Next(binariesSegmentSize)

	segment.Msb = binary.BigEndian.Uint64(data[binariesSegmentMsbOffset:])
	segment.Lsb = binary.BigEndian.Uint64(data[binariesSegmentLsbOffset:])
	nReferences := int(binary.BigEndian.Uint32(data[binariesSegmentCountOffset:]))

	for i := 0; i < nReferences; i++ {
		data := b.Next(binariesReferenceSize)
		size := int(binary.BigEndian.Uint32(data[binariesReferenceSizeOffset:]))
		segment.References = append(segment.References, string(b.Next(size)))
	}
}
