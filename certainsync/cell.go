package certainsync

// IBFCell represents a single cell in the Invertible Bloom Filter
type IBFCell struct {
	Count   int64
	XorSum  Symbol
	HashSum Hash
}

// Init assign defualt values to cell's fields
func (c *IBFCell) Init(symbolType string) {
	c.Count = 0

	switch symbolType {
	case HashSymbolType:
		c.XorSum = HashSymbol{}
	case Uint64SymbolType:
		c.XorSum = Uint64Symbol(0)
	case Uint32SymbolType:
		c.XorSum = Uint32Symbol(0)
	default:
		// Handle an unexpected type if necessary
		panic("unsupported Symbol type")
	}

	c.HashSum = c.XorSum.Hash()
}

// Insert adds a symbol to the cell
func (c *IBFCell) Insert(s Symbol) {
	c.Count++
	c.XorSum = c.XorSum.Xor(s)
	c.HashSum = c.HashSum.Xor(s.Hash())
}

// Subtract removes another cell's contents from this cell
func (c *IBFCell) Subtract(c2 IBFCell) {
	c.Count -= c2.Count
	c.XorSum = c.XorSum.Xor(c2.XorSum)
	c.HashSum = c.HashSum.Xor(c2.HashSum)
}

// IsExtended checks if the cell is extended type
func (c *IBFCell) IsExtended() bool {
	return false
}

// IsPure checks if the cell contains exactly one element
func (c *IBFCell) IsPure() bool {
	return (c.Count == 1 || c.Count == -1) &&
		c.HashSum == c.XorSum.Hash()
}

// IsZero checks if the cell is empty
func (c *IBFCell) IsZero() bool {
	return c.Count == 0 &&
		c.XorSum.IsZero() &&
		c.HashSum.IsZero()
}

// GetXorSum is a getter of XorSum field of IBFCell
func (c *IBFCell) GetXorSum() Symbol {
	return c.XorSum
}

func (c *IBFCell) DeepCopy(c2 IBFCell) {
	c.Count = c2.Count
	c.XorSum = c2.XorSum.DeepCopy()
	c.HashSum = c2.HashSum.DeepCopy()
}

func (c *IBFCell) SizeInBits() uint64 {
	size := uint64(64) // Count is an int64, so 64 bits
	size += c.XorSum.SizeInBits()
	size += c.HashSum.SizeInBits()
	return size
}
