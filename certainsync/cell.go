package certainsync

// Cell is an interface that can be either an IBFCell
// or an ExtendedIBFCell
type Cell interface {
	Init(symbolType string, seed ...uint32)
	Insert(s Symbol, orgSymbol ...Symbol)
	Subtract(c2 Cell)
	IsExtended() bool
	IsPure() bool
	IsZero() bool
	IsZeroExtended() bool
	GetXorSum() Symbol
	DeepCopy(c2 Cell)
	GetTransmittedBitsSize() uint64
}

// IBFCell represents a single cell in the Invertible Bloom Filter
type IBFCell struct {
	Count   int64
	XorSum  Symbol
	HashSum Hash
}

// Init assign defualt values to cell's fields
func (c *IBFCell) Init(symbolType string, seed ...uint32) {
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
func (c *IBFCell) Insert(s Symbol, orgSymbol ...Symbol) {
	c.Count++
	c.XorSum = c.XorSum.Xor(s)
	c.HashSum = c.HashSum.Xor(s.Hash())
}

// Subtract removes another cell's contents from this cell
func (c *IBFCell) Subtract(c2 Cell) {
	cell := c2.(*IBFCell)
	c.Count -= cell.Count
	c.XorSum = c.XorSum.Xor(cell.XorSum)
	c.HashSum = c.HashSum.Xor(cell.HashSum)
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

func (c *IBFCell) IsZeroExtended() bool {
	return c.IsZero()
}

// GetXorSum is a getter of XorSum field of IBFCell
func (c *IBFCell) GetXorSum() Symbol {
	return c.XorSum
}

func (c *IBFCell) DeepCopy(c2 Cell) {
	c.Count = c2.(*IBFCell).Count
	c.XorSum = c2.(*IBFCell).XorSum.DeepCopy()
	c.HashSum = c2.(*IBFCell).HashSum.DeepCopy()
}

// GetTransmittedBitsSize calculates the size of the IBFCell in bits,
// which is transmitted with all its fields.
func (c *IBFCell) GetTransmittedBitsSize() uint64 {
	size := uint64(64) // Count is an int64, so 64 bits
	size += c.XorSum.SizeInBits()
	size += c.HashSum.SizeInBits()
	return size
}

// ExtendedIBFCell extends IBFCell with a full hash sum capability
// for blockchain-specific applications
type ExtendedIBFCell struct {
	IBFCell
	Seed       uint32
	FullXorSum HashSymbol
}

// Init assign defualt values to extended cell's fields
func (c *ExtendedIBFCell) Init(symbolType string, seed ...uint32) {
	c.IBFCell.Init(symbolType, seed...)
	c.Seed = seed[0]
	c.FullXorSum = HashSymbol{}
}

// Insert adds a symbol to the extended cell, including FullXorSum
func (c *ExtendedIBFCell) Insert(s Symbol, orgSymbol ...Symbol) {
	c.Count++
	c.XorSum = c.XorSum.Xor(s)
	c.HashSum = c.HashSum.Xor(s.Hash(c.Seed))
	c.FullXorSum = c.FullXorSum.Xor(orgSymbol[0]).(HashSymbol)
}

// Subtract removes another cell's contents from this extended cell
func (c *ExtendedIBFCell) Subtract(c2 Cell) {
	cell := c2.(*ExtendedIBFCell)
	c.IBFCell.Subtract(&cell.IBFCell)
	c.FullXorSum = c.FullXorSum.Xor(cell.FullXorSum).(HashSymbol)
}

// IsExtended checks if the cell is extended type
func (c *ExtendedIBFCell) IsExtended() bool {
	return true
}

// IsPure checks if the extended cell contains exactly one element
func (c *ExtendedIBFCell) IsPure() bool {
	return (c.Count == 1 || c.Count == -1) &&
		c.HashSum == c.XorSum.Hash(c.Seed)
}

// IsZero checks if the extended cell is empty
func (c *ExtendedIBFCell) IsZero() bool {
	return c.IBFCell.IsZero()
}

// IsZeroExtended checks if the extended cell is empty,
// including its extra FullXorSum field.
func (c *ExtendedIBFCell) IsZeroExtended() bool {
	return c.IsZero() && c.FullXorSum.IsZero()
}

// GetXorSum is a getter of XorSum field of ExtendedIBFCell
func (c *ExtendedIBFCell) GetXorSum() Symbol {
	return c.XorSum
}

func (c *ExtendedIBFCell) DeepCopy(c2 Cell) {
	c.Count = c2.(*ExtendedIBFCell).Count
	c.XorSum = c2.(*ExtendedIBFCell).XorSum.DeepCopy()
	c.HashSum = c2.(*ExtendedIBFCell).HashSum.DeepCopy()
	c.Seed = c2.(*ExtendedIBFCell).Seed
	c.FullXorSum = c2.(*ExtendedIBFCell).FullXorSum.DeepCopy().(HashSymbol)
}

// GetTransmittedBitsSize calculates the size of the ExtendedIBFCell
// in bits, which is transmitted with all its fields except from seed
// that is given in initialization once.
func (c *ExtendedIBFCell) GetTransmittedBitsSize() uint64 {
	size := c.IBFCell.GetTransmittedBitsSize()
	// Seed is a uint32, so 32 bits
	size += uint64(32)
	size += c.FullXorSum.SizeInBits()
	return size
}
