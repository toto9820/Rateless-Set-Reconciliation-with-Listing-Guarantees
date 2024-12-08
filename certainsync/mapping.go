package certainsync

import "github.com/holiman/uint256"

// MappingMethod is an interface for defining symbol-to-cell mapping methods.
// It provides methods for determining the target cell for a given symbol
// and calculating the number of additional cells required for each iteration.
type MappingMethod interface {
	// MapSymbol maps a symbol to a specific cell for the given iteration.
	MapSymbol(s *uint256.Int, iteration uint64) uint64
	// GetAdditionalCellsCount returns the number of additional cells required
	// for the given iteration.
	GetAdditionalCellsCount(iteration uint64) uint64
}

// EGHMapping is a mapping method using a prime-based modulo operation.
type EGHMapping struct{}

// MapSymbol maps a symbol based on the current prime in the primes array.
func (e *EGHMapping) MapSymbol(s *uint256.Int, iteration uint64) uint64 {
	curPrime := primes[iteration-1]
	modResult := new(uint256.Int).Mod(s, uint256.NewInt(curPrime))
	return modResult.Uint64()
}

// GetAdditionalCellsCount returns the additional cell count based
// on the current prime.
func (e *EGHMapping) GetAdditionalCellsCount(iteration uint64) uint64 {
	curPrime := primes[iteration-1]
	return curPrime
}

// OLSMapping is a mapping method using an Orthogonal Latin Square approach.
type OLSMapping struct {
	Order uint64 // Order of each Latin square
}

// MapSymbol maps a symbol using the OLS method for a given iteration.
func (o *OLSMapping) MapSymbol(symbol *uint256.Int, iteration uint64) uint64 {
	latinSquareNum := iteration - 1

	// Copy symbol and subtract 1 (simulating `symbol - 1`).
	symbolIndex := new(uint256.Int).Sub(symbol, uint256.NewInt(1))

	// Calculate row and column for the symbol in the Latin square.
	row := new(uint256.Int).Div(symbolIndex, uint256.NewInt(o.Order)).Uint64()
	col := new(uint256.Int).Mod(symbolIndex, uint256.NewInt(o.Order)).Uint64()

	if latinSquareNum == 0 {
		return row
	}

	// Calculate mapped value for non-zero Latin square numbers.
	mappedValue := (col + (row * latinSquareNum)) % o.Order
	return mappedValue
}

// GetAdditionalCellsCount returns the order of the Latin square.
func (o *OLSMapping) GetAdditionalCellsCount(iteration uint64) uint64 {
	return o.Order
}
