package certainsync

import (
	"bytes"
	"encoding/binary"
	"errors"
	"fmt"

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
	// Detecting duplicates
	// seenSymbols := make(map[*uint256.Int]bool)

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

		// Check for duplicates
		// if seenSymbols[xorSum] {
		// 	continue
		// }

		//symmetricDiff = append(symmetricDiff, xorSum)
		//seenSymbols[xorSum] = true

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

func (ibf *InvertibleBloomFilter) Reset() {
	ibf.Cells = nil
	ibf.Size = 0
}

// GetTransmittedBitsSize returns the bit size of the actively transmitted cells.
// This reflects only the cells that have been added to the IBF and excludes unutilized cells.
// Other field of the IBF are either private to him like iteration, or
// agreed ahead like hash seed, mapping method with others.
func (ibf *InvertibleBloomFilter) GetTransmittedBitsSize() uint64 {
	var totalSize uint64
	for _, cell := range ibf.Cells {
		totalSize += cell.BitsLen()
	}
	return totalSize
}

// Serialize converts the InvertibleBloomFilter into a byte slice
func (ibf *InvertibleBloomFilter) Serialize() ([]byte, error) {
	buf := new(bytes.Buffer)

	// Serialize each IBFCell
	for _, cell := range ibf.Cells {
		cellBytes, err := cell.Serialize()

		if err != nil {
			return nil, err
		}

		cellByteLen := len(cellBytes)

		// Convert the length of the cell to a byte slice
		lenBytes := []byte{uint8(cellByteLen)} // Serialize the length as uint8

		if _, err := buf.Write(lenBytes); err != nil {
			return nil, err
		}

		if _, err := buf.Write(cellBytes); err != nil {
			return nil, err
		}
	}

	ibf.Reset()

	return buf.Bytes(), nil
}

// Deserialize converts a byte slice back into an InvertibleBloomFilter object.
// Each cell has its own deserialization method, and the size of each cell is indicated by a uint8 byte.
func (ibf *InvertibleBloomFilter) Deserialize(data []byte) error {
	buf := bytes.NewBuffer(data)

	// Iterate through the byte slice, reading one cell at a time
	for buf.Len() > 0 {
		var cellSize uint8
		// Read the size of the current cell
		if err := binary.Read(buf, binary.LittleEndian, &cellSize); err != nil {
			return fmt.Errorf("failed to read cell size: %v", err)
		}

		// Now, read the bytes for the current cell based on its size
		cellData := make([]byte, cellSize)
		if _, err := buf.Read(cellData); err != nil {
			return fmt.Errorf("failed to read cell data: %v", err)
		}

		// Create the appropriate cell and deserialize the data
		cell := &IBFCell{}
		if err := cell.Deserialize(cellData); err != nil {
			return fmt.Errorf("failed to deserialize IBFCell: %v", err)
		}

		// Append the deserialized cell to the IBF Cells slice
		ibf.Cells = append(ibf.Cells, *cell)
	}

	return nil
}
