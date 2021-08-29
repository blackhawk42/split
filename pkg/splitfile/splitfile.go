package splitfile

import (
	"encoding/hex"
	"errors"
)

// ChecksumAlgo represents a checksum algorithm.
type ChecksumAlgo string

const (
	ChecksumAlgoNone  ChecksumAlgo = "None"
	ChecksumAlgoCRC32 ChecksumAlgo = "CRC32"
)

var (
	ErrUnknownChecksumAlgo = errors.New("unknown checksum algorithm")
)

// Splitfile is the structure of a of a .split file.
type Splitfile struct {
	// Algorithm used for all checksums in the file.
	ChecksumAlgo ChecksumAlgo `json:"checksum_algo"`

	// The whole main file the chunks were derived from. The checksum is of the
	// whole file.
	MainFile *File `json:"main_file"`
	// The chunks that were derived from the main file. Number of chunks can be
	// derived from the number of elements, and their order represents their
	// actual order of concatenation to recreate the main file.
	Chunks []*File `json:"chunks"`
}

// File represents a file, either a main file or an specific chunk.
type File struct {
	// Original filename at time of chunk creation.
	Filename string `json:"filename"`
	// Checksum of the file.
	Checksum Checksum `json:"checksum"`
	// Size of the file.
	Size int64 `json:"size"`
}

// Checksum represents a hash checksum.
//
// Implmentens TextMarshaler and TextUnmarshaler.
type Checksum []byte

func (chk Checksum) MarshalText() (text []byte, err error) {
	text = make([]byte, hex.EncodedLen(len(chk)))
	hex.Encode(text, chk)

	return
}

func (chk *Checksum) UnmarshalText(text []byte) error {
	*chk = make(Checksum, hex.DecodedLen(len(text)))
	_, err := hex.Decode(*chk, text)

	return err
}
