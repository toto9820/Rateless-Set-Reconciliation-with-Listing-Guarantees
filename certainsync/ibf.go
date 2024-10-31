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
	Cells         []IBFCell
	Iteration     uint64
	Size          uint64
	SymbolType    string
	MappingMethod MappingMethod
}

// NewIBF creates a new InvertibleBloomFilter instance
func NewIBF(size uint64, symbolType string, mapping MappingMethod) *InvertibleBloomFilter {
	cells := make([]IBFCell, size)

	for i := range cells {
		cells[i].Init(symbolType)
	}

	return &InvertibleBloomFilter{
		Cells:         cells,
		Iteration:     0,
		Size:          0,
		SymbolType:    symbolType,
		MappingMethod: mapping,
	}
}

func (ibf *InvertibleBloomFilter) Copy(ibf2 *InvertibleBloomFilter) {
	ibf.Cells = make([]IBFCell, len(ibf2.Cells))

	for i := range ibf.Cells {
		ibf.Cells[i].Init(ibf.SymbolType)
		ibf.Cells[i].DeepCopy(ibf2.Cells[i])
	}

	ibf.Iteration = ibf2.Iteration
	ibf.Size = ibf2.Size
	ibf.SymbolType = ibf2.SymbolType
	ibf.MappingMethod = ibf2.MappingMethod
}

func (ibf *InvertibleBloomFilter) AddSymbols(symbols []Symbol) {
	ibf.Iteration++
	additionalCellsCount := ibf.MappingMethod.GetAdditionalCellsCount(ibf.SymbolType, ibf.Iteration)

	if ibf.Size+additionalCellsCount > uint64(len(ibf.Cells)) {
		newCapacity := ibf.Size + additionalCellsCount

		newCells := make([]IBFCell, newCapacity)

		for i := range newCells {
			newCells[i].Init(ibf.SymbolType)
		}

		copy(newCells, ibf.Cells)

		ibf.Cells = newCells
	}

	// Add symbols to cells
	for _, s := range symbols {
		j := ibf.Size + ibf.MappingMethod.MapSymbol(s, ibf.Iteration)
		ibf.Cells[j].Insert(s)
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

// GetTransmittedBitsSize returns the bit size of the actively transmitted cells.
// This reflects only the cells that have been added to the IBF and excludes unutilized cells.
// Other field of the IBF are either private to him like iteration, or
// agreed ahead like hash seed, mapping method with others.
func (ibf *InvertibleBloomFilter) GetTransmittedBitsSize() uint64 {
	cellSize := ibf.Cells[0].SizeInBits()
	return ibf.Size * cellSize
}
