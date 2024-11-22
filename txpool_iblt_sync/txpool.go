package main

import (
	"context"
	"encoding/csv"
	"encoding/hex"
	"fmt"
	"log"
	"os"
	"path/filepath"
	"strings"

	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/rpc"
	"github.com/holiman/uint256"
)

type MappingType string

const (
	EGH MappingType = "egh"
	OLS MappingType = "ols"
)

// TxPoolContent represents the structure of the
// Ethereum transaction pool content.
type TxPoolContent struct {
	Pending map[string]map[string]Transaction `json:"pending"`
	Queued  map[string]map[string]Transaction `json:"queued"`
}

// Transaction represents a transaction in the
// Ethereum pool with its hash.
type Transaction struct {
	Hash common.Hash `json:"hash"`
}

// fetchTxPoolContent fetches the transaction pool content
// from an Ethereum node.
func fetchTxPoolContent(client *rpc.Client, ctx context.Context) (TxPoolContent, error) {
	var txpoolData TxPoolContent
	err := client.CallContext(ctx, &txpoolData, "txpool_content")
	return txpoolData, err
}

// getTransactionHashes extracts transaction hashes
// from the txpool data.
func getTransactionsHashes(txpoolData TxPoolContent) []*uint256.Int {
	var hashes []*uint256.Int

	// Collect pending transaction hashes
	for _, txs := range txpoolData.Pending {
		for _, tx := range txs {
			hashes = append(hashes, uint256.NewInt(0).SetBytes(tx.Hash[:]))
		}
	}

	// Collect queued transaction hashes
	for _, txs := range txpoolData.Queued {
		for _, tx := range txs {
			hashes = append(hashes, uint256.NewInt(0).SetBytes(tx.Hash[:]))
		}
	}

	return hashes
}

func getTransactionsHashesFromFile(hashesFilePath string) []*uint256.Int {
	// var hashes []Symbol
	var hashes []*uint256.Int

	// Open the CSV file
	file, err := os.Open(hashesFilePath)
	if err != nil {
		log.Fatalf("Failed to open file: %s", err)
	}
	defer file.Close()

	// Create a CSV reader
	reader := csv.NewReader(file)

	// Read all lines from the CSV
	records, err := reader.ReadAll()
	if err != nil {
		log.Fatalf("Failed to read CSV: %s", err)
	}

	for _, record := range records {
		if len(record) > 0 {
			// Parse the hash string from the record
			// (one hash per line)
			hashStr := record[0]

			// Remove 0x prefix if present
			hashStr = strings.TrimPrefix(hashStr, "0x")

			// Decode hex string to bytes
			hashBytes, err := hex.DecodeString(hashStr)

			if err != nil {
				panic(err)
			}

			// Convert to common.Hash
			// hash := common.HexToHash(hashStr)
			hash := uint256.NewInt(0).SetBytes(hashBytes)

			// Append to the slice
			hashes = append(hashes, hash)
		}
	}

	return hashes
}

// saveHashesToCSV saves the transaction hashes to a CSV file.
func saveHashesToCSV(txpoolData TxPoolContent, nodeName string, dirPath string, timestamp int) error {
	filePath := filepath.Join(dirPath, fmt.Sprintf("%s_txpool_hashes_%d.csv", nodeName, timestamp))

	file, err := os.Create(filePath)
	if err != nil {
		return fmt.Errorf("error creating file: %v", err)
	}
	defer file.Close()

	writer := csv.NewWriter(file)
	defer writer.Flush()

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

// saveTransactionStatsToCSV saves transaction pool
// statistics (total, queued, pending) to CSV.
func saveTransactionStatsToCSV(txpoolData TxPoolContent, iteration int, dirPath string) error {
	totalTx, queuedTx, pendingTx := countTransactions(txpoolData)

	// Helper function to write the transaction count to the file
	writeCountToFile := func(filePath string, count int) error {
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
			header := []string{"Time (minutes)", "Total Transactions"}
			if err := writer.Write(header); err != nil {
				return err
			}
		}

		record := []string{fmt.Sprintf("%d", iteration), fmt.Sprintf("%d", count)}
		if err := writer.Write(record); err != nil {
			return fmt.Errorf("error writing record to %s: %v", filePath, err)
		}
		return nil
	}

	totalPath := filepath.Join(dirPath, "total_tx.csv")
	queuedPath := filepath.Join(dirPath, "total_queued_tx.csv")
	pendingPath := filepath.Join(dirPath, "total_pending_tx.csv")

	if err := writeCountToFile(totalPath, totalTx); err != nil {
		return err
	}
	if err := writeCountToFile(queuedPath, queuedTx); err != nil {
		return err
	}
	if err := writeCountToFile(pendingPath, pendingTx); err != nil {
		return err
	}

	return nil
}

// countTransactions counts total, queued, and pending transactions.
func countTransactions(txpoolData TxPoolContent) (int, int, int) {
	var totalTx, queuedTx, pendingTx int

	// Count pending transactions
	for _, txs := range txpoolData.Pending {
		pendingTx += len(txs)
	}

	// Count queued transactions
	for _, txs := range txpoolData.Queued {
		queuedTx += len(txs)
	}

	totalTx = pendingTx + queuedTx
	return totalTx, queuedTx, pendingTx
}
