package riblt_with_certainty

import "github.com/ethereum/go-ethereum/common"

// SymbolHash is an interface that can be either a
// common.Hash, uint64 or uint32.
type Hash interface {
	Xor(Hash) Hash
	IsZero() bool
	Equal(Hash) bool
}

// CommonHash wraps common.Hash to implement
// the Hash interface
type CommonHash common.Hash

func (h CommonHash) Xor(other Hash) Hash {
	o := other.(CommonHash)
	var result CommonHash
	for i := 0; i < common.HashLength; i++ {
		result[i] = h[i] ^ o[i]
	}
	return result
}

func (h CommonHash) IsZero() bool {
	return h == CommonHash{}
}

func (h CommonHash) Equal(other Hash) bool {
	o := other.(CommonHash)
	return h == o
}

// Uint64Hash wraps uint64 to implement
// the Hash interface
type Uint64Hash uint64

func (h Uint64Hash) Xor(other Hash) Hash {
	o := other.(Uint64Hash)
	var result Uint64Hash
	result = h ^ o
	return result
}

func (h Uint64Hash) IsZero() bool {
	return h == 0
}

func (h Uint64Hash) Equal(other Hash) bool {
	o := other.(Uint64Hash)
	return h == o
}

// Uint32Hash wraps uint32 to implement
// the Hash interface
type Uint32Hash uint32

func (h Uint32Hash) Xor(other Hash) Hash {
	o := other.(Uint32Hash)
	var result Uint32Hash
	result = h ^ o
	return result
}

func (h Uint32Hash) IsZero() bool {
	return h == 0
}

func (h Uint32Hash) Equal(other Hash) bool {
	o := other.(Uint32Hash)
	return h == o
}
