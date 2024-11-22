package certainsync

import (
	"crypto/sha256"

	"github.com/cespare/xxhash/v2"
	"github.com/holiman/uint256"
	"github.com/spaolacci/murmur3"
)

// CellHasher is used to compute the hash of the provided data.
type CellHasher interface {
	Hash(data []byte) *uint256.Int
}

// Murmur3Hash implements CellHasher using Murmur3
type Murmur3Hash struct{}

func (h Murmur3Hash) Hash(data []byte) *uint256.Int {
	hash := uint64(murmur3.Sum32(data))
	return uint256.NewInt(hash)
}

// XXHash64Hash implements CellHasher using XXHash64
type XXHash64Hash struct{}

func (h XXHash64Hash) Hash(data []byte) *uint256.Int {
	hash := xxhash.Sum64(data)
	return uint256.NewInt(hash)
}

// Sha256Hash implements CellHasher using SHA-256
type Sha256Hash struct{}

func (h Sha256Hash) Hash(data []byte) *uint256.Int {
	hash := sha256.Sum256(data)
	return new(uint256.Int).SetBytes(hash[:])
}
