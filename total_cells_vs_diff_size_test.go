package riblt_with_certainty

import (
	"encoding/csv"
	"fmt"
	"log"
	"math"
	"math/rand"
	"os"
	"sort"
	"testing"
	"time"
)

// runTrial simulates a reconciliation trial for benchmarking.
func runTrialTotalCellsVsDiffSize(trialNumber int,
	universeSize int,
	symmetricDiffSize int,
	rng *rand.Rand) uint64 {
	// For superset assumption
	// Bob's set will include all elements from 1 to universeSize.
	bob := make([]uint64, 0, universeSize)
	for i := 1; i <= universeSize; i++ {
		bob = append(bob, uint64(i))
	}

	// Alice's set will include universeSize - symmetricDiffSize elements.
	alice := make([]uint64, 0, universeSize-symmetricDiffSize)

	// Randomly choose indices from Bob's set to include in Alice's set.
	chosenIndices := rng.Perm(universeSize)[:universeSize-symmetricDiffSize] // Random permutation.
	for _, idx := range chosenIndices {
		alice = append(alice, bob[idx]) // idx is within 0 to universeSize-1
	}

	// // Sort Alice's set.
	sort.Slice(alice, func(i, j int) bool {
		return alice[i] < alice[j]
	})

	intialCells := uint64(1000)

	ibfAlice := NewIBF(intialCells)
	ibfBob := NewIBF(intialCells)
	cost := uint64(0)

	for {
		ibfAlice.AddSymbols(alice)
		ibfBob.AddSymbols(bob)

		// Subtract the two IBFs and Decode the result to find the differences
		ibfDiff := ibfBob.Subtract(ibfAlice)
		bobWithoutAlice, _, ok := ibfDiff.Decode()

		if ok == false {
			continue
		}

		if len(bobWithoutAlice) == symmetricDiffSize {
			cost = ibfDiff.Size
			break
		}
	}

	fmt.Printf("Trial %d for EGH method for Rateless LFFZ IBLT, Symmetric Difference len: %d with %d cells", trialNumber, symmetricDiffSize, cost)
	fmt.Println()

	log.Printf("Trial %d for EGH method for Rateless LFFZ IBLT, Symmetric Difference len: %d with %d cells\n", trialNumber, symmetricDiffSize, cost)

	// Return number of coded symbols transmitted
	return cost
}

// BenchmarkReconciliation benchmarks the reconciliation
// process with a fixed universe size and different
// symmetric difference sizes.
func BenchmarkTotalCellsVsDiffSize(b *testing.B) {
	benches := []struct {
		symmetricDiffSize int
	}{
		{1},
		{10},
		{100},
		{1000},
		{10000},
	}

	// benches := []struct {
	// 	symmetricDiffSize int
	// }{
	// 	// {1},
	// 	{10},
	// }

	// Create a local random number generator with a time-based seed
	rng := rand.New(rand.NewSource(time.Now().UnixNano()))

	// Prepare a CSV file to store the results.
	file, err := os.Create("egh_total_cells_vs_diff_size_set_inside_set.csv")
	if err != nil {
		fmt.Println("Error creating file:", err)
		return
	}
	defer file.Close()

	writer := csv.NewWriter(file)
	defer writer.Flush()

	// Write the header row to the CSV file.
	writer.Write([]string{"Symmetric Diff Size", "Total Cells Transmitted"})

	// Set the number of trials
	// numTrials := 100
	numTrials := 10

	// universeSize := int(math.Pow(10, 2))
	universeSize := int(math.Pow(10, 6))

	for _, bench := range benches {
		var totalCellsTransmitted uint64
		b.Run(fmt.Sprintf("Universe=%d, Diff=%d", universeSize, bench.symmetricDiffSize), func(b *testing.B) {
			totalCellsTransmitted = 0
			for i := 0; i < numTrials; i++ {
				totalCellsTransmitted += runTrialTotalCellsVsDiffSize(i+1, universeSize, bench.symmetricDiffSize, rng)
			}
		})

		averageFloatCellsTransmitted := float64(totalCellsTransmitted) / float64(numTrials)
		averageCellsTransmitted := int(math.Ceil(averageFloatCellsTransmitted))

		// Write the result to the CSV file.
		writer.Write([]string{
			fmt.Sprintf("%d", bench.symmetricDiffSize),
			fmt.Sprintf("%d", averageCellsTransmitted),
		})
	}
}
