package riblt_with_certainty

import (
	"crypto/sha256"

	"github.com/ethereum/go-ethereum/common"
	"github.com/holiman/uint256"
)

// Symbol is an interface that can be either a
// common.Hash or a uint64
type Symbol interface {
	Xor(Symbol) Symbol
	Hash() [32]byte
	IsZero() bool
	Equal(Symbol) bool
}

// HashSymbol wraps common.Hash to implement the Symbol interface
type HashSymbol common.Hash

func (h HashSymbol) Xor(other Symbol) Symbol {
	o := other.(HashSymbol)
	var result HashSymbol
	for i := 0; i < common.HashLength; i++ {
		result[i] = h[i] ^ o[i]
	}
	return result
}

func (h HashSymbol) Hash() [32]byte {
	return sha256.Sum256(h[:])
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

func (u Uint64Symbol) Xor(other Symbol) Symbol {
	o := other.(Uint64Symbol)
	return Uint64Symbol(uint64(u) ^ uint64(o))
}

func (u Uint64Symbol) Hash() [32]byte {
	buf := uint256.NewInt(uint64(u)).Bytes()
	return sha256.Sum256(buf)
}

func (u Uint64Symbol) IsZero() bool {
	return u == 0
}

func (u Uint64Symbol) Equal(other Symbol) bool {
	o := other.(Uint64Symbol)
	return u == o
}
