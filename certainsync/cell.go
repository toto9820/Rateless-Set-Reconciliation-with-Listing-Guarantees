package certainsync

import (
	"bytes"
	"encoding/binary"
	"fmt"

	"github.com/holiman/uint256"
)

// IBFCell represents a single cell in the Invertible Bloom Filter.
// It maintains count, XOR sum of elements, and hash sum for verification.
type IBFCell struct {
	Count   int64
	XorSum  *uint256.Int
	HashSum *uint256.Int
	Hasher  CellHasher
}

// NewIBFCell creates a new initialized IBFCell with
// appropriate hasher based on universe size
func NewIBFCell(universeSize *uint256.Int) IBFCell {
	cell := IBFCell{
		Count:   0,
		XorSum:  uint256.NewInt(0),
		HashSum: uint256.NewInt(0),
	}
	cell.setHasher(universeSize)
	// cell.HashSum = cell.Hasher.Hash(nil)
	return cell
}

// setHasher selects appropriate hash function based on universe size
// to maintain collision probability below 0.1%:
// 1 - e^(-n(n-1)/(2*m)) < 0.001
func (c *IBFCell) setHasher(universeSize *uint256.Int) {
	threshold1 := uint256.NewInt(2500)
	threshold2 := uint256.NewInt(150_000_000)

	switch {
	case universeSize.Cmp(threshold1) < 0:
		c.Hasher = Murmur3Hash{}
	case universeSize.Cmp(threshold2) < 0:
		c.Hasher = XXHash64Hash{}
	default:
		c.Hasher = Sha256Hash{}
	}
}

// Insert adds a symbol to the cell
func (c *IBFCell) Insert(s *uint256.Int) {
	if s == nil {
		return
	}
	c.Count++
	c.XorSum.Xor(c.XorSum, s)

	symbolHash := c.Hasher.Hash(s.Bytes())
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

	calcHashSum := c.Hasher.Hash(c.XorSum.Bytes())
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
		Hasher:  c.Hasher,
	}
}

// ByteLen returns the total size of the cell in bytes, which
// corresponds to the size of the cell when serialized.
func (c *IBFCell) ByteLen() uint8 {
	return 8 + // Count (int64)
		uint8(c.XorSum.ByteLen()) +
		uint8(c.HashSum.ByteLen())
}

// BitsLen returns the total size of the cell in bits, which
// corresponds to the size of the cell when serialized
func (c *IBFCell) BitsLen() uint64 {
	return uint64(c.ByteLen()) * 8
}

func (c *IBFCell) Serialize() ([]byte, error) {
	var buf bytes.Buffer

	// Serialize the Count as int64 (8 bytes)
	if err := binary.Write(&buf, binary.LittleEndian, c.Count); err != nil {
		return nil, fmt.Errorf("failed to write Count: %v", err)
	}

	// Serialize XorSum (byte slice)
	var xorSumLen uint8 = uint8(c.XorSum.ByteLen())

	if err := binary.Write(&buf, binary.LittleEndian, xorSumLen); err != nil {
		return nil, fmt.Errorf("failed to write XorSum length: %v", err)
	}

	xorSumBytes := c.XorSum.Bytes()

	if err := binary.Write(&buf, binary.LittleEndian, xorSumBytes); err != nil {
		return nil, fmt.Errorf("failed to write XorSum: %v", err)
	}

	// Serialize HashSum (byte slice)

	var hashSumLen uint8 = uint8(c.HashSum.ByteLen())

	if err := binary.Write(&buf, binary.LittleEndian, hashSumLen); err != nil {
		return nil, fmt.Errorf("failed to write HashSum length: %v", err)
	}

	hashSumBytes := c.HashSum.Bytes()

	if err := binary.Write(&buf, binary.LittleEndian, hashSumBytes); err != nil {
		return nil, fmt.Errorf("failed to write HashSum: %v", err)
	}

	return buf.Bytes(), nil
}

// Deserialize converts a byte slice back into an IBFCell object
func (c *IBFCell) Deserialize(data []byte) error {
	buf := bytes.NewBuffer(data)

	// Deserialize the Count field
	if err := binary.Read(buf, binary.LittleEndian, &c.Count); err != nil {
		return fmt.Errorf("failed to read Count: %v", err)
	}

	// Deserialize the XorSum field as a byte slice
	var xorSumLen uint8
	if err := binary.Read(buf, binary.LittleEndian, &xorSumLen); err != nil {
		return fmt.Errorf("failed to read XorSum length: %v", err)
	}

	xorSumBytes := make([]byte, xorSumLen)
	if err := binary.Read(buf, binary.LittleEndian, &xorSumBytes); err != nil {
		return fmt.Errorf("failed to read XorSum: %v", err)
	}
	c.XorSum = uint256.NewInt(0).SetBytes(xorSumBytes)

	// Deserialize the HashSum field as a byte slice
	var hashSumLen uint8
	if err := binary.Read(buf, binary.LittleEndian, &hashSumLen); err != nil {
		return fmt.Errorf("failed to read HashSum length: %v", err)
	}

	hashSumBytes := make([]byte, hashSumLen)
	if err := binary.Read(buf, binary.LittleEndian, &hashSumBytes); err != nil {
		return fmt.Errorf("failed to read HashSum: %v", err)
	}
	c.HashSum = uint256.NewInt(0).SetBytes(hashSumBytes)

	return nil
}
