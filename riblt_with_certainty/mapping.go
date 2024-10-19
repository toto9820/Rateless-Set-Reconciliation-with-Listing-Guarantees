package riblt_with_certainty

type MappingMethod interface {
	MapSymbol(s Symbol, iteration uint64) uint64
	GetAdditionalCellsCount(symbolType string, iteration uint64) uint64
}

type EGHMapping struct{}

func (e *EGHMapping) MapSymbol(s Symbol, iteration uint64) uint64 {
	curPrime := primes[iteration]
	return ModSymbolUint64(s, curPrime)
}

func (e *EGHMapping) GetAdditionalCellsCount(symbolType string, iteration uint64) uint64 {
	curPrime := primes[iteration]
	return curPrime
}

// const defualtOLSChunk uint64 = 1000

// type OLSMapping struct {
// 	s uint64 // Order of each Latin square
// 	n uint64 // Total number of symbols
// }

// func (o *OLSMapping) MapSymbol(symbol Symbol, iteration uint64) uint64 {
// 	row := iteration % uint64(len(o.squares))                   // Select square based on iteration
// 	col := ModSymbolUint64(symbol, uint64(len(o.squares[row]))) // Use symbol to determine column
// 	return o.squares[row][col]                                  // Return mapped value from the OLS
// }

// func (o *OLSMapping) GetAdditionalCellsCount(symbolType string, iteration uint64) uint64 {
// 	switch symbolType {
// 	// No need for sqrt(n) as n is really big for hash type,
// 	// while its transaction pool relatively small compared to it.
// 	case "hash":
// 		return defualtOLSChunk
// 	case "uint64":
// 		return o.s
// 	default:
// 		panic("Invalid symbol type")
// 	}
// }
