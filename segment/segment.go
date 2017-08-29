package segment

import (
	"bytes"
	"encoding/binary"
	"fmt"
	"io"
)

// A Segment is a container for records.
type Segment struct {
	Version        int
	Generation     int
	FullGeneration int
	Compacted      bool
	References     []Reference
	Records        []Record
}

// A Reference represents a link towards another segment.
type Reference struct {
	Msb uint64
	Lsb uint64
}

// A Record is a pointer to a record stored in a segment.
type Record struct {
	Number int
	Type   RecordType
	Offset int
}

// A RecordType is a type of a record.
type RecordType int

const (
	// RecordTypeMapLeaf is the type of a map leaf.
	RecordTypeMapLeaf RecordType = iota
	// RecordTypeMapBranch is the type of a map branch.
	RecordTypeMapBranch
	// RecordTypeListBucket is the type of a list bucket.
	RecordTypeListBucket
	// RecordTypeList is a list, it points to a list bucket.
	RecordTypeList
	// RecordTypeValue is the type of a simple value record
	RecordTypeValue
	// RecordTypeBlock the type of a block record.
	RecordTypeBlock
	// RecordTypeTemplate is the type of a node template record.
	RecordTypeTemplate
	// RecordTypeNode is the type of a node record.
	RecordTypeNode
	// RecordTypeBlobID is the type of a binary object identifier.
	RecordTypeBlobID
)

// ReadFrom reads the content of the segment from a 'reader'. It returns the
// number of bytes read and an optional error.
func (segment *Segment) ReadFrom(reader io.Reader) (int64, error) {
	var buffer bytes.Buffer

	n, err := buffer.ReadFrom(reader)

	if err != nil {
		return n, err
	}

	return n, segment.parseFrom(buffer.Bytes())
}

const (
	v12 = 12
	v13 = 13
)

func (segment *Segment) parseFrom(data []byte) error {
	if len(data) < 4 {
		return fmt.Errorf("invalid data")
	}

	version := data[3]

	if version == v12 {
		return segment.parsev12From(data)
	}
	if version == v13 {
		return segment.parsev13From(data)
	}

	return fmt.Errorf("invalid version %02x", version)
}

func (segment *Segment) parsev12From(data []byte) error {
	const (
		headerSize      = 32
		headerMagic     = "0aK"
		headerVersion   = v12
		headerMagicSize = 3
		referenceSize   = 16
		recordSize      = 9
	)

	const (
		headerMagicOffset          = 0
		headerVersionOffset        = 3
		headerGenerationOffet      = 10
		headerReferenceCountOffset = 14
		headerRecordCountOffset    = 18
	)

	const (
		referenceMsbOffset = 0
		referenceLsbOffset = 8
	)

	const (
		recordNumberOffset = 0
		recordTypeOffset   = 4
		recordOffsetOffset = 5
	)

	if len(data) < headerSize {
		return fmt.Errorf("Segment too small")
	}

	var (
		magic       = string(data[headerMagicOffset : headerMagicOffset+headerMagicSize])
		version     = int(data[headerVersionOffset])
		generation  = int(binary.BigEndian.Uint32(data[headerGenerationOffet:]))
		nreferences = int(binary.BigEndian.Uint32(data[headerReferenceCountOffset:]))
		nrecords    = int(binary.BigEndian.Uint32(data[headerRecordCountOffset:]))
	)

	if magic != headerMagic {
		return fmt.Errorf("Invalid magic")
	}

	if len(data) < headerSize+nreferences*referenceSize+nrecords*recordSize {
		return fmt.Errorf("Invalid size or segment header")
	}

	if version != headerVersion {
		return fmt.Errorf("invalid version %02x", version)
	}

	segment.Generation = generation
	segment.FullGeneration = generation
	segment.Compacted = true
	segment.Version = version
	segment.References = make([]Reference, nreferences)
	segment.Records = make([]Record, nrecords)

	for i := range segment.References {
		referenceData := data[headerSize+i*referenceSize:]
		segment.References[i].Msb = binary.BigEndian.Uint64(referenceData[referenceMsbOffset:])
		segment.References[i].Lsb = binary.BigEndian.Uint64(referenceData[referenceLsbOffset:])
	}

	for i := range segment.Records {
		recordData := data[headerSize+nreferences*referenceSize+i*recordSize:]
		segment.Records[i].Number = int(binary.BigEndian.Uint32(recordData[recordNumberOffset:]))
		segment.Records[i].Type = RecordType(recordData[recordTypeOffset])
		segment.Records[i].Offset = int(binary.BigEndian.Uint32(recordData[recordOffsetOffset:]))
	}

	return nil
}

func (segment *Segment) parsev13From(data []byte) error {
	const (
		headerSize      = 32
		headerMagic     = "0aK"
		headerVersion   = v13
		headerMagicSize = 3
		referenceSize   = 16
		recordSize      = 9
	)

	const (
		headerMagicOffset          = 0
		headerVersionOffset        = 3
		headerFullGenerationOffset = 4
		headerGenerationOffset     = 10
		headerReferenceCountOffset = 14
		headerRecordCountOffset    = 18
	)

	const (
		referenceMsbOffset = 0
		referenceLsbOffset = 8
	)

	const (
		recordNumberOffset = 0
		recordTypeOffset   = 4
		recordOffsetOffset = 5
	)

	if len(data) < headerSize {
		return fmt.Errorf("Segment too small")
	}

	var (
		magic          = string(data[headerMagicOffset : headerMagicOffset+headerMagicSize])
		version        = int(data[headerVersionOffset])
		fullGeneration = int(binary.BigEndian.Uint32(data[headerFullGenerationOffset:]) & 0x7fffffff)
		compacted      = (data[headerFullGenerationOffset] & 0x80) != 0
		generation     = int(binary.BigEndian.Uint32(data[headerGenerationOffset:]))
		nreferences    = int(binary.BigEndian.Uint32(data[headerReferenceCountOffset:]))
		nrecords       = int(binary.BigEndian.Uint32(data[headerRecordCountOffset:]))
	)

	if magic != headerMagic {
		return fmt.Errorf("Invalid magic")
	}

	if len(data) < headerSize+nreferences*referenceSize+nrecords*recordSize {
		return fmt.Errorf("Invalid size or segment header")
	}

	if version != headerVersion {
		return fmt.Errorf("invalid version %02x", version)
	}

	segment.Generation = generation
	segment.FullGeneration = fullGeneration
	segment.Compacted = compacted
	segment.Version = version
	segment.References = make([]Reference, nreferences)
	segment.Records = make([]Record, nrecords)

	for i := range segment.References {
		referenceData := data[headerSize+i*referenceSize:]
		segment.References[i].Msb = binary.BigEndian.Uint64(referenceData[referenceMsbOffset:])
		segment.References[i].Lsb = binary.BigEndian.Uint64(referenceData[referenceLsbOffset:])
	}

	for i := range segment.Records {
		recordData := data[headerSize+nreferences*referenceSize+i*recordSize:]
		segment.Records[i].Number = int(binary.BigEndian.Uint32(recordData[recordNumberOffset:]))
		segment.Records[i].Type = RecordType(recordData[recordTypeOffset])
		segment.Records[i].Offset = int(binary.BigEndian.Uint32(recordData[recordOffsetOffset:]))
	}

	return nil
}
