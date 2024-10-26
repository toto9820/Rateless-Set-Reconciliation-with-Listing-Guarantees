package riblt_with_certainty

type IBFCell struct {
	Count   int64
	XorSum  Symbol
	HashSum Hash
}

type InvertibleBloomFilter struct {
	Cells         []IBFCell
	Iteration     uint64
	Size          uint64
	SymbolType    string        // "hash" or "uint64"
	MappingMethod MappingMethod // "EGH", "OLS" or "Extended Hamming"
}

func NewIBF(size uint64, symbolType string, mapping MappingMethod) *InvertibleBloomFilter {
	var zeroSymbol Symbol
	var zeroHashSum Hash
	switch symbolType {
	case "hash":
		zeroSymbol = HashSymbol{}
		zeroHashSum = CommonHash{}
	case "uint64":
		zeroSymbol = Uint64Symbol(0)
		zeroHashSum = Uint64Hash(0)
	default:
		panic("Invalid symbol type")
	}

	cells := make([]IBFCell, size)

	for i := range cells {
		cells[i].XorSum = zeroSymbol
		cells[i].HashSum = zeroHashSum
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
	copy(ibf.Cells, ibf2.Cells)
	ibf.Iteration = ibf2.Iteration
	ibf.Size = ibf2.Size
	ibf.SymbolType = ibf2.SymbolType
	ibf.MappingMethod = ibf2.MappingMethod
}

func (c *IBFCell) Insert(s Symbol) {
	c.Count++
	c.XorSum = c.XorSum.Xor(s)
	c.HashSum = c.HashSum.Xor(s.Hash())
}

func (c *IBFCell) Subtract(c2 *IBFCell) {
	c.Count -= c2.Count
	c.XorSum = c.XorSum.Xor(c2.XorSum)
	c.HashSum = c.HashSum.Xor(c2.HashSum)
}

func (c *IBFCell) IsPure() bool {
	return (c.Count == 1 || c.Count == -1) && c.HashSum == c.XorSum.Hash()
}

func (c *IBFCell) IsZero() bool {
	return c.Count == 0 && c.XorSum.IsZero() && c.HashSum.IsZero()
}

func (ibf *InvertibleBloomFilter) AddSymbols(symbols []Symbol) {
	additionalCellsCount := ibf.MappingMethod.GetAdditionalCellsCount(ibf.SymbolType, ibf.Iteration)

	if ibf.Size+additionalCellsCount > uint64(len(ibf.Cells)) {
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
		difference.Cells[j].Subtract(&ibf2.Cells[j])
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

		count := ibf.Cells[j].Count
		xorSum := ibf.Cells[j].XorSum

		symmetricDiff = append(symmetricDiff, xorSum)

		offset := uint64(0)
		for i := uint64(0); i < ibf.Iteration; i++ {
			cellIdx := offset + ibf.MappingMethod.MapSymbol(xorSum, i)
			ibf.Cells[cellIdx].Count -= count
			ibf.Cells[cellIdx].XorSum = ibf.Cells[cellIdx].XorSum.Xor(xorSum)
			ibf.Cells[cellIdx].HashSum = ibf.Cells[cellIdx].HashSum.Xor(xorSum.Hash())

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
