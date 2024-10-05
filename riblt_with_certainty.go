package riblt_with_certainty

import (
	"encoding/binary"

	"github.com/cespare/xxhash"
	"github.com/kavehmz/prime"
)

// TODO - add documentation at the end.

// Package-level variable to hold primes.
var primes []uint64

// This function runs automatically when the
// package is initialized.
func init() {
	primes = prime.Primes(1000000)
}

type IBFCell struct {
	Count   int64
	XorSum  uint64
	HashSum uint64
}

type InvertibleBloomFilter struct {
	Cells     []IBFCell
	Iteration uint64
	Size      uint64
}

func NewIBF(size uint64) *InvertibleBloomFilter {
	return &InvertibleBloomFilter{
		Cells:     make([]IBFCell, size),
		Iteration: 0,
		Size:      0,
	}
}

// Copy method to copy the contents of one
// InvertibleBloomFilter to another.
func (ibf *InvertibleBloomFilter) Copy(ibf2 *InvertibleBloomFilter) {
	// Make a deep copy of the Cells slice
	ibf.Cells = make([]IBFCell, len(ibf2.Cells))
	copy(ibf.Cells, ibf2.Cells)

	// Copy primitive types directly
	ibf.Iteration = ibf2.Iteration
	ibf.Size = ibf2.Size
}

func Hash(s uint64) uint64 {
	buf := [8]byte{}
	binary.LittleEndian.PutUint64(buf[:], s)
	return xxhash.Sum64(buf[:])
}

func (c *IBFCell) Insert(s uint64) {
	c.Count++
	c.XorSum ^= s
	c.HashSum ^= Hash(s)
}

func (c *IBFCell) Subtract(c2 *IBFCell) {
	c.Count -= c2.Count
	c.XorSum ^= c2.XorSum
	c.HashSum ^= c2.HashSum
}

func (c *IBFCell) IsPure() bool {
	return (c.Count == 1 || c.Count == -1) && c.HashSum == Hash(c.XorSum)
}

func (c *IBFCell) IsZero() bool {
	return c.Count == 0 && c.XorSum == 0 && c.HashSum == 0
}

// For now - just egh implementing to check if really faster.
func (ibf *InvertibleBloomFilter) AddSymbols(symbols []uint64) {
	// Get the next prime based on the current iteration
	curPrime := primes[ibf.Iteration]

	// Ensure enough capacity for new cells
	if ibf.Size+curPrime > uint64(len(ibf.Cells)) {
		// Doubling the capacity instead of a fixed increase
		newCapacity := len(ibf.Cells) * 2

		if newCapacity < int(ibf.Size+curPrime) {
			newCapacity = int(ibf.Size + curPrime)
		}

		newCells := make([]IBFCell, newCapacity)
		copy(newCells, ibf.Cells)
		ibf.Cells = newCells
	}

	// Add symbols to the IBF cells
	for _, s := range symbols {
		j := ibf.Size + s%curPrime
		ibf.Cells[j].Insert(s)
	}

	// Update the size of the filter
	ibf.Size += curPrime

	// Increment the iteration count
	ibf.Iteration++
}

// Subtract computes the difference between 2 Invertible Bloom Filters
func (ibf *InvertibleBloomFilter) Subtract(ibf2 *InvertibleBloomFilter) *InvertibleBloomFilter {
	// Create a new InvertibleBloomFilter to hold the result
	difference := NewIBF(ibf.Size)
	difference.Copy(ibf)

	// Subtract the Cells of the second IBF from the first
	for j := uint64(0); j < ibf.Size; j++ {
		difference.Cells[j].Subtract(&ibf2.Cells[j])
	}

	return difference
}

func (ibf *InvertibleBloomFilter) Decode() (aWithoutB []uint64, bWithoutA []uint64, ok bool) {
	pureList := make([]uint64, 0)

	for {
		n := len(pureList) - 1

		if n == -1 {
			for j := uint64(0); j < ibf.Size; j++ {
				if ibf.Cells[j].IsPure() {
					pureList = append(pureList, j)
				}
			}
			if len(pureList) == 0 {
				break
			}
			continue
		}

		j := pureList[n]
		pureList = pureList[:n]

		if !ibf.Cells[j].IsPure() {
			continue
		}

		count := ibf.Cells[j].Count
		xorSum := ibf.Cells[j].XorSum

		if count > 0 {
			aWithoutB = append(aWithoutB, xorSum)
		} else {
			bWithoutA = append(bWithoutA, xorSum)
		}

		primesSum := uint64(0)

		for _, prime := range primes[:ibf.Iteration] {
			cellIdx := primesSum + xorSum%prime
			ibf.Cells[cellIdx].Count -= count
			ibf.Cells[cellIdx].XorSum ^= xorSum
			ibf.Cells[cellIdx].HashSum ^= Hash(xorSum)

			primesSum += prime
		}
	}

	for j := uint64(0); j < ibf.Size; j++ {
		if !ibf.Cells[j].IsZero() {
			ok = false
			return
		}
	}

	ok = true
	return
}
