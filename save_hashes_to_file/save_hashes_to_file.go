package main

import (
	"context"
	"encoding/csv"
	"fmt"
	"log"
	"os"
	"path/filepath"
	"time"

	"github.com/ethereum/go-ethereum/rpc"
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
			record := []string{tx.Hash}
			err := writer.Write(record)
			if err != nil {
				return fmt.Errorf("failed to write record to CSV: %v", err)
			}
		}
	}

	// Write queued transactions to CSV
	for _, txs := range txpoolData.Queued {
		for _, tx := range txs {
			record := []string{tx.Hash}
			err := writer.Write(record)
			if err != nil {
				return fmt.Errorf("failed to write record to CSV: %v", err)
			}
		}
	}

	return nil
}

// Define a structure for the txpool content response
type TxPoolContent struct {
	Pending map[string]map[string]Transaction `json:"pending"`
	Queued  map[string]map[string]Transaction `json:"queued"`
}

// Define a structure for the transaction
type Transaction struct {
	// Hash string `json:"hash"`
	Hash string `json:"hash"`
}

func main() {
	// Connect to Ethereum clients using IPC
	node1, err := rpc.Dial("/data/data_Tomer/node1/geth.ipc") // Using IPC connection for Node 1
	if err != nil {
		log.Fatalf("Failed to connect to Node 1 Ethereum client: %v", err)
	}

	node2, err := rpc.Dial("/data/data_Tomer/node2/geth.ipc") // Using IPC connection for Node 2
	if err != nil {
		log.Fatalf("Failed to connect to Node 2 Ethereum client: %v", err)
	}

	// Set the directories for saving CSV files
	node1Dir := "/home/tomer_local/Rateless-Set-Reconciliation-with-Listing-Guarantees/data/blockchain/node1"
	node2Dir := "/home/tomer_local/Rateless-Set-Reconciliation-with-Listing-Guarantees/data/blockchain/node2"

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
		timestamp := iteration_count
		iteration_count += 1

		// Get transaction pool content for node 1
		var txpool1Data TxPoolContent
		ctx := context.Background()
		err = node1.CallContext(ctx, &txpool1Data, "txpool_content")
		if err != nil {
			log.Printf("Failed to fetch txpool content for Node 1: %v", err)
		} else {
			// Save the data to CSV for Node 1
			err = saveHashesToCSV(txpool1Data, "node1", node1Dir, timestamp)
			if err != nil {
				log.Printf("Error saving Node 1 hashes to CSV: %v", err)
			}
		}

		// Get transaction pool content for node 2
		var txpool2Data TxPoolContent
		err = node2.CallContext(ctx, &txpool2Data, "txpool_content")
		if err != nil {
			log.Printf("Failed to fetch txpool content for Node 2: %v", err)
		} else {
			// Save the data to CSV for Node 2
			err = saveHashesToCSV(txpool2Data, "node2", node2Dir, timestamp)
			if err != nil {
				log.Printf("Error saving Node 2 hashes to CSV: %v", err)
			}
		}

		// Wait for 1 minute before the next iteration
		time.Sleep(1 * time.Minute)
	}
}
