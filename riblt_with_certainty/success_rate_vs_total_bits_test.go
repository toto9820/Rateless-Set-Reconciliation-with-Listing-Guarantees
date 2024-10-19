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

// Define a struct to hold both SuccessRate and TotalCells
type Result struct {
	SuccessRate float64
	TotalCells  int
}

// runTrialSuccessRateVsTotalCells simulates a reconciliation trial and computes the success rate vs. total cells.
func runTrialSuccessRateVsTotalCells(trialNumber int, universeSize int, symmetricDiffSize int, rng *rand.Rand) []Result {
	// For superset assumption
	// Bob's set will include all elements from 1 to universeSize.
	bob := make([]Symbol, 0, universeSize)
	for i := 1; i <= universeSize; i++ {
		bob = append(bob, Uint64Symbol(i))
	}

	// Alice's set will include universeSize - symmetricDiffSize elements.
	alice := make([]Symbol, 0, universeSize-symmetricDiffSize)

	// Randomly choose indices from Bob's set to include in Alice's set.
	chosenIndices := rng.Perm(universeSize)[:universeSize-symmetricDiffSize] // Random permutation.
	for _, idx := range chosenIndices {
		alice = append(alice, bob[idx]) // idx is within 0 to universeSize-1
	}

	// Sort Alice's set.
	sort.Slice(alice, func(i, j int) bool {
		return uint64(alice[i].(Uint64Symbol)) < uint64(alice[j].(Uint64Symbol))
	})

	intialCells := uint64(1000)

	ibfAlice := NewIBF(intialCells, "uint64")
	ibfBob := NewIBF(intialCells, "uint64")

	// Initialize the results to store success rate vs. total cells
	results := []Result{}
	totalCells := 0
	curSymmetricDiffSize := 0

	// Continue transmitting coded symbols until symmetricDiffSize elements are decoded
	for curSymmetricDiffSize < symmetricDiffSize {
		ibfAlice.AddSymbols(alice)
		ibfBob.AddSymbols(bob)

		// Subtract the two IBFs and Decode the result to find the differences
		ibfDiff := ibfBob.Subtract(ibfAlice)
		bobWithoutAlice, _ := ibfDiff.Decode()

		if len(bobWithoutAlice) > 0 {
			curSymmetricDiffSize = len(bobWithoutAlice)

			// Calculate the current success rate
			successRate := float64(curSymmetricDiffSize) / float64(symmetricDiffSize)
			// Append both success rate and total cells to results
			results = append(results, Result{
				SuccessRate: successRate,
				TotalCells:  int(ibfDiff.Size),
			})

			// Stop when success rate reaches 1.0
			if successRate == 1.0 {
				totalCells = int(ibfDiff.Size)
				break
			}
		}
	}

	fmt.Printf("Trial %d with universe size %d for CertainSync IBLT, Symmetric Difference len: %d, Total Cells: %d\n", trialNumber, universeSize, symmetricDiffSize, totalCells)
	log.Printf("Trial %d with universe size %d for CertainSync IBLT, Symmetric Difference len: %d, Total Cells: %d\n", trialNumber, universeSize, symmetricDiffSize, totalCells)

	return results
}

// BenchmarkSuccessRateVsTotalBits benchmarks the reconciliation process with fixed universe size and varying symmetric difference sizes.
func BenchmarkSuccessRateVsTotalBits(b *testing.B) {
	// Define the symmetric difference sizes to test
	symmetricDiffSizes := []int{1, 3, 30, 100, 300, 1000}

	// symmetricDiffSizes := []int{100}

	// Fix the universe size
	universeSize := int(math.Pow(10, 6))

	// Create a local random number generator with a time-based seed
	rng := rand.New(rand.NewSource(time.Now().UnixNano()))

	// Bits per IBLT cell (3 fields - count, xorSum, checkSum)
	// Each field is 64 bit.
	cellSizeInBits := 64 * 3

	// Set the number of trials
	numTrials := 10

	for _, symmetricDiffSize := range symmetricDiffSizes {
		// Prepare a CSV file to store the results for the current symmetric difference size.
		file, err := os.Create(fmt.Sprintf("egh_success_rate_vs_total_bits_diff_size_%d_set_inside_set.csv", symmetricDiffSize))
		if err != nil {
			fmt.Println("Error creating file:", err)
			return
		}
		defer file.Close()

		writer := csv.NewWriter(file)
		defer writer.Flush()

		// Write the header row to the CSV file.
		writer.Write([]string{"Total Bits Transmitted", "Success Probability"})

		// Initialize to store all trials' results
		allResults := [][]Result{}
		maxLength := 0

		// Run trials for each symmetricDiffSize
		for i := 0; i < numTrials; i++ {
			// Perform the trial and capture the results
			results := runTrialSuccessRateVsTotalCells(i+1, universeSize, symmetricDiffSize, rng)

			// Keep track of the longest result (for padding purposes)
			if len(results) > maxLength {
				maxLength = len(results)
			}

			// Store the trial results
			allResults = append(allResults, results)
		}

		// Average the results and pad shorter ones with 1.0
		avgResults := make([]Result, maxLength)

		for _, results := range allResults {
			for j := 0; j < maxLength; j++ {
				if j < len(results) {
					avgResults[j].SuccessRate += results[j].SuccessRate
					avgResults[j].TotalCells += results[j].TotalCells
				} else {
					avgResults[j].SuccessRate += 1.0 // Pad with 1.0 for shorter trials
					avgResults[j].TotalCells += results[len(results)-1].TotalCells
				}
			}
		}

		// Divide by the number of trials to get the average
		for j := 0; j < maxLength; j++ {
			avgResults[j].SuccessRate = avgResults[j].SuccessRate / float64(numTrials)
			avgResults[j].TotalCells = int(math.Ceil(float64(avgResults[j].TotalCells) / float64(numTrials)))
		}

		// Write in first line 0,0.
		writer.Write([]string{
			fmt.Sprintf("%d", 0),
			fmt.Sprintf("%.4f", 0.0),
		})

		// Write the averaged results to the CSV file
		for _, avgResult := range avgResults {
			writer.Write([]string{
				fmt.Sprintf("%d", avgResult.TotalCells*cellSizeInBits), // Adding 1 because the index starts from 0
				fmt.Sprintf("%.4f", avgResult.SuccessRate),
			})
		}

		// Flush the data to the file.
		writer.Flush()
	}
}
