package certainsync

import (
	"crypto/sha256"
	"encoding/binary"

	"github.com/cespare/xxhash/v2"
	"github.com/ethereum/go-ethereum/common"
	"github.com/spaolacci/murmur3"
)

// Symbol is an interface that can be either a
// common.Hash or a uint64 or uint32
type Symbol interface {
	NewSymbol() Symbol
	Xor(Symbol) Symbol
	Hash(seed ...uint32) Hash
	IsZero() bool
	Equal(Symbol) bool
}

// HashSymbol wraps common.Hash to implement the Symbol interface
type HashSymbol common.Hash

func (h HashSymbol) NewSymbol() Symbol {
	return HashSymbol{}
}

func (h HashSymbol) Xor(other Symbol) Symbol {
	o := other.(HashSymbol)
	var result HashSymbol
	for i := 0; i < common.HashLength; i++ {
		result[i] = h[i] ^ o[i]
	}
	return result
}

// Change the Hash method to return a byte array
func (h HashSymbol) Hash(seed ...uint32) Hash {
	return CommonHash(sha256.Sum256(h[:]))
}

func (h HashSymbol) IsZero() bool {
	return h == HashSymbol{}
}

func (h HashSymbol) Equal(other Symbol) bool {
	o := other.(HashSymbol)
	return h == o
}

// Uint64Symbol wraps uint64 to implement the Symbol interface
type Uint64Symbol uint64

func (h Uint64Symbol) NewSymbol() Symbol {
	return Uint64Symbol(0)
}

func (u Uint64Symbol) Xor(other Symbol) Symbol {
	o := other.(Uint64Symbol)
	return Uint64Symbol(uint64(u) ^ uint64(o))
}

func (u Uint64Symbol) Hash(seed ...uint32) Hash {
	buf := make([]byte, 8)
	binary.BigEndian.PutUint64(buf, uint64(u))
	hash := xxhash.Sum64(buf)
	return Uint64Hash(hash)
}

func (u Uint64Symbol) IsZero() bool {
	return u == 0
}

func (u Uint64Symbol) Equal(other Symbol) bool {
	o := other.(Uint64Symbol)
	return u == o
}

// Uint64Symbol wraps uint64 to implement the Symbol interface
type Uint32Symbol uint32

func (h Uint32Symbol) NewSymbol() Symbol {
	return Uint32Symbol(0)
}

func (u Uint32Symbol) Xor(other Symbol) Symbol {
	o := other.(Uint32Symbol)
	return Uint64Symbol(uint32(u) ^ uint32(o))
}

func (u Uint32Symbol) Hash(seed ...uint32) Hash {
	buf := make([]byte, 4)
	binary.BigEndian.PutUint32(buf, uint32(u))

	var hashSeed uint32
	if len(seed) > 0 {
		hashSeed = seed[0]
	} else {
		hashSeed = 0 // Default seed if none is provided
	}

	hash := murmur3.Sum32WithSeed(buf, hashSeed)
	return Uint32Hash(hash)
}

func (u Uint32Symbol) IsZero() bool {
	return u == 0
}

func (u Uint32Symbol) Equal(other Symbol) bool {
	o := other.(Uint32Symbol)
	return u == o
}
