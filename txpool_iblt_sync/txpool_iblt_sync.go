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

	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/rpc"
	. "github.com/toto9820/Rateless-Set-Reconciliation-with-Listing-Guarantees/riblt_with_certainty"
)

// saveHashesToCSV saves the transaction hashes (pending and queued) to a CSV file
func saveHashesToCSV(txpoolData TxPoolContent, nodeName string, dirPath string, timestamp int) error {
	// Create a file path using the node name and timestamp
	filePath := filepath.Join(dirPath, fmt.Sprintf("%s_txpool_hashes_%d.csv", nodeName, timestamp))

	// Create the CSV file
	file, err := os.Create(filePath)
	if err != nil {
		return fmt.Errorf("error creating file: %v", err)
	}
	defer file.Close()

	writer := csv.NewWriter(file)
	defer writer.Flush()

	// // Write the header row to the CSV file.
	// writer.Write([]string{"Hashes"})

	// Write pending transactions to CSV
	for _, txs := range txpoolData.Pending {
		for _, tx := range txs {
			record := []string{tx.Hash.String()}
			err := writer.Write(record)
			if err != nil {
				return fmt.Errorf("failed to write record to CSV: %v", err)
			}
		}
	}

	// Write queued transactions to CSV
	for _, txs := range txpoolData.Queued {
		for _, tx := range txs {
			record := []string{tx.Hash.String()}
			err := writer.Write(record)
			if err != nil {
				return fmt.Errorf("failed to write record to CSV: %v", err)
			}
		}
	}

	return nil
}

// getTransactionHashes extracts transaction hashes from the txpool data
func getTransactionHashes(txpoolData TxPoolContent) []Symbol {
	var hashes []Symbol

	// Collect pending transaction hashes
	for _, txs := range txpoolData.Pending {
		for _, tx := range txs {
			hashes = append(hashes, HashSymbol(tx.Hash))
		}
	}

	// Collect queued transaction hashes
	for _, txs := range txpoolData.Queued {
		for _, tx := range txs {
			hashes = append(hashes, HashSymbol(tx.Hash))
		}
	}

	return hashes
}

// Define a structure for the txpool content response
type TxPoolContent struct {
	Pending map[string]map[string]Transaction `json:"pending"`
	Queued  map[string]map[string]Transaction `json:"queued"`
}

// Define a structure for the transaction
type Transaction struct {
	// Hash string `json:"hash"`
	Hash common.Hash `json:"hash"`
}

// Define a structure to hold the config
type Config struct {
	Node1IPC string `json:"node1_ipc"`
	Node2IPC string `json:"node2_ipc"`
}

// Function to load config from a JSON file
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

func main() {
	// Connect to Ethereum clients using IPC
	// Load the config file
	config, err := loadConfig("config.json")
	if err != nil {
		log.Fatalf("Failed to load config: %v", err)
	}

	// Connect to Ethereum clients using IPC
	node1, err := rpc.Dial(config.Node1IPC)
	if err != nil {
		log.Fatalf("Failed to connect to Node 1 Ethereum client: %v", err)
	}

	node2, err := rpc.Dial(config.Node2IPC)
	if err != nil {
		log.Fatalf("Failed to connect to Node 2 Ethereum client: %v", err)
	}

	// Get the current working directory
	cwd, err := os.Getwd()
	if err != nil {
		log.Fatalf("Failed to get current working directory: %v", err)
	}

	// Set the directories for saving CSV files
	node1Dir := filepath.Join(cwd, "data", "blockchain", "node1")
	node2Dir := filepath.Join(cwd, "data", "blockchain", "node2")

	// Create the directories if they don't exist
	err = os.MkdirAll(node1Dir, os.ModePerm)
	if err != nil {
		log.Fatalf("Failed to create directory for Node 1: %v", err)
	}
	err = os.MkdirAll(node2Dir, os.ModePerm)
	if err != nil {
		log.Fatalf("Failed to create directory for Node 2: %v", err)
	}

	// Set the duration for how long the process should run (15 minutes)
	endTime := time.Now().Add(15 * time.Minute)

	// just for check
	// endTime := time.Now().Add(3 * time.Minute)

	iteration_count := 1

	for time.Now().Before(endTime) {
		// timestamp := iteration_count
		iteration_count += 1

		var hashes1 []Symbol
		var hashes2 []Symbol
		var symmetricDiffSize int

		// Get transaction pool content for node 1
		var txpool1Data TxPoolContent
		ctx := context.Background()
		err = node1.CallContext(ctx, &txpool1Data, "txpool_content")
		if err != nil {
			log.Printf("Failed to fetch txpool content for Node 1: %v", err)
		} else {
			// Get all transaction hashes for Node 1
			hashes1 = getTransactionHashes(txpool1Data)
			fmt.Printf("Node 1 Transaction Hashes length: %v\n", len(hashes1))

			// // Save the data to CSV for Node 1
			// err = saveHashesToCSV(txpool1Data, "node1", node1Dir, timestamp)
			// if err != nil {
			// 	log.Printf("Error saving Node 1 hashes to CSV: %v", err)
			// }
		}

		// Get transaction pool content for node 2
		var txpool2Data TxPoolContent
		err = node2.CallContext(ctx, &txpool2Data, "txpool_content")
		if err != nil {
			log.Printf("Failed to fetch txpool content for Node 2: %v", err)
		} else {
			// Get all transaction hashes for Node 2
			hashes2 = getTransactionHashes(txpool2Data)
			fmt.Printf("Node 2 Transaction Hashes length: %v\n", len(hashes2))

			// // Save the data to CSV for Node 2
			// err = saveHashesToCSV(txpool2Data, "node2", node2Dir, timestamp)
			// if err != nil {
			// 	log.Printf("Error saving Node 2 hashes to CSV: %v", err)
			// }
		}

		intialCells := uint64(1000)

		ibfNode1 := NewIBF(intialCells, "hash")
		ibfNode2 := NewIBF(intialCells, "hash")
		cost := uint64(0)

		for {
			ibfNode1.AddSymbols(hashes1)
			ibfNode2.AddSymbols(hashes2)

			// Subtract the two IBFs and Decode the result to find the differences
			ibfDiff := ibfNode1.Subtract(ibfNode2)
			symmetricDiff, ok := ibfDiff.Decode()

			if ok == false {
				continue
			}

			symmetricDiffSize = len(symmetricDiff)

			cost = ibfDiff.Size
			break
		}

		fmt.Printf("EGH method for CertainSync IBLT, Symmetric Difference len: %d with %d cells", symmetricDiffSize, cost)
		fmt.Println()

		// Wait for 1 minute before the next iteration
		time.Sleep(1 * time.Minute)
	}
}
