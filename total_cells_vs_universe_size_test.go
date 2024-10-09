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
func runTrialTotalCellsVsUniverseSize(trialNumber int,
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

// BenchmarkTotalBitsVsUniverseSize benchmarks the reconciliation
// process with fixed symmetric difference sizes and varying universe sizes.
func BenchmarkTotalBitsVsUniverseSize(b *testing.B) {
	// Define the symmetric difference sizes to test
	symmetricDiffSizes := []int{1, 3, 30, 90}
	// symmetricDiffSizes := []int{90}

	// Define the universe sizes to test
	universeSizes := []int{
		int(math.Pow(10, 3)), // 1,000
		int(math.Pow(10, 4)), // 10,000
		int(math.Pow(10, 5)), // 100,000
		int(math.Pow(10, 6)), // 1,000,000
		int(math.Pow(10, 7)), // 10,000,000
	}

	// Create a local random number generator with a time-based seed
	rng := rand.New(rand.NewSource(time.Now().UnixNano()))

	// Bits per IBLT cell (3 fields - count, xorSum, checkSum)
	// Each field is 64 bit.
	cellSizeInBits := 64 * 3

	for _, symmetricDiffSize := range symmetricDiffSizes {
		// Prepare a CSV file to store the results for the current symmetric difference size.
		file, err := os.Create(fmt.Sprintf("egh_total_bits_vs_universe_size_for_diff_size_%d_set_inside_set.csv", symmetricDiffSize))
		if err != nil {
			fmt.Println("Error creating file:", err)
			return
		}
		defer file.Close()

		writer := csv.NewWriter(file)
		defer writer.Flush()

		// Write the header row to the CSV file.
		writer.Write([]string{"Universe Size", "Total Bits Transmitted"})

		// Set the number of trials
		numTrials := 10

		for _, universeSize := range universeSizes {
			var totalCellsTransmitted uint64
			b.Run(fmt.Sprintf("DiffSize=%d, Universe=%d", symmetricDiffSize, universeSize), func(b *testing.B) {
				totalCellsTransmitted = 0
				for i := 0; i < numTrials; i++ {
					totalCellsTransmitted += runTrialTotalCellsVsUniverseSize(i+1, universeSize, symmetricDiffSize, rng)
				}
			})

			averageFloatCellsTransmitted := float64(totalCellsTransmitted) / float64(numTrials)
			averageCellsTransmitted := int(math.Ceil(averageFloatCellsTransmitted))

			// Write the result to the CSV file.
			writer.Write([]string{
				fmt.Sprintf("%d", universeSize),
				fmt.Sprintf("%d", averageCellsTransmitted*cellSizeInBits),
			})
		}

		// Flush the data to the file.
		writer.Flush()
	}
}
