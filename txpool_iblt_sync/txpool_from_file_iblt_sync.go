package main

import (
	"fmt"
	"log"
	"os"
	"path/filepath"

	"github.com/holiman/uint256"
)

func txpool_sync_from_file_certain_sync() {
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

	maxIterations := 15
	universeSize := uint256.NewInt(0).SetAllOne()

	mappingTypes := []MappingType{EGH}

	for _, mappingType := range mappingTypes {
		symmetricDiffStatsFilePath := filepath.Join(cwd, "data", "blockchain", fmt.Sprintf("%s_certain_sync_file_symmetric_diff_stats.csv", mappingType))

		for i := 0; i < maxIterations; i++ {
			iterationCount := i + 1

			node1HashesFilePath := filepath.Join(node1Dir, fmt.Sprintf("node1_txpool_hashes_%d.csv", iterationCount))
			node2HashesFilePath := filepath.Join(node2Dir, fmt.Sprintf("node2_txpool_hashes_%d.csv", iterationCount))

			hashes1 := getTransactionsHashesFromFile(node1HashesFilePath)
			hashes2 := getTransactionsHashesFromFile(node2HashesFilePath)

			symDiffSize, totalCells := certainSync(hashes1, hashes2, universeSize, mappingType)
			fmt.Printf("MappingType %s, Iteration %d: Symmetric Difference: %d\n", mappingType, iterationCount, symDiffSize)

			err = saveSymmetricDiffStatsToCSV(symmetricDiffStatsFilePath, iterationCount, uint64(symDiffSize), totalCells)
			if err != nil {
				log.Printf("Error saving symmetric difference stats to CSV: %v", err)
			}
		}
	}
}

func txpool_sync_from_file_universe_reduce_sync() {
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

	maxIterations := 15

	mappingTypes := []MappingType{EGH, OLS}

	// delta - max collisions allowed.
	deltaSizes := []float64{
		100,
		10,
		1,
	}

	for i := 0; i < maxIterations; i++ {
		iterationCount := i + 1

		node1HashesFilePath := filepath.Join(node1Dir, fmt.Sprintf("node1_txpool_hashes_%d.csv", iterationCount))
		node2HashesFilePath := filepath.Join(node2Dir, fmt.Sprintf("node2_txpool_hashes_%d.csv", iterationCount))

		hashes1 := getTransactionsHashesFromFile(node1HashesFilePath)
		hashes2 := getTransactionsHashesFromFile(node2HashesFilePath)

		for _, mappingType := range mappingTypes {
			for _, deltaSize := range deltaSizes {
				symmetricDiffStatsFilePath := filepath.Join(cwd, "data", "blockchain", fmt.Sprintf("%s_universe_reduce_sync_file_symmetric_diff_stats_delta_%d.csv", mappingType, uint64(deltaSize)))

				symDiffSize, totalTransmittedBits := universeReduceSync(hashes1, hashes2, deltaSize, mappingType)
				fmt.Printf("MappingType %s, Iteration %d, Delta Size %d: Symmetric Difference: %d, Total Transmitted Bits: %d\n", mappingType, iterationCount, uint64(deltaSize), symDiffSize, totalTransmittedBits)

				err = saveSymmetricDiffStatsToCSV(symmetricDiffStatsFilePath, iterationCount, uint64(symDiffSize), totalTransmittedBits)
				if err != nil {
					log.Printf("Error saving symmetric difference stats to CSV: %v", err)
				}
			}
		}
	}
}
