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
	Cells         []IBFCell
	UniverseSize  *uint256.Int
	Iteration     uint64
	Size          uint64
	MappingMethod MappingMethod
}

// NewIBF creates a new InvertibleBloomFilter instance
func NewIBF(universeSize *uint256.Int, mapping MappingMethod) *InvertibleBloomFilter {
	return &InvertibleBloomFilter{
		Cells:         nil,
		UniverseSize:  universeSize.Clone(),
		Iteration:     0,
		Size:          0,
		MappingMethod: mapping,
	}
}

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

func (ibf *InvertibleBloomFilter) AddSymbols(symbols []*uint256.Int) {
	ibf.Iteration++
	additionalCellsCount := ibf.MappingMethod.GetAdditionalCellsCount(ibf.Iteration)

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

func (ibf *InvertibleBloomFilter) Subtract(ibf2 *InvertibleBloomFilter) *InvertibleBloomFilter {
	difference := NewIBF(ibf.UniverseSize, ibf.MappingMethod)
	difference.Copy(ibf)

	for j := uint64(0); j < ibf.Size; j++ {
		difference.Cells[j].Subtract(ibf2.Cells[j])
	}

	return difference
}

func (ibf *InvertibleBloomFilter) Decode() (bWithoutA []*uint256.Int, aWithoutB []*uint256.Int, ok bool) {
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

		if ibf.Cells[j].Count > 0 {
			bWithoutA = append(bWithoutA, xorSum)
		} else {
			aWithoutB = append(aWithoutB, xorSum)
		}

		offset := uint64(0)
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

	for j := uint64(0); j < ibf.Size; j++ {
		if !ibf.Cells[j].IsZero() {
			ok = false
			return
		}
	}

	ok = true
	return
}

func (ibf *InvertibleBloomFilter) IsEmpty() bool {
	for _, cell := range ibf.Cells {
		if !cell.IsZero() {
			return false
		}
	}
	return true
}

// GetTransmittedBitsSize returns the bit size of the actively transmitted cells.
// This reflects only the cells that have been added to the IBF.
// Other field of the IBF are either private to him like iteration, or
// agreed ahead like universe size, hash seed, mapping method with others.
func (ibf *InvertibleBloomFilter) GetTransmittedBitsSize() uint64 {
	var totalSize uint64
	for _, cell := range ibf.Cells {
		totalSize += cell.BitsLen()
	}
	return totalSize
}
