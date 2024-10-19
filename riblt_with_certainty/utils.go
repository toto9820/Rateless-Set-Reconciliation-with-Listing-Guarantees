package riblt_with_certainty

import "github.com/holiman/uint256"

func xorBytes(a, b [32]byte) [32]byte {
	var result [32]byte
	for i := 0; i < 32; i++ {
		result[i] = a[i] ^ b[i]
	}
	return result
}

func isZeroBytes(a [32]byte) bool {
	for _, b := range a {
		if b != 0 {
			return false
		}
	}
	return true
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
	default:
		panic("Unsupported Symbol type")
	}
}
