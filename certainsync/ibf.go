package certainsync

import (
	"errors"
)

// Common errors
var (
	ErrInvalidSymbolType = errors.New("invalid symbol type")
	ErrSizeMismatch      = errors.New("IBF size mismatch")
	ErrNilIBF            = errors.New("nil IBF reference")
)

// InvertibleBloomFilter represents the basic CertainSync
// data structure
type InvertibleBloomFilter struct {
	Cells         []Cell
	Iteration     uint64
	Size          uint64
	SymbolType    string
	MappingMethod MappingMethod
	HashSeed      uint32
}

// NewIBF creates a new InvertibleBloomFilter instance
func NewIBF(size uint64, symbolType string, mapping MappingMethod, seed ...uint32) *InvertibleBloomFilter {
	var hashSeed uint32 = 0

	if len(seed) > 0 {
		hashSeed = seed[0]
	}

	cells := make([]Cell, size)

	for i := range cells {
		cells[i] = &IBFCell{}
		cells[i].Init(symbolType)
	}

	return &InvertibleBloomFilter{
		Cells:         cells,
		Iteration:     0,
		Size:          0,
		SymbolType:    symbolType,
		MappingMethod: mapping,
		HashSeed:      hashSeed,
	}
}

// NewIBF creates a new InvertibleBloomFilter instance
// with cells of type ExtendedIBFCell.
func NewIBFExtended(size uint64, symbolType string, mapping MappingMethod, hashSeed uint32) *InvertibleBloomFilter {
	ibf := NewIBF(size, symbolType, mapping, hashSeed)

	cells := make([]Cell, size)

	for i := range cells {
		ibf.Cells[i] = &ExtendedIBFCell{}
		ibf.Cells[i].Init(symbolType, hashSeed)
	}

	return ibf
}

func (ibf *InvertibleBloomFilter) Copy(ibf2 *InvertibleBloomFilter) {
	ibf.Cells = make([]Cell, len(ibf2.Cells))

	for i := range ibf.Cells {
		if ibf2.Cells[0].IsExtended() {
			ibf.Cells[i] = &ExtendedIBFCell{}
		} else {
			ibf.Cells[i] = &IBFCell{}
		}

		ibf.Cells[i].Init(ibf.SymbolType, ibf.HashSeed)
		ibf.Cells[i].DeepCopy(ibf2.Cells[i])
	}

	ibf.Iteration = ibf2.Iteration
	ibf.Size = ibf2.Size
	ibf.SymbolType = ibf2.SymbolType
	ibf.MappingMethod = ibf2.MappingMethod
	ibf.HashSeed = ibf2.HashSeed
}

func (ibf *InvertibleBloomFilter) AddSymbols(symbols []Symbol, mapping ...map[Symbol]Symbol) {
	ibf.Iteration++
	additionalCellsCount := ibf.MappingMethod.GetAdditionalCellsCount(ibf.SymbolType, ibf.Iteration)

	if ibf.Size+additionalCellsCount > uint64(len(ibf.Cells)) {
		newCapacity := ibf.Size + additionalCellsCount

		newCells := make([]Cell, newCapacity)

		for i := range newCells {
			if ibf.Cells[0].IsExtended() {
				newCells[i] = &ExtendedIBFCell{}
			} else {
				newCells[i] = &IBFCell{}
			}

			newCells[i].Init(ibf.SymbolType, ibf.HashSeed)
		}

		copy(newCells, ibf.Cells)

		ibf.Cells = newCells
	}

	// Check if mapping was provided
	var symbolMap map[Symbol]Symbol
	if len(mapping) > 0 {
		symbolMap = mapping[0]
	}

	// Add symbols to cells
	for _, s := range symbols {
		j := ibf.Size + ibf.MappingMethod.MapSymbol(s, ibf.Iteration)

		if symbolMap != nil {
			// If we have a mapping, store both the transformed symbol
			// and its original
			originalSymbol := symbolMap[s]
			ibf.Cells[j].Insert(s, originalSymbol)
		} else {
			// Original behavior without mapping
			ibf.Cells[j].Insert(s)
		}
	}

	ibf.Size += additionalCellsCount
}

func (ibf *InvertibleBloomFilter) Subtract(ibf2 *InvertibleBloomFilter) *InvertibleBloomFilter {
	difference := NewIBF(ibf.Size, ibf.SymbolType, ibf.MappingMethod)
	difference.Copy(ibf)

	for j := uint64(0); j < ibf.Size; j++ {
		difference.Cells[j].Subtract(ibf2.Cells[j])
	}

	return difference
}

func (ibf *InvertibleBloomFilter) Decode() (symmetricDiff []Symbol, ok bool) {
	pureList := make([]uint64, 0)
	// Detecting duplicates
	seenSymbols := make(map[Symbol]bool)

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

		xorSum := ibf.Cells[j].GetXorSum()

		// Check for duplicates
		if seenSymbols[xorSum] {
			continue
		}

		symmetricDiff = append(symmetricDiff, xorSum)
		seenSymbols[xorSum] = true

		offset := uint64(0)
		for i := uint64(1); i <= ibf.Iteration; i++ {
			cellIdx := offset + ibf.MappingMethod.MapSymbol(xorSum, i)

			// Empty the pure cell at index j at the end
			if cellIdx != j {
				ibf.Cells[cellIdx].Subtract(ibf.Cells[j])
			}

			offset += ibf.MappingMethod.GetAdditionalCellsCount(ibf.SymbolType, i)
		}

		ibf.Cells[j].Subtract(ibf.Cells[j])
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

func (ibf *InvertibleBloomFilter) IsFullyEmpty() bool {
	for j := uint64(0); j < ibf.Size; j++ {
		if !ibf.Cells[j].IsZeroExtended() {
			return false
		}
	}
	return true
}
