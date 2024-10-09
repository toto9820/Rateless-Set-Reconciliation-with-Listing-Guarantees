package riblt_with_certainty

import (
	"encoding/csv"
	"fmt"
	"math"
	"math/rand"
	"os"
	"sort"
	"testing"
	"time"
)

// runTrialAdditionalCellsVsDiffSize simulates a reconciliation trial for benchmarking.
func runTrialAdditionalCellsVsDiffSize(trialNumber int,
	universeSize int,
	symmetricDiffSizes []int,
	rng *rand.Rand) []int {
	// For superset assumption
	// Bob's set will include all elements from 1 to universeSize.
	bob := make([]uint64, 0, universeSize)
	for i := 1; i <= universeSize; i++ {
		bob = append(bob, uint64(i))
	}

	// Sort the symmetric difference sizes
	sort.Ints(symmetricDiffSizes)

	maxSymmetricDiffSize := symmetricDiffSizes[len(symmetricDiffSizes)-1]

	// Alice's set will include universeSize - symmetricDiffSize elements.
	alice := make([]uint64, 0, universeSize-maxSymmetricDiffSize)

	// Randomly choose indices from Bob's set to include in Alice's set.
	chosenIndices := rng.Perm(universeSize)[:universeSize-maxSymmetricDiffSize] // Random permutation.

	// fmt.Println("Trial:", trialNumber, "Chosen Indices:", chosenIndices)

	for _, idx := range chosenIndices {
		// idx is within 0 to universeSize-1
		alice = append(alice, bob[idx])
	}

	// Sort Alice's set.
	sort.Slice(alice, func(i, j int) bool {
		return alice[i] < alice[j]
	})

	intialCells := uint64(1000)

	ibfAlice := NewIBF(intialCells)
	ibfBob := NewIBF(intialCells)

	// Prepare a results list for storing the number of cells for each symmetric difference size
	results := make([]int, len(symmetricDiffSizes))
	idx := 0
	curSymmetricDiffSize := 0
	prevCellsSize := 0

	for {
		ibfAlice.AddSymbols(alice)
		ibfBob.AddSymbols(bob)

		// Subtract the two IBFs and Decode the result to find the differences
		ibfDiff := ibfBob.Subtract(ibfAlice)
		bobWithoutAlice, _, _ := ibfDiff.Decode()

		if len(bobWithoutAlice) > 0 {
			curSymmetricDiffSize = len(bobWithoutAlice)

			for (idx < len(symmetricDiffSizes)) &&
				(curSymmetricDiffSize >= symmetricDiffSizes[idx]) {
				results[idx] = int(ibfDiff.Size) - prevCellsSize
				prevCellsSize = int(ibfDiff.Size)

				idx++
			}
		}

		if curSymmetricDiffSize >= maxSymmetricDiffSize {
			break
		}

		// Print the current symmetric difference size
		fmt.Println("Trial:", trialNumber, "Current Symmetric Difference Size:", curSymmetricDiffSize)
	}

	return results
}

// BenchmarkReconciliation benchmarks the reconciliation
// process with a fixed universe size and different
// symmetric difference sizes.
func BenchmarkAdditionalBitsVsDiffSize(b *testing.B) {
	// This is the maximum symmetric difference size to test
	maxSymmetricDiffSize := 100000
	// maxSymmetricDiffSize := 1000

	// Generate symmetric difference sizes as powers of 10 up to the maximum
	var symmetricDiffSizes []int
	for i := 0; i <= int(math.Log10(float64(maxSymmetricDiffSize))); i++ {
		symmetricDiffSizes = append(symmetricDiffSizes, int(math.Pow(10, float64(i))))
	}

	// Prepare a CSV file to store the results.
	file, err := os.Create("egh_additional_bits_vs_diff_size_set_inside_set.csv")
	if err != nil {
		fmt.Println("Error creating file:", err)
		return
	}
	defer file.Close()

	writer := csv.NewWriter(file)
	defer writer.Flush()

	// Write the header row to the CSV file.
	writer.Write([]string{"Symmetric Diff Size", "Additional Bits Transmitted"})

	// Set the number of trials
	numTrials := 10
	// numTrials := 1

	// Set a global seed
	globalSeed := time.Now().UnixNano()

	// Bits per IBLT cell (3 fields - count, xorSum, checkSum)
	// Each field is 64 bit.
	cellSizeInBits := 64 * 3

	// Define the universe size
	// universeSize := int(math.Pow(10, 2))
	universeSize := int(math.Pow(10, 6))

	aggregatedAdditionalCells := make([]int, len(symmetricDiffSizes))

	// Create a slice to hold trial results
	trialResults := make([][]int, numTrials)

	b.Run(fmt.Sprintf("Universe=%d, Max Diff=%d", universeSize, symmetricDiffSizes[len(symmetricDiffSizes)-1]), func(b *testing.B) {
		for i := 0; i < numTrials; i++ {
			// Create a new random number generator with a unique seed for each trial
			trialSeed := globalSeed + int64(i) + rand.Int63()
			rng := rand.New(rand.NewSource(trialSeed))

			// Run the trial and get the additional cells transmitted
			results := runTrialAdditionalCellsVsDiffSize(i+1, universeSize, symmetricDiffSizes, rng)
			trialResults[i] = results

			// Add a small delay between trials
			time.Sleep(time.Millisecond)
		}
	})

	// Aggregate additional cells from each trial result
	for _, trialResult := range trialResults {
		for idx, additionalCells := range trialResult {
			aggregatedAdditionalCells[idx] += additionalCells
		}
	}

	// Calculate the average additional cells transmitted for this symmetric difference size
	for idx := range aggregatedAdditionalCells {
		avgAdditionalCells := int(math.Ceil(float64(aggregatedAdditionalCells[idx]) / float64(numTrials)))
		// Write the result to the CSV file.
		writer.Write([]string{
			fmt.Sprintf("%d", symmetricDiffSizes[idx]),
			fmt.Sprintf("%d", avgAdditionalCells*cellSizeInBits),
		})
	}
}
