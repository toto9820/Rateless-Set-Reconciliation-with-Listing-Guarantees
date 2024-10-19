// package main

// import (
// 	"encoding/csv"
// 	"fmt"
// 	"log"
// 	"os"
// 	"strconv"
// )

// var hashMapping map[string]int

// // Counter to assign unique IDs to hashes
// var uniqueCount int

// func processNode(nodeNumber int) {
// 	// Initialize the Cuckoo Filter

// 	// cf = cuckoo.NewScalableCuckooFilter()
// 	// cf := cuckoo.NewFilter(6000)

// 	// Create a map to store the hash to index mappings
// 	// hashMapping := make(map[string]int)

// 	// Counter to assign unique IDs to hashes
// 	// idCounter := 0

// 	// Loop through CSV files from 1 to 15
// 	for i := 1; i <= 15; i++ {
// 		// Build the file path for each CSV file
// 		filePath := fmt.Sprintf("C:\\Users\\Tomer_Keniagin\\Desktop\\עתודה - טכניון\\Thesis\\Code\\Rateless-Set-Reconciliation-with-Listing-Guarantees\\data\\blockchain\\node%d\\node%d_txpool_hashes_%d.csv", nodeNumber, nodeNumber, i)

// 		// Open the CSV file
// 		file, err := os.Open(filePath)
// 		if err != nil {
// 			log.Fatalf("Failed to open file %s: %v", filePath, err)
// 		}
// 		defer file.Close()

// 		// Create a CSV reader
// 		reader := csv.NewReader(file)

// 		// Read the CSV file line by line
// 		records, err := reader.ReadAll()
// 		if err != nil {
// 			log.Fatalf("Failed to read file %s: %v", filePath, err)
// 		}

// 		for _, record := range records {
// 			if len(record) > 0 {
// 				txHash := record[0]

// 				if _, exists := hashMapping[txHash]; !exists {
// 					hashMapping[txHash] = uniqueCount
// 					uniqueCount++
// 				}
// 			}
// 		}

// 		// Write the new mappings to a CSV file
// 		outputFilePath := fmt.Sprintf("C:\\Users\\Tomer_Keniagin\\Desktop\\עתודה - טכניון\\Thesis\\Code\\Rateless-Set-Reconciliation-with-Listing-Guarantees\\data\\blockchain\\node%d\\node%d_txpool_hashes_mapped_%d.csv", nodeNumber, nodeNumber, i)

// 		outputFile, err := os.Create(outputFilePath)
// 		if err != nil {
// 			log.Fatalf("Failed to create output file: %v", err)
// 		}
// 		defer outputFile.Close()

// 		writer := csv.NewWriter(outputFile)
// 		// defer writer.Flush()

// 		// Write the mappings to the CSV file
// 		for _, record := range records {
// 			if len(record) > 0 {
// 				txHash := record[0]

// 				count := hashMapping[txHash]

// 				err := writer.Write([]string{strconv.Itoa(count)})
// 				if err != nil {
// 					log.Fatalf("Error writing to CSV: %v", err)
// 				}
// 			}
// 		}

// 		writer.Flush()
// 	}
// }

// func main() {
// 	hashMapping = make(map[string]int)
// 	uniqueCount = 0

// 	// Process nodes 1, 2.
// 	for nodeNumber := 1; nodeNumber <= 2; nodeNumber++ {
// 		fmt.Printf("Processing node %d\n", nodeNumber)
// 		processNode(nodeNumber)
// 	}
// }

package main

import (
	"container/list"
	"fmt"
	"sync"

	"github.com/ethereum/go-ethereum/common"
)

const (
	// In order to determine constant universe size which
	// is MaxMappings.
	MaxMappings      = 100000
	RemovalBatchSize = 10000
)

// LimitedHashMapper provides methods to map Ethereum transaction hashes to uint64 identifiers
type LimitedHashMapper struct {
	counter   uint64
	hashToID  map[common.Hash]uint64
	orderList *list.List
	mutex     sync.RWMutex
}

