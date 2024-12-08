package certainsync

import "github.com/kavehmz/prime"

// primes holds a list of prime numbers generated up to a specified limit.
// It is initialized once during package initialization.
var primes []uint64

// init initializes the primes slice with prime numbers up to 1,000,000.
// This function is automatically called when the package is imported.
// The primes are generated using the `Primes` function from the `prime` package.
func init() {
	primes = prime.Primes(1000000)
}
