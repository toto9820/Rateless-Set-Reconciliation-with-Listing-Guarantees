package certainsync

import (
	"time"

	"github.com/holiman/uint256"
	"golang.org/x/exp/rand"
)

func GenerateRandomSeed() uint32 {
	// Seed the random number generator
	rand.Seed(uint64(time.Now().UnixNano()))
	// Generate a random uint32 seed
	return rand.Uint32()
}

func XorBytes(a, b [32]byte) [32]byte {
	var result [32]byte
	for i := 0; i < 32; i++ {
		result[i] = a[i] ^ b[i]
	}
	return result
}

func IsZeroBytes(a [32]byte) bool {
	for _, b := range a {
		if b != 0 {
			return false
		}
	}
	return true
}

func SubstractSymbolUint64(s Symbol, m uint64) uint64 {
	switch sym := s.(type) {
	case HashSymbol:
		var bigIntHash *uint256.Int = new(uint256.Int)
		bigIntHash = bigIntHash.SetBytes(sym[:])
		bigIntMod := uint256.NewInt(m)
		bigIntHash.Sub(bigIntHash, bigIntMod)
		return bigIntHash.Uint64()
	case Uint64Symbol:
		return uint64(sym) - m
	case Uint32Symbol:
		return uint64(uint32(sym)) - m
	default:
		panic("Unsupported Symbol type")
	}
}

func DivideSymbolUint64(s Symbol, m uint64) uint64 {
	switch sym := s.(type) {
	case HashSymbol:
		var bigIntHash *uint256.Int = new(uint256.Int)
		bigIntHash = bigIntHash.SetBytes(sym[:])
		bigIntMod := uint256.NewInt(m)
		bigIntHash.Div(bigIntHash, bigIntMod)
		return bigIntHash.Uint64()
	case Uint64Symbol:
		return uint64(sym) / m
	case Uint32Symbol:
		return uint64(uint32(sym)) / m
	default:
		panic("Unsupported Symbol type")
	}
}

func ModSymbolUint64(s Symbol, m uint64) uint64 {
	switch sym := s.(type) {
	case HashSymbol:
		var bigIntHash *uint256.Int = new(uint256.Int)
		bigIntHash = bigIntHash.SetBytes(sym[:])
		bigIntMod := uint256.NewInt(m)
		bigIntHash.Mod(bigIntHash, bigIntMod)
		return bigIntHash.Uint64()
	case Uint64Symbol:
		return uint64(sym) % m
	case Uint32Symbol:
		return uint64(uint32(sym)) % m
	default:
		panic("Unsupported Symbol type")
	}
}
