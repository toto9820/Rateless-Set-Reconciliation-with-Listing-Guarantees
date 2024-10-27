package certainsync

import (
	"errors"
)

// Symbol types supported by the IBF
const (
	HashSymbolType   = "hash"
	Uint64SymbolType = "uint64"
	Uint32SymbolType = "uint32"
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
}

// ExtendedInvertibleBloomFilter represents the extended CertainSync
// data structure
type ExtendedInvertibleBloomFilter struct {
	*InvertibleBloomFilter
	HashSeed uint32
}

// NewIBF creates a new InvertibleBloomFilter instance
func NewIBF(size uint64, symbolType string, mapping MappingMethod) *InvertibleBloomFilter {
	cells := make([]Cell, size)

	for i := range cells {
		cells[i] = &IBFCell{}
		cells[i].Init()
	}

	return &InvertibleBloomFilter{
		Cells:         cells,
		Iteration:     0,
		Size:          size,
		SymbolType:    symbolType,
		MappingMethod: mapping,
	}
}

// NewIBF creates a new InvertibleBloomFilter instance
// with cells of type ExtendedIBFCell.
func NewIBFExtended(size uint64, symbolType string, mapping MappingMethod, hashSeed uint32) *ExtendedInvertibleBloomFilter {
	ibf := NewIBF(size, symbolType, mapping)

	cells := make([]Cell, size)

	for i := range cells {
		ibf.Cells[i] = &ExtendedIBFCell{}
		ibf.Cells[i].Init()
	}

	return &ExtendedInvertibleBloomFilter{
		InvertibleBloomFilter: ibf,
		HashSeed:              hashSeed,
	}
}

func (ibf *InvertibleBloomFilter) Copy(ibf2 *InvertibleBloomFilter) {
	ibf.Cells = make([]Cell, len(ibf2.Cells))
	copy(ibf.Cells, ibf2.Cells)
	ibf.Iteration = ibf2.Iteration
	ibf.Size = ibf2.Size
	ibf.SymbolType = ibf2.SymbolType
	ibf.MappingMethod = ibf2.MappingMethod
}

func (ibf *ExtendedInvertibleBloomFilter) Copy(ibf2 *ExtendedInvertibleBloomFilter) {
	ibf.InvertibleBloomFilter = ibf2.InvertibleBloomFilter
	ibf.HashSeed = ibf2.HashSeed
}

func (ibf *InvertibleBloomFilter) AddSymbols(symbols []Symbol) {
	additionalCellsCount := ibf.MappingMethod.GetAdditionalCellsCount(ibf.SymbolType, ibf.Iteration)

	if ibf.Size+additionalCellsCount > uint64(len(ibf.Cells)) {
		newCapacity := len(ibf.Cells) * 2

		newCells := make([]Cell, newCapacity)

		for i := range newCells {
			newCells[i].Init()
		}

		copy(newCells, ibf.Cells)

		ibf.Cells = newCells
	}

	for _, s := range symbols {
		j := ibf.Size + ibf.MappingMethod.MapSymbol(s, ibf.Iteration)
		ibf.Cells[j].Insert(s)
	}

	ibf.Size += additionalCellsCount
	ibf.Iteration++
}

func (ibf *InvertibleBloomFilter) Subtract(ibf2 *InvertibleBloomFilter) *InvertibleBloomFilter {
	difference := NewIBF(ibf.Size, ibf.SymbolType, ibf.MappingMethod)
	difference.Copy(ibf)

	for j := uint64(0); j < ibf.Size; j++ {
		difference.Cells[j].Subtract(ibf2.Cells[j])
	}

	return difference
}

func (ibf *ExtendedInvertibleBloomFilter) Subtract(ibf2 *ExtendedInvertibleBloomFilter) *ExtendedInvertibleBloomFilter {
	difference := NewIBFExtended(ibf.Size, ibf.SymbolType, ibf.MappingMethod, ibf.HashSeed)
	difference.Copy(ibf)

	for j := uint64(0); j < ibf.Size; j++ {
		difference.Cells[j].Subtract(ibf2.Cells[j])
	}

	return difference
}

func (ibf *InvertibleBloomFilter) Decode() (symmetricDiff []Symbol, ok bool) {
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

		xorSum := ibf.Cells[j].GetXorSum()

		symmetricDiff = append(symmetricDiff, xorSum)

		offset := uint64(0)
		for i := uint64(0); i < ibf.Iteration; i++ {
			cellIdx := offset + ibf.MappingMethod.MapSymbol(xorSum, i)
			ibf.Cells[cellIdx].Subtract(ibf.Cells[j])

			offset += ibf.MappingMethod.GetAdditionalCellsCount(ibf.SymbolType, i)
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

func (ibf *ExtendedInvertibleBloomFilter) IsFullyEmpty() bool {
	for j := uint64(0); j < ibf.Size; j++ {
		if !ibf.Cells[j].IsZeroExtended() {
			return false
		}
	}
	return true
}
