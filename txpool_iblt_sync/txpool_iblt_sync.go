package main

import (
	"context"
	"encoding/csv"
	"encoding/json"
	"fmt"
	"log"
	"os"
	"path/filepath"
	"time"

	"github.com/ethereum/go-ethereum/rpc"
	. "github.com/toto9820/Rateless-Set-Reconciliation-with-Listing-Guarantees/certainsync"
)

// Config represents the structure of the configuration file.
type Config struct {
	Node1IPC string `json:"node1_ipc"`
	Node2IPC string `json:"node2_ipc"`
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

// compareIBFs generates IBFs for two sets of
// transaction hashes, compares them, and finds the
// symmetric difference.
func compareIBFs(hashes1, hashes2 []Symbol, initialCells uint64) (int, uint64) {
	ibfNode1 := NewIBF(initialCells, "hash", &EGHMapping{})
	ibfNode2 := NewIBF(initialCells, "hash", &EGHMapping{})

	for {
		ibfNode1.AddSymbols(hashes1)
		ibfNode2.AddSymbols(hashes2)

		// Subtract the two IBFs
		ibfDiff := ibfNode1.Subtract(ibfNode2)
		symmetricDiff, ok := ibfDiff.Decode()

		if ok {
			return len(symmetricDiff), ibfDiff.Size
		}
	}
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

func main() {
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

		hashes1 := getTransactionHashes(txpool1Data)
		hashes2 := getTransactionHashes(txpool2Data)

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

		symDiffSize, totalCells := compareIBFs(hashes1, hashes2, 1000)
		fmt.Printf("Iteration %d: Symmetric Difference: %d\n", iterationCount, symDiffSize)

		err = saveSymmetricDiffStatsToCSV(symmetricDiffStatsFilePath, iterationCount, uint64(symDiffSize), totalCells)
		if err != nil {
			log.Printf("Error saving symmetric difference stats to CSV: %v", err)
		}

		// time.Sleep(10 * time.Second)
		time.Sleep(time.Minute)
	}
}
