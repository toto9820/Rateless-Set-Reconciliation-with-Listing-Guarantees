package certainsync

import (
	"errors"

	"github.com/holiman/uint256"
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
	Cells         []IBFCell     // Array of IBLT cells
	UniverseSize  *uint256.Int  // The size of the universe
	Iteration     uint64        // Current iteration (number of times symbols are added)
	Size          uint64        // Number of cells in the filter
	MappingMethod MappingMethod // Method type used for mapping symbols to cells
}

// NewIBF creates a new InvertibleBloomFilter instance.
func NewIBF(universeSize *uint256.Int, mapping MappingMethod) *InvertibleBloomFilter {
	return &InvertibleBloomFilter{
		Cells:         nil,
		UniverseSize:  universeSize.Clone(),
		Iteration:     0,
		Size:          0,
		MappingMethod: mapping,
	}
}

// Copy copies the contents of another IBF into the current IBF.
func (ibf *InvertibleBloomFilter) Copy(ibf2 *InvertibleBloomFilter) {
	ibf.Cells = make([]IBFCell, len(ibf2.Cells))

	for i := range ibf.Cells {
		ibf.Cells[i] = NewIBFCell(ibf2.UniverseSize)
		ibf.Cells[i] = ibf2.Cells[i].Clone()
	}

	ibf.UniverseSize = ibf2.UniverseSize.Clone()
	ibf.Iteration = ibf2.Iteration
	ibf.Size = ibf2.Size
	ibf.MappingMethod = ibf2.MappingMethod
}

// AddSymbols adds a list of symbols to the IBF.
func (ibf *InvertibleBloomFilter) AddSymbols(symbols []*uint256.Int) {
	ibf.Iteration++
	additionalCellsCount := ibf.MappingMethod.GetAdditionalCellsCount(ibf.Iteration)

	// Expand the cell array if needed
	if ibf.Size+additionalCellsCount > uint64(len(ibf.Cells)) {
		newCapacity := ibf.Size + additionalCellsCount

		newCells := make([]IBFCell, newCapacity)

		for i := range newCells {
			newCells[i] = NewIBFCell(ibf.UniverseSize)
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

// Subtract subtracts another IBF from the current one.
func (ibf *InvertibleBloomFilter) Subtract(ibf2 *InvertibleBloomFilter) *InvertibleBloomFilter {
	difference := NewIBF(ibf.UniverseSize, ibf.MappingMethod)
	difference.Copy(ibf)

	for j := uint64(0); j < ibf.Size; j++ {
		difference.Cells[j].Subtract(ibf2.Cells[j])
	}

	return difference
}

// Decode attempts to extract the symbols unique to each set represented by the IBF.
// bWithoutA: Symbols in the second set but not in the first.
// aWithoutB: Symbols in the first set but not in the second.
// ok: Whether decoding was successful.
func (ibf *InvertibleBloomFilter) Decode() (bWithoutA []*uint256.Int, aWithoutB []*uint256.Int, ok bool) {
	pureList := make([]uint64, 0)

	for {
		n := len(pureList) - 1

		if n == -1 {
			// Identify pure cells
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

		if ibf.Cells[j].Count > 0 {
			bWithoutA = append(bWithoutA, xorSum)
		} else {
			aWithoutB = append(aWithoutB, xorSum)
		}

		offset := uint64(0)
		// Removed symbol (xorSum) from cells its mapped to.
		for i := uint64(1); i <= ibf.Iteration; i++ {
			cellIdx := offset + ibf.MappingMethod.MapSymbol(xorSum, i)

			// Empty the pure cell at index j at the end
			if cellIdx != j {
				ibf.Cells[cellIdx].Subtract(ibf.Cells[j])
			}

			offset += ibf.MappingMethod.GetAdditionalCellsCount(i)
		}

		ibf.Cells[j].Subtract(ibf.Cells[j])
	}

	// Verify the IBF is empty after decoding
	for j := uint64(0); j < ibf.Size; j++ {
		if !ibf.Cells[j].IsZero() {
			ok = false
			return
		}
	}

	ok = true
	return
}

// IsEmpty checks if the IBF is empty or not.
func (ibf *InvertibleBloomFilter) IsEmpty() bool {
	for _, cell := range ibf.Cells {
		if !cell.IsZero() {
			return false
		}
	}
	return true
}

// GetTransmittedBitsSize calculates the bit size of all transmitted cells.
// This size reflects only the cells used by the IBF.
func (ibf *InvertibleBloomFilter) GetTransmittedBitsSize() uint64 {
	var totalSize uint64
	for _, cell := range ibf.Cells {
		totalSize += cell.BitsLen()
	}
	return totalSize
}
