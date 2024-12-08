package main

import (
	"context"
	"encoding/csv"
	"encoding/json"
	"fmt"
	"log"
	"math"
	"os"
	"path/filepath"
	"time"

	"github.com/ethereum/go-ethereum/rpc"
	"github.com/holiman/uint256"
	"github.com/spaolacci/murmur3"
	. "github.com/toto9820/Rateless-Set-Reconciliation-with-Listing-Guarantees/certainsync"
)

// Config represents the structure of the configuration file.
type Config struct {
	Node1IPC       string `json:"node1_ipc"`
	Node2IPC       string `json:"node2_ipc"`
	Node1HashesDir string `json:"node1_hashes_dir"`
	Node2HashesDir string `json:"node2_hashes_dir"`
}

// loadConfig loads the configuration from a JSON file.
func loadConfig(filePath string) (*Config, error) {
	file, err := os.ReadFile(filePath)
	if err != nil {
		return nil, err
	}
	var config Config
	err = json.Unmarshal(file, &config)
	if err != nil {
		return nil, err
	}
	return &config, nil
}

// certainSync generates IBFs for two sets of
// transaction hashes, compares them, and finds the
// symmetric difference.
func certainSync(hashes1, hashes2 []*uint256.Int, universeSize *uint256.Int, mappingType MappingType) (int, uint64) {
	var ibfNode1, ibfNode2 *InvertibleBloomFilter

	var mapping MappingMethod
	switch mappingType {
	case EGH:
		mapping = &EGHMapping{}
	case OLS:
		mapping = &OLSMapping{
			Order: uint64(math.Ceil(math.Sqrt(float64(universeSize.Uint64())))),
		}
	default:
		panic("unsupported mapping type")
	}

	ibfNode1 = NewIBF(universeSize, mapping)
	ibfNode2 = NewIBF(universeSize, mapping)

	transmittedBits := uint64(0)

	for {
		ibfNode1.AddSymbols(hashes1)

		transmittedBits += ibfNode1.GetTransmittedBitsSize()

		ibfNode2.AddSymbols(hashes2)

		// Subtract the two IBFs
		ibfDiff := ibfNode2.Subtract(ibfNode1)
		hashes2Not1, hashes1Not2, ok := ibfDiff.Decode()

		if ok {
			// Sending back to node1 transactions of 2/1 where each
			// transaction is 256 bit.
			transmittedBits += uint64(len(hashes2Not1)) * 256

			return len(hashes2Not1) + len(hashes1Not2), transmittedBits
		}
	}
}

