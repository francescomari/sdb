package segment

import (
	"bytes"
	"encoding/binary"
	"fmt"
	"io"
)

const (
	headerSize      = 32
	headerMagic     = "0aK"
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

// A Segment is a container for records.
type Segment struct {
	Version    int
	Generation int
	References []Reference
	Records    []Record
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

func (segment *Segment) parseFrom(data []byte) error {
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

	segment.Generation = generation
	segment.Version = version
	segment.References = make([]Reference, nreferences)
	segment.Records = make([]Record, nrecords)

	for i := range segment.References {
		segment.References[i].parseFrom(data[headerSize+i*referenceSize:])
	}

	for i := range segment.Records {
		segment.Records[i].parseFrom(data[headerSize+nreferences*referenceSize+i*recordSize:])
	}

	return nil
}

func (reference *Reference) parseFrom(data []byte) {
	reference.Msb = binary.BigEndian.Uint64(data[referenceMsbOffset:])
	reference.Lsb = binary.BigEndian.Uint64(data[referenceLsbOffset:])
}

func (record *Record) parseFrom(data []byte) {
	record.Number = int(binary.BigEndian.Uint32(data[recordNumberOffset:]))
	record.Type = RecordType(data[recordTypeOffset])
	record.Offset = int(binary.BigEndian.Uint32(data[recordOffsetOffset:]))
}
