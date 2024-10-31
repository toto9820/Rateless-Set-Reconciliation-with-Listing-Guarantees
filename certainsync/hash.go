package certainsync

import "github.com/ethereum/go-ethereum/common"

// SymbolHash is an interface that can be either a
// common.Hash, uint64 or uint32.
type Hash interface {
	NewHash() Hash
	Xor(Hash) Hash
	IsZero() bool
	Equal(Hash) bool
	DeepCopy() Hash
	SizeInBits() uint64
}

// CommonHash wraps common.Hash to implement
// the Hash interface
type CommonHash common.Hash

func (h CommonHash) NewHash() Hash {
	return CommonHash{}
}

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

// DeepCopy creates a deep copy of CommonHash
func (h CommonHash) DeepCopy() Hash {
	if h.IsZero() {
		return CommonHash{}
	}

	newHash := make([]byte, common.HashLength)
	copy(newHash, h[:])
	return CommonHash(newHash)
}

// Implement SizeInBits for CommonHash
func (h CommonHash) SizeInBits() uint64 {
	return 256 // common.Hash is 32 bytes or 256 bits
}

// Uint64Hash wraps uint64 to implement
// the Hash interface
type Uint64Hash uint64

func (h Uint64Hash) NewHash() Hash {
	return Uint64Hash(0)
}

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

// DeepCopy creates a deep copy of Uint64Hash
func (h Uint64Hash) DeepCopy() Hash {
	return Uint64Hash(uint64(h))
}

// Implement SizeInBits for Uint64Hash
func (h Uint64Hash) SizeInBits() uint64 {
	return 64 // Uint64Hash is 64 bits
}

// Uint32Hash wraps uint32 to implement
// the Hash interface
type Uint32Hash uint32

func (h Uint32Hash) NewHash() Hash {
	return Uint32Hash(0)
}

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

// DeepCopy creates a deep copy of Uint32Hash
func (h Uint32Hash) DeepCopy() Hash {
	return Uint32Hash(uint32(h))
}

// Implement SizeInBits for Uint32Hash
func (h Uint32Hash) SizeInBits() uint64 {
	return 32 // Uint32Hash is 32 bits
}