// UniverseReduceSync compares two sets of transaction hashes using Invertible Bloom Filters
// and finds their symmetric difference. It supports different mapping methods (EGH or OLS).
// Returns the size of the symmetric difference and the total number of transmitted bits.
func universeReduceSync(originalHashes1, originalHashes2 []*uint256.Int, delta float64, mappingType MappingType) (int, uint64) {
	// Create working copies of the input slices
	totalHashes1 := make([]*uint256.Int, len(originalHashes1))
	totalHashes2 := make([]*uint256.Int, len(originalHashes2))

	copy(totalHashes1, originalHashes1)
	copy(totalHashes2, originalHashes2)

	var allHashes1Not2, allHashes2Not1 []*uint256.Int
	var totalSymmetricDiff []*uint256.Int
	var roundNumber uint64 = 1
	transmittedBits := uint64(0)

	for {
		var sizeS1, sizeS2 uint64
		var convertedHashes2Not1, convertedHashes1Not2 []*uint256.Int
		var ok bool
		var roundSymmetricDiffSize int
		var roundTransmittedBits uint64 = 0
		var ibfDiff *InvertibleBloomFilter

		sizeS1 = uint64(len(totalHashes1))
		sizeS2 = uint64(len(totalHashes2))
		reducedUniverSizeUint64 := universeSizeReduction(sizeS1, sizeS2, delta)
		reducedUniverseSize := uint256.NewInt(reducedUniverSizeUint64)

		var mapping MappingMethod
		switch mappingType {
		case EGH:
			mapping = &EGHMapping{}
		case OLS:
			mapping = &OLSMapping{
				Order: uint64(math.Ceil(math.Sqrt(float64(reducedUniverSizeUint64)))),
			}
		default:
			panic("unsupported mapping type")
		}

		// Convert each symbol to a reduced symbol
		// Maps to convert reduced symbol back to original symbol
		convertedHashes1, hashMap1 := certainMapping(totalHashes1, roundNumber, reducedUniverseSize)
		convertedHashes2, hashMap2 := certainMapping(totalHashes2, roundNumber, reducedUniverseSize)

		ibfNode1 := NewIBF(reducedUniverseSize, mapping)
		ibfNode2 := NewIBF(reducedUniverseSize, mapping)

		for {
			ibfNode1.AddSymbols(convertedHashes1)

			roundTransmittedBits = ibfNode1.GetTransmittedBitsSize()

			ibfNode2.AddSymbols(convertedHashes2)

			// Subtract the two IBFs
			ibfDiff = ibfNode2.Subtract(ibfNode1)
			convertedHashes2Not1, convertedHashes1Not2, ok = ibfDiff.Decode()

			roundSymmetricDiffSize = len(convertedHashes2Not1) + len(convertedHashes1Not2)

			// Checking if IBLT of symmmetric difference is empty
			if ok {
				break
			}
		}

		transmittedBits += roundTransmittedBits

		// Split the symmetric difference into 1\2 and 2\1
		var hashes1Not2, hashes2Not1 []*uint256.Int

		for _, hash := range convertedHashes1Not2 {
			hashes1Not2 = append(hashes1Not2, hashMap1[hash.String()]...)
		}

		for _, hash := range convertedHashes2Not1 {
			hashes2Not1 = append(hashes2Not1, hashMap2[hash.String()]...)
		}

		// Sending to node1 transactions of 2/1 where each
		// transaction is 256 bit, and 1/2 of converted transactions.
		transmittedBits += uint64(len(convertedHashes1Not2)*reducedUniverseSize.ByteLen()) * 8
		transmittedBits += uint64(len(hashes2Not1)) * 256

		// Sending to node2 transactions of 1/2 where each
		// transaction is 256 bit.
		transmittedBits += uint64(len(hashes1Not2)) * 256

		// Accumulate found differences
		allHashes1Not2 = append(allHashes1Not2, hashes1Not2...)
		allHashes2Not1 = append(allHashes2Not1, hashes2Not1...)

		totalSymmetricDiff = append(totalSymmetricDiff, hashes1Not2...)
		totalSymmetricDiff = append(totalSymmetricDiff, hashes2Not1...)

		// Remove found elements from remaining hashes
		//remainingHashes1 = removeSymbols(remainingHashes1, hashes1Not2)
		//remainingHashes2 = removeSymbols(remainingHashes2, hashes2Not1)

		totalHashes1 = addSymbols(totalHashes1, hashes2Not1)
		totalHashes2 = addSymbols(totalHashes2, hashes1Not2)

		if roundSymmetricDiffSize == 0 {
			totalDiffSize := len(allHashes1Not2) + len(allHashes2Not1)
			return totalDiffSize, transmittedBits
		}

		roundNumber++

	}
}

// Function to calculate the probability of a collision
func collisionProbability(nr uint64, m uint64) float64 {
	// Calculate the product n_r * (n_r - 1) * ... * (n_r - m + 1)
	// Using log to avoid overflow in large numbers
	logNoCollisionProb := float64(0)

	// Compute the log of the product term by term
	for i := uint64(0); i < m; i++ {
		logNoCollisionProb += math.Log(float64(nr-i)) - math.Log(float64(nr))
	}

	// The probability is 1 - exp(logProb)
	collisionProb := float64(1) - math.Exp(logNoCollisionProb)
	return collisionProb
}

// Function to calculate the expected number of collisions
func expectedCollisions(nr uint64, m uint64) float64 {
	// Expected number of collisions - m * (m - 1) / 2 * n
	expectedCollisions := float64(m*(m-1)) / (2 * float64(nr))

	return expectedCollisions
}

// Function to implement the UniverseSizeReduction algorithm
func universeSizeReduction(sizeS1 uint64, sizeS2 uint64, delta float64) uint64 {
	i := uint64(0)
	m := sizeS1 + sizeS2

	for {
		// Calculate n_r = 2^ceil(log2(m) + i)
		nr := uint64(math.Pow(2, math.Ceil(math.Log2(float64(m))+float64(i))))

		// Calculate the collision probability
		// probability := collisionProbability(nr, m)

		// Calculate the expected number of collisions
		expectedCollisions := expectedCollisions(nr, m)

		// If the collision probability is below the threshold, break the loop
		// if probability < delta {
		// 	return nr
		// }

		// delta is collisions threshold
		if expectedCollisions <= delta {
			return nr
		}

		// Increment i to increase n_r in the next iteration
		i++
	}

}

