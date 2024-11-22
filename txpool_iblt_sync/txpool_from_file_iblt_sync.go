package main

import (
	"fmt"
	"log"
	"os"
	"path/filepath"

	"github.com/holiman/uint256"
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

	symmetricDiffStatsFilePath := filepath.Join(cwd, "data", "blockchain", "egh_file_symmetric_diff_stats.csv")

	iterationCount := 0
	maxIterations := 10
	universeSize := uint256.NewInt(0).SetAllOne()

	for i := 0; i < maxIterations; i++ {
		iterationCount++

		node1HashesFilePath := filepath.Join(node1Dir, fmt.Sprintf("node1_txpool_hashes_%d.csv", iterationCount))
		node2HashesFilePath := filepath.Join(node2Dir, fmt.Sprintf("node2_txpool_hashes_%d.csv", iterationCount))

		hashes1 := getTransactionsHashesFromFile(node1HashesFilePath)
		hashes2 := getTransactionsHashesFromFile(node2HashesFilePath)

		symDiffSize, totalCells := compareIBFs(hashes1, hashes2, universeSize)
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

	symmetricDiffStatsFilePath := filepath.Join(cwd, "data", "blockchain", "ols_file_symmetric_diff_stats.csv")

	iterationCount := 0
	maxIterations := 10

	// universeSizes := []*uint256.Int{
	// 	uint256.NewInt(uint64(math.Pow(2, 14))),
	// 	// uint256.NewInt(uint64(math.Pow(2, 16))),
	// 	// uint256.NewInt(uint64(math.Pow(2, 32))),
	// 	// uint256.NewInt(uint64(math.Pow(2, 64))),
	// }

	// deltaSizes := []float64{
	// 	0.5,
	// 	0.1,
	// 	0.01,
	// 	0.01,
	// }

	deltaSizes := []float64{
		100,
		10,
		1,
	}

	for i := 0; i < maxIterations; i++ {
		iterationCount++

		node1HashesFilePath := filepath.Join(node1Dir, fmt.Sprintf("node1_txpool_hashes_%d.csv", 2))
		node2HashesFilePath := filepath.Join(node2Dir, fmt.Sprintf("node2_txpool_hashes_%d.csv", 2))

		hashes1 := getTransactionsHashesFromFile(node1HashesFilePath)
		hashes2 := getTransactionsHashesFromFile(node2HashesFilePath)

		for _, deltaSize := range deltaSizes {
			symDiffSize, totalCells := compareIBFsExtended(hashes1, hashes2, deltaSize)
			fmt.Printf("Iteration %d, Delta Size %f: Symmetric Difference: %d, Total Cells: %d\n", iterationCount, deltaSize, symDiffSize, totalCells)

			err = saveSymmetricDiffStatsToCSV(symmetricDiffStatsFilePath, iterationCount, uint64(symDiffSize), totalCells)
			if err != nil {
				log.Printf("Error saving symmetric difference stats to CSV: %v", err)
			}
		}
	}
}
