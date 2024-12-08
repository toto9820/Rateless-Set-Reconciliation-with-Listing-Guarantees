package certainsync

import (
	"golang.org/x/exp/rand"
)

// GenerateRandomSalt generates a random 32-bit integer
// seed based on a given round number.
func GenerateRandomSalt(roundNumber uint64) uint32 {
	// Seed the random number generator
	rand.Seed(roundNumber)
	// Generate a random uint32 seed
	return rand.Uint32()
}

// XorBytes performs a byte-wise XOR operation
// on two 32-byte arrays.
func XorBytes(a, b [32]byte) [32]byte {
	var result [32]byte
	for i := 0; i < 32; i++ {
		result[i] = a[i] ^ b[i]
	}
	return result
}

// IsZeroBytes checks whether a 32-byte array is
// filled entirely with zeros.
func IsZeroBytes(a [32]byte) bool {
	for _, b := range a {
		if b != 0 {
			return false
		}
	}
	return true
}