// certainMapping hashes each Symbol to a symbol in a reduced universe
// size with the use of generated hash salt based on the round number, and
// returns the hashed symbols along with a map to convert back to the
// original hashes.
func certainMapping(symbols []*uint256.Int, roundNumber uint64, universeSize *uint256.Int) ([]*uint256.Int, map[string][]*uint256.Int) {
	convertedSymbols := make([]*uint256.Int, 0)
	symbolMap := make(map[string][]*uint256.Int)
	seen := make(map[string]bool)

	hashSalt := GenerateRandomSalt(roundNumber)

	for _, symbol := range symbols {
		convertedSymbolUint64 := murmur3.Sum64WithSeed(symbol.Bytes(), hashSalt)
		convertedSymbol := uint256.NewInt(convertedSymbolUint64)
		convertedSymbol = convertedSymbol.Mod(convertedSymbol, universeSize)
		// Converted symbols are positive
		convertedSymbol = convertedSymbol.AddUint64(convertedSymbol, 1)

		convertedSymbolStr := convertedSymbol.String()

		// If this is the first time seeing this converted symbol, add it
		// to uniqueSymbols
		if !seen[convertedSymbolStr] {
			convertedSymbols = append(convertedSymbols, convertedSymbol)
			seen[convertedSymbolStr] = true
			// Initialize the slice for this converted symbol
			symbolMap[convertedSymbolStr] = make([]*uint256.Int, 0)
		}

		symbolMap[convertedSymbolStr] = append(symbolMap[convertedSymbolStr], symbol)
	}

	return convertedSymbols, symbolMap
}

// removeSymbols removes the specified symbols from the source slice
func removeSymbols(source []*uint256.Int, toRemove []*uint256.Int) []*uint256.Int {
	if len(toRemove) == 0 {
		return source
	}

	// Create a map for quick lookup of symbols to remove
	removeMap := make(map[string]bool)
	for _, sym := range toRemove {
		removeMap[sym.String()] = true
	}

	// Create new slice with non-removed elements
	result := make([]*uint256.Int, 0, len(source))
	for _, sym := range source {
		if !removeMap[sym.String()] {
			result = append(result, sym)
		}
	}

	return result
}

// addSymbols adds the specified symbols to the source slice
func addSymbols(source []*uint256.Int, toAdd []*uint256.Int) []*uint256.Int {
	if len(toAdd) == 0 {
		return source
	}

	// Create a map to track existing symbols for efficient deduplication
	existingSymbols := make(map[string]bool)
	for _, sym := range source {
		existingSymbols[sym.String()] = true
	}

	// Create a new slice with the original source symbols
	result := make([]*uint256.Int, len(source))
	copy(result, source)

	// Add new symbols that are not already present
	for _, sym := range toAdd {
		symStr := sym.String()
		if !existingSymbols[symStr] {
			result = append(result, sym)
			existingSymbols[symStr] = true
		}
	}

	return result
}

// saveSymmetricDiffStatsToCSV saves the time,
// symmetric difference size, and total cells to a CSV file.
func saveSymmetricDiffStatsToCSV(filePath string, iterationCount int, symDiffSize, totalCells uint64) error {
	fileExists := true

	// Check if the file already exists
	if _, err := os.Stat(filePath); os.IsNotExist(err) {
		fileExists = false
	}

	file, err := os.OpenFile(filePath, os.O_WRONLY|os.O_CREATE|os.O_APPEND, 0644)
	if err != nil {
		return err
	}
	defer file.Close()

	writer := csv.NewWriter(file)
	defer writer.Flush()

	// Write header if the file does not exist
	if !fileExists {
		header := []string{"Time (minutes)", "Symmetric Difference Size", "Total Bits"}
		if err := writer.Write(header); err != nil {
			return err
		}
	}

	cellSizeInBits := uint64(64 * 3)

	// Write data row
	record := []string{
		fmt.Sprintf("%d", iterationCount),
		fmt.Sprintf("%d", symDiffSize),
		fmt.Sprintf("%d", totalCells*cellSizeInBits),
	}

	if err := writer.Write(record); err != nil {
		return err
	}

	return nil
}

