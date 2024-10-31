package certainsync

import "github.com/kavehmz/prime"

var primes []uint64

func init() {
	primes = prime.Primes(1000000)
}
