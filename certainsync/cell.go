package certainsync

import (
	"github.com/holiman/uint256"
)

// Static hasher instance
var cellHasher CellHasher = nil

// SetHasher selects appropriate hash function based on universe size.
func SetHasher(universeSize *uint256.Int) {
	if universeSize.IsUint64() {
		cellHasher = XXHash64Hash{}
	} else {
		cellHasher = Sha256Hash{}
	}
}

// IBFCell represents a single cell in the Invertible Bloom Filter.
// It maintains count, XOR sum of elements, and hash sum for verification.
type IBFCell struct {
	Count   int64
	XorSum  *uint256.Int
	HashSum *uint256.Int
}

// NewIBFCell creates a new initialized IBFCell with
// appropriate hasher based on universe size
func NewIBFCell(universeSize *uint256.Int) IBFCell {
	cell := IBFCell{
		Count:   0,
		XorSum:  uint256.NewInt(0),
		HashSum: uint256.NewInt(0),
	}

	SetHasher(universeSize)
	return cell
}

// Insert adds a symbol to the cell
func (c *IBFCell) Insert(s *uint256.Int) {
	if s == nil {
		return
	}
	c.Count++
	c.XorSum.Xor(c.XorSum, s)

	symbolHash := cellHasher.Hash(s.Bytes())
	c.HashSum.Xor(c.HashSum, symbolHash)
}

// Subtract removes another cell's contents from this cell
func (c *IBFCell) Subtract(other IBFCell) {
	c.Count -= other.Count
	c.XorSum.Xor(c.XorSum, other.XorSum)
	c.HashSum.Xor(c.HashSum, other.HashSum)
}

// IsPure checks if the cell contains exactly one element by verifying
// the count is ±1 and the hash sum matches the computed hash of XorSum
func (c *IBFCell) IsPure() bool {
	if c.Count != 1 && c.Count != -1 {
		return false
	}

	calcHashSum := cellHasher.Hash(c.XorSum.Bytes())
	return c.HashSum.Cmp(calcHashSum) == 0
}

// IsZero checks if the cell is empty
func (c *IBFCell) IsZero() bool {
	return c.Count == 0 &&
		c.XorSum.IsZero() &&
		c.HashSum.IsZero()
}

// GetXorSum returns a copy of the XorSum to prevent external modification
func (c *IBFCell) GetXorSum() *uint256.Int {
	return uint256.NewInt(0).Set(c.XorSum)
}

// Clone creates a deep copy of the cell
func (c *IBFCell) Clone() IBFCell {
	return IBFCell{
		Count:   c.Count,
		XorSum:  uint256.NewInt(0).Set(c.XorSum),
		HashSum: uint256.NewInt(0).Set(c.HashSum),
	}
}

// ByteLen returns the total size of the cell in bytes.
func (c *IBFCell) ByteLen() uint8 {
	// Count (int64)
	var countBytes uint8 = 8

	var xorSumBytes uint8 = 0
	var hashSumBytes uint8 = 0

	switch cellHasher.(type) {
	case XXHash64Hash:
		xorSumBytes = 8 // 64 bits for XXHash64
		hashSumBytes = 8
	case Sha256Hash:
		xorSumBytes = 32 // 64 bits for Sha256Hash
		hashSumBytes = 32
	}

	return countBytes + xorSumBytes + hashSumBytes
}

// BitsLen returns the total size of the cell in bits,
func (c *IBFCell) BitsLen() uint64 {
	return uint64(c.ByteLen()) * 8
}
