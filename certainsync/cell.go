package certainsync

// Cell is an interface that can be either an IBFCell
// or an ExtendedIBFCell
type Cell interface {
	Init(seed ...uint32)
	Insert(s Symbol)
	Subtract(c Cell)
	IsPure() bool
	IsZero() bool
	GetXorSum() Symbol
}

// IBFCell represents a single cell in the Invertible Bloom Filter
type IBFCell struct {
	Count   int64
	XorSum  Symbol
	HashSum Hash
}

// Init assign defualt values to cell's fields
func (c *IBFCell) Init(seed ...uint32) {
	c.Count = 0
	c.XorSum = c.XorSum.NewSymbol()
	c.HashSum = c.HashSum.NewHash()
}

// Insert adds a symbol to the cell
func (c *IBFCell) Insert(s Symbol) {
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

// ExtendedIBFCell extends IBFCell with a full hash sum capability
// for blockchain-specific applications
type ExtendedIBFCell struct {
	IBFCell
	Seed       uint32
	FullXorSum HashSymbol
}

// Init assign defualt values to extended cell's fields
func (c *ExtendedIBFCell) Init(seed ...uint32) {
	c.IBFCell.Init()
	c.Seed = seed[0]
	c.FullXorSum = HashSymbol{}
}

// Insert adds a symbol to the extended cell, including FullXorSum
func (c *ExtendedIBFCell) Insert(s Symbol) {
	c.Count++
	c.XorSum = c.XorSum.Xor(s)
	c.HashSum = c.HashSum.Xor(s.Hash(c.Seed))
	c.FullXorSum = c.FullXorSum.Xor(s).(HashSymbol)
}

// Subtract removes another cell's contents from this extended cell
func (c *ExtendedIBFCell) Subtract(c2 Cell) {
	cell := c2.(*ExtendedIBFCell)
	c.IBFCell.Subtract(cell)
	c.FullXorSum = c.FullXorSum.Xor(cell.FullXorSum).(HashSymbol)
}

// IsPure checks if the extended cell contains exactly one element
func (c *ExtendedIBFCell) IsPure() bool {
	return (c.Count == 1 || c.Count == -1) &&
		c.HashSum == c.XorSum.Hash(c.Seed)
}

// IsZero checks if the extended cell is empty
func (c *ExtendedIBFCell) IsZero() bool {
	return c.IBFCell.IsZero() && c.FullXorSum.IsZero()
}

// GetXorSum is a getter of XorSum field of ExtendedIBFCell
func (c *ExtendedIBFCell) GetXorSum() Symbol {
	return c.XorSum
}