// txpool_sync performs TxPool synchronization between
// two blockchain nodes in real time.
func txpool_sync() {
	cwd, err := os.Getwd()
	if err != nil {
		log.Fatalf("Failed to get current working directory: %v", err)
	}

	configPath := filepath.Join(cwd, "Configuration", "config.json")
	config, err := loadConfig(configPath)
	if err != nil {
		log.Fatalf("Failed to load config: %v", err)
	}

	node1, err := rpc.Dial(config.Node1IPC)
	if err != nil {
		log.Fatalf("Failed to connect to Node 1 Ethereum client: %v", err)
	}

	node2, err := rpc.Dial(config.Node2IPC)
	if err != nil {
		log.Fatalf("Failed to connect to Node 2 Ethereum client: %v", err)
	}

	node1Dir := filepath.Join(cwd, "data", "blockchain", "node1")
	node2Dir := filepath.Join(cwd, "data", "blockchain", "node2")

	// Create directories if not exist
	if err := os.MkdirAll(node1Dir, os.ModePerm); err != nil {
		log.Fatalf("Failed to create directory for Node 1: %v", err)
	}
	if err := os.MkdirAll(node2Dir, os.ModePerm); err != nil {
		log.Fatalf("Failed to create directory for Node 2: %v", err)
	}

	symmetricDiffStatsFilePath := filepath.Join(cwd, "data", "blockchain", "symmetric_diff_stats.csv")

	// Set the duration for how long the
	// process should run (1 hour)
	// 	endTime := time.Now().Add(time.Hour + time.Minute)

	// just for check
	endTime := time.Now().Add(3 * time.Minute)

	iterationCount := 0

	// mappingTypes := []MappingType{EGH, OLS}

	for time.Now().Before(endTime) {
		iterationCount++

		ctx := context.Background()
		txpool1Data, err := fetchTxPoolContent(node1, ctx)
		if err != nil {
			log.Printf("Failed to fetch txpool content for Node 1: %v", err)
			continue
		}

		txpool2Data, err := fetchTxPoolContent(node2, ctx)
		if err != nil {
			log.Printf("Failed to fetch txpool content for Node 2: %v", err)
			continue
		}

		hashes1 := getTransactionsHashes(txpool1Data)
		hashes2 := getTransactionsHashes(txpool2Data)

		// err = saveHashesToCSV(txpool1Data, "node1", node1Dir, iterationCount)
		// if err != nil {
		// 	log.Printf("Error saving Node 1 hashes to CSV: %v", err)
		// }

		err = saveTransactionStatsToCSV(txpool1Data, iterationCount, node1Dir)
		if err != nil {
			log.Printf("Error saving Node 1 stats to CSV: %v", err)
		}

		// err = saveHashesToCSV(txpool2Data, "node2", node2Dir, iterationCount)
		// if err != nil {
		// 	log.Printf("Error saving Node 2 hashes to CSV: %v", err)
		// }

		err = saveTransactionStatsToCSV(txpool2Data, iterationCount, node2Dir)
		if err != nil {
			log.Printf("Error saving Node 2 stats to CSV: %v", err)
		}

		// relevant only for egh for now (ols universe reduction)
		universeSize := uint256.NewInt(0).SetAllOne()

		// TODO - fix later to support other mapping types.
		symDiffSize, totalCells := certainSync(hashes1, hashes2, universeSize, EGH)
		fmt.Printf("Iteration %d: Symmetric Difference: %d\n", iterationCount, symDiffSize)

		err = saveSymmetricDiffStatsToCSV(symmetricDiffStatsFilePath, iterationCount, uint64(symDiffSize), totalCells)
		if err != nil {
			log.Printf("Error saving symmetric difference stats to CSV: %v", err)
		}

		// time.Sleep(10 * time.Second)
		time.Sleep(time.Minute)
	}
}

func main() {
	// txpool_sync()

	// txpool_sync_from_nodes_certain_sync()

	// txpool_sync_from_nodes_universe_reduce_sync()

	// txpool_sync_from_file_certain_sync()

	txpool_sync_from_file_universe_reduce_sync()
}
