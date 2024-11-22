package certainsync

import (
	"golang.org/x/exp/rand"
)

func GenerateRandomSalt(roundNumber uint64) uint32 {
	// Seed the random number generator
	rand.Seed(roundNumber)
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
