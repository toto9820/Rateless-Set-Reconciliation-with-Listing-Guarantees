package certainsync

type MappingMethod interface {
	MapSymbol(s Symbol, iteration uint64) uint64
	GetAdditionalCellsCount(symbolType string, iteration uint64) uint64
}

type EGHMapping struct{}

func (e *EGHMapping) MapSymbol(s Symbol, iteration uint64) uint64 {
	curPrime := primes[iteration-1]
	return ModSymbolUint64(s, curPrime)
}

func (e *EGHMapping) GetAdditionalCellsCount(symbolType string, iteration uint64) uint64 {
	curPrime := primes[iteration-1]
	return curPrime
}

type OLSMapping struct {
	Order uint64 // Order of each Latin square
}

func (o *OLSMapping) MapSymbol(symbol Symbol, iteration uint64) uint64 {
	latinSquareNum := iteration - 1
	symbolIndex := SubstractSymbolUint64(symbol, 1)

	// Calculate row and column for the symbol
	row := symbolIndex / o.Order
	col := symbolIndex % o.Order

	if latinSquareNum == 0 {
		return row
	}

	mappedValue := (col + (row * latinSquareNum)) % o.Order

	return mappedValue
}

func (o *OLSMapping) GetAdditionalCellsCount(symbolType string, iteration uint64) uint64 {
	return o.Order
}
