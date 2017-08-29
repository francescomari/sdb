package binaries

import (
	"bytes"
	"encoding/binary"
	"fmt"
	"hash/crc32"
	"io"
)

// Binaries is the set of binary references in a TAR files grouped by generation
// and segment.
type Binaries struct {
	Generations []Generation
}

// Generation is the set of binary references belonging to a specific
// generation. The binary references are grouped by segment.
type Generation struct {
	Generation     int
	FullGeneration int
	Compacted      bool
	Segments       []Segment
}

// Segment is the set of binary references belonging to a specific segment.
type Segment struct {
	Msb        uint64
	Lsb        uint64
	References []string
}

// ReadFrom reads the binary references from its serialized representation.
// Returns the number of bytes read and an optional error.
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

const (
	magicV1 = 0x0a30420a
	magicV2 = 0x0a31420a
)

func (binaries *Binaries) parseFrom(data []byte) error {
	n := len(data)

	if n < 4 {
		return fmt.Errorf("invalid data")
	}

	magic := int(binary.BigEndian.Uint32(data[n-4:]))

	if magic == magicV1 {
		return binaries.parseV1From(data)
	}
	if magic == magicV2 {
		return binaries.parseV2From(data)
	}

	return fmt.Errorf("unrecognized magic %08x", magic)
}

func (binaries *Binaries) parseV1From(data []byte) error {
	const (
		binariesMagic          = magicV1
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

	binaries.Generations = nil

	buffer := bytes.NewBuffer(entries)

	for i := 0; i < count; i++ {
		var (
			generation    = int(binary.BigEndian.Uint32(buffer.Next(4)))
			segmentsCount = int(binary.BigEndian.Uint32(buffer.Next(4)))
			segments      = make([]Segment, segmentsCount)
		)

		for i := 0; i < segmentsCount; i++ {
			var (
				msb             = binary.BigEndian.Uint64(buffer.Next(8))
				lsb             = binary.BigEndian.Uint64(buffer.Next(8))
				referencesCount = int(binary.BigEndian.Uint32(buffer.Next(4)))
				references      = make([]string, referencesCount)
			)

			for i := 0; i < referencesCount; i++ {
				var (
					size      = int(binary.BigEndian.Uint32(buffer.Next(4)))
					reference = string(buffer.Next(size))
				)

				references[i] = reference
			}

			segments = append(segments, Segment{
				Msb:        msb,
				Lsb:        lsb,
				References: references,
			})
		}

		binaries.Generations = append(binaries.Generations, Generation{
			Generation:     generation,
			FullGeneration: generation,
			Compacted:      true,
			Segments:       segments,
		})
	}

	return nil
}

func (binaries *Binaries) parseV2From(data []byte) error {
	const (
		binariesMagic          = magicV2
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

	binaries.Generations = make([]Generation, count)

	buffer := bytes.NewBuffer(entries)

	for i := 0; i < count; i++ {
		var (
			generation     = int(binary.BigEndian.Uint32(buffer.Next(4)))
			fullGeneration = int(binary.BigEndian.Uint32(buffer.Next(4)))
			compacted      = buffer.Next(1)[0] != 0
			segmentsCount  = int(binary.BigEndian.Uint32(buffer.Next(4)))
			segments       = make([]Segment, segmentsCount)
		)

		for i := 0; i < segmentsCount; i++ {
			var (
				msb             = binary.BigEndian.Uint64(buffer.Next(8))
				lsb             = binary.BigEndian.Uint64(buffer.Next(8))
				referencesCount = int(binary.BigEndian.Uint32(buffer.Next(4)))
				references      = make([]string, referencesCount)
			)

			for i := 0; i < referencesCount; i++ {
				var (
					size      = int(binary.BigEndian.Uint32(buffer.Next(4)))
					reference = string(buffer.Next(size))
				)

				references[i] = reference
			}

			segments[i] = Segment{
				Msb:        msb,
				Lsb:        lsb,
				References: references,
			}
		}

		binaries.Generations[i] = Generation{
			Generation:     generation,
			FullGeneration: fullGeneration,
			Compacted:      compacted,
			Segments:       segments,
		}
	}

	return nil
}
