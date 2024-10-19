package riblt_with_certainty

import (
	"crypto/sha256"

	"github.com/ethereum/go-ethereum/common"
	"github.com/holiman/uint256"
	"github.com/kavehmz/prime"
)

var primes []uint64

func init() {
	primes = prime.Primes(1000000)
}

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

type IBFCell struct {
	Count   int64
	XorSum  Symbol
	HashSum common.Hash
}

type InvertibleBloomFilter struct {
	Cells      []IBFCell
	Iteration  uint64
	Size       uint64
	SymbolType string // "hash" or "uint64"
}

func NewIBF(size uint64, symbolType string) *InvertibleBloomFilter {
	var zeroSymbol Symbol
	switch symbolType {
	case "hash":
		zeroSymbol = HashSymbol{}
	case "uint64":
		zeroSymbol = Uint64Symbol(0)
	default:
		panic("Invalid symbol type")
	}

	cells := make([]IBFCell, size)

	for i := range cells {
		cells[i].XorSum = zeroSymbol
	}

	return &InvertibleBloomFilter{
		Cells:      cells,
		Iteration:  0,
		Size:       0,
		SymbolType: symbolType,
	}
}

func (ibf *InvertibleBloomFilter) Copy(ibf2 *InvertibleBloomFilter) {
	ibf.Cells = make([]IBFCell, len(ibf2.Cells))
	copy(ibf.Cells, ibf2.Cells)
	ibf.Iteration = ibf2.Iteration
	ibf.Size = ibf2.Size
	ibf.SymbolType = ibf2.SymbolType
}

func (c *IBFCell) Insert(s Symbol) {
	c.Count++
	c.XorSum = c.XorSum.Xor(s)
	c.HashSum = xorBytes(c.HashSum, s.Hash())
}

func (c *IBFCell) Subtract(c2 *IBFCell) {
	c.Count -= c2.Count
	c.XorSum = c.XorSum.Xor(c2.XorSum)
	c.HashSum = xorBytes(c.HashSum, c2.HashSum)
}

func (c *IBFCell) IsPure() bool {
	return (c.Count == 1 || c.Count == -1) && c.HashSum == c.XorSum.Hash()
}

func (c *IBFCell) IsZero() bool {
	return c.Count == 0 && c.XorSum.IsZero() && isZeroBytes(c.HashSum)
}

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

func (ibf *InvertibleBloomFilter) AddSymbols(symbols []Symbol) {
	curPrime := primes[ibf.Iteration]

	if ibf.Size+curPrime > uint64(len(ibf.Cells)) {
		newCapacity := len(ibf.Cells) * 2

		newCells := make([]IBFCell, newCapacity)

		var zeroSymbol Symbol

		switch ibf.SymbolType {
		case "hash":
			zeroSymbol = HashSymbol{}
		case "uint64":
			zeroSymbol = Uint64Symbol(0)
		default:
			panic("Invalid symbol type")
		}

		for i := range newCells {
			newCells[i].XorSum = zeroSymbol
		}

		copy(newCells, ibf.Cells)

		ibf.Cells = newCells
	}

	for _, s := range symbols {
		j := ibf.Size + ModSymbolUint64(s, curPrime)
		ibf.Cells[j].Insert(s)
	}

	ibf.Size += curPrime
	ibf.Iteration++
}

func ModSymbolUint64(s Symbol, m uint64) uint64 {
	switch sym := s.(type) {
	case HashSymbol:
		var bigIntHash *uint256.Int
		bigIntHash.SetBytes(sym[:])
		bigIntMod := uint256.NewInt(m)
		bigIntHash.Mod(bigIntHash, bigIntMod)
		return bigIntHash.Uint64()
	case Uint64Symbol:
		return uint64(sym) % m
	default:
		panic("Unsupported Symbol type")
	}
}

func (ibf *InvertibleBloomFilter) Subtract(ibf2 *InvertibleBloomFilter) *InvertibleBloomFilter {
	difference := NewIBF(ibf.Size, ibf.SymbolType)
	difference.Copy(ibf)

	for j := uint64(0); j < ibf.Size; j++ {
		difference.Cells[j].Subtract(&ibf2.Cells[j])
	}

	return difference
}

func (ibf *InvertibleBloomFilter) Decode() (aWithoutB []Symbol, bWithoutA []Symbol, ok bool) {
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
			cellIdx := primesSum + ModSymbolUint64(xorSum, prime)
			ibf.Cells[cellIdx].Count -= count
			ibf.Cells[cellIdx].XorSum = ibf.Cells[cellIdx].XorSum.Xor(xorSum)
			ibf.Cells[cellIdx].HashSum = xorBytes(ibf.Cells[cellIdx].HashSum, xorSum.Hash())

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