// NewLimitedHashMapper creates a new LimitedHashMapper
func NewLimitedHashMapper() *LimitedHashMapper {
	return &LimitedHashMapper{
		counter:      1,
		hashToID:     make(map[common.Hash]uint64),
		orderList:    list.New(),
		availableIDs: list.New(),
	}
}

// MapToUint64 maps an Ethereum transaction hash
// to a uint64 identifier
func (lhm *LimitedHashMapper) MapToUint64(txHash common.Hash) uint64 {
	lhm.mutex.Lock()
	defer lhm.mutex.Unlock()

	if id, exists := lhm.hashToID[txHash]; exists {
		lhm.moveToFront(txHash)
		return id
	}

	if lhm.orderList.Len() >= MaxMappings {
		lhm.removeOldestBatch()
	}

	var id uint64

	if lhm.availableIDs.Len() > 0 {
		id = lhm.availableIDs.Remove(lhm.availableIDs.Front()).(uint64)
	} else {
		id = lhm.counter
		lhm.counter++
		if lhm.counter == 0 { // Handle uint64 overflow
			lhm.counter = 1
		}
	}

	return id
}

// removeOldestBatch removes the oldest RemovalBatchSize
// mappings when the limit is reached
func (lhm *LimitedHashMapper) removeOldestBatch() {
	for i := 0; i < RemovalBatchSize && lhm.orderList.Len() > 0; i++ {
		oldest := lhm.orderList.Back()
		if oldest != nil {
			lhm.orderList.Remove(oldest)
			oldestHash := oldest.Value.(common.Hash)
			id := lhm.hashToID[oldestHash]
			delete(lhm.hashToID, oldestHash)
			// Recycle the ID
			lhm.availableIDs.PushBack(id)
		}
	}
}

// moveToFront moves a recently accessed hash to the
// front of the list
func (lhm *LimitedHashMapper) moveToFront(hash common.Hash) {
	for e := lhm.orderList.Front(); e != nil; e = e.Next() {
		if e.Value.(common.Hash) == hash {
			lhm.orderList.MoveToFront(e)
			break
		}
	}
}

// Size returns the current number of mappings
func (lhm *LimitedHashMapper) Size() int {
	lhm.mutex.RLock()
	defer lhm.mutex.RUnlock()
	return lhm.orderList.Len()
}

func main() {
	mapper := NewLimitedHashMapper()

	// Example usage
	txHash1 := common.HexToHash("0x1234567890abcdef1234567890abcdef1234567890abcdef1234567890abcdef")
	txHash2 := common.HexToHash("0xfedcba0987654321fedcba0987654321fedcba0987654321fedcba0987654321")

	id1 := mapper.MapToUint64(txHash1)
	id2 := mapper.MapToUint64(txHash2)
	id1Again := mapper.MapToUint64(txHash1)

	fmt.Printf("Transaction Hash 1: %s\n", txHash1.Hex())
	fmt.Printf("Mapped to uint64: %d\n\n", id1)

	fmt.Printf("Transaction Hash 2: %s\n", txHash2.Hex())
	fmt.Printf("Mapped to uint64: %d\n\n", id2)

	fmt.Printf("Transaction Hash 1 (again): %s\n", txHash1.Hex())
	fmt.Printf("Mapped to uint64: %d\n\n", id1Again)

	// Simulate reaching the mapping limit
	for i := 0; i < MaxMappings; i++ {
		dummyHash := common.BytesToHash([]byte(fmt.Sprintf("dummy%d", i)))
		mapper.MapToUint64(dummyHash)
	}

	fmt.Printf("Current number of mappings: %d\n\n", mapper.Size())

	// Try mapping txHash1 again after reaching the limit
	id1AfterLimit := mapper.MapToUint64(txHash1)
	fmt.Printf("Transaction Hash 1 (after reaching limit): %s\n", txHash1.Hex())
	fmt.Printf("Mapped to uint64: %d\n", id1AfterLimit)
}
