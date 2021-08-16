package splitfile

import (
	"hash"
	"hash/crc32"
)

// HasherFunc is a function that returns a new implementation of hash.Hash.
//
// There is one implementation for every supported checksum algorithm.
// ChecksumAlgoNone returns a nil HasherFunc and no error.
type HasherFunc func() hash.Hash

func GetHasher(checksumAlgo ChecksumAlgo) (hasherFunc HasherFunc, err error) {
	switch checksumAlgo {
	case ChecksumAlgoCRC32:
		hasherFunc = func() hash.Hash {
			hasher := crc32.NewIEEE()
			return hasher
		}

	case ChecksumAlgoNone:
		hasherFunc = nil
	default:
		err = ErrUnknownChecksumAlgo
	}

	return
}
