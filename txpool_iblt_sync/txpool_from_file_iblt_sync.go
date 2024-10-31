package main

import (
	"fmt"
	"log"
	"math"
	"os"
	"path/filepath"
)

func txpool_sync_from_file_egh() {
	cwd, err := os.Getwd()
	if err != nil {
		log.Fatalf("Failed to get current working directory: %v", err)
	}

	configPath := filepath.Join(cwd, "Configuration", "config.json")
	config, err := loadConfig(configPath)
	if err != nil {
		log.Fatalf("Failed to load config: %v", err)
	}

	node1Dir := config.Node1HashesDir
	node2Dir := config.Node2HashesDir

	symmetricDiffStatsFilePath := filepath.Join(cwd, "data", "blockchain", "file_symmetric_diff_stats.csv")

	iterationCount := 0
	maxIterations := 10
	initialCells := uint64(1000)

	for i := 0; i < maxIterations; i++ {
		iterationCount++

		node1HashesFilePath := filepath.Join(node1Dir, fmt.Sprintf("node1_txpool_hashes_%d.csv", iterationCount))
		node2HashesFilePath := filepath.Join(node2Dir, fmt.Sprintf("node2_txpool_hashes_%d.csv", iterationCount))

		hashes1 := getTransactionsHashesFromFile(node1HashesFilePath)
		hashes2 := getTransactionsHashesFromFile(node2HashesFilePath)

		symDiffSize, totalCells := compareIBFs(hashes1, hashes2, initialCells)
		fmt.Printf("Iteration %d: Symmetric Difference: %d\n", iterationCount, symDiffSize)

		err = saveSymmetricDiffStatsToCSV(symmetricDiffStatsFilePath, iterationCount, uint64(symDiffSize), totalCells)
		if err != nil {
			log.Printf("Error saving symmetric difference stats to CSV: %v", err)
		}
	}
}

func txpool_sync_from_file_ols() {
	cwd, err := os.Getwd()
	if err != nil {
		log.Fatalf("Failed to get current working directory: %v", err)
	}

	configPath := filepath.Join(cwd, "Configuration", "config.json")
	config, err := loadConfig(configPath)
	if err != nil {
		log.Fatalf("Failed to load config: %v", err)
	}

	node1Dir := config.Node1HashesDir
	node2Dir := config.Node2HashesDir

	symmetricDiffStatsFilePath := filepath.Join(cwd, "data", "blockchain", "file_symmetric_diff_stats.csv")

	iterationCount := 0
	maxIterations := 10
	initialCells := uint64(1000)
	universeSize := uint64(math.Pow(2, 32))

	for i := 0; i < maxIterations; i++ {
		iterationCount++

		node1HashesFilePath := filepath.Join(node1Dir, fmt.Sprintf("node1_txpool_hashes_%d.csv", iterationCount))
		node2HashesFilePath := filepath.Join(node2Dir, fmt.Sprintf("node2_txpool_hashes_%d.csv", iterationCount))

		hashes1 := getTransactionsHashesFromFile(node1HashesFilePath)
		hashes2 := getTransactionsHashesFromFile(node2HashesFilePath)

		symDiffSize, totalCells := compareIBFsExtended(hashes1, hashes2, initialCells, universeSize)
		fmt.Printf("Iteration %d: Symmetric Difference: %d\n", iterationCount, symDiffSize)

		err = saveSymmetricDiffStatsToCSV(symmetricDiffStatsFilePath, iterationCount, uint64(symDiffSize), totalCells)
		if err != nil {
			log.Printf("Error saving symmetric difference stats to CSV: %v", err)
		}
	}
}
