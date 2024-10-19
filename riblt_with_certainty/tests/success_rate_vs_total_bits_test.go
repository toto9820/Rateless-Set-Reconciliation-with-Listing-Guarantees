package riblt_with_certainty_test

import (
	"encoding/csv"
	"fmt"
	"log"
	"math"
	"math/rand"
	"os"
	"sort"
	"sync"
	"testing"
	"time"

	. "github.com/toto9820/Rateless-Set-Reconciliation-with-Listing-Guarantees/riblt_with_certainty"
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

	ibfAlice := NewIBF(intialCells, "uint64", &EGHMapping{})
	ibfBob := NewIBF(intialCells, "uint64", &EGHMapping{})

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
	symmetricDiffSizes := []int{1, 3, 30, 100, 300, 1000}
	universeSize := int(math.Pow(10, 6))
	cellSizeInBits := 64 * 3
	numTrials := 10

	for _, symmetricDiffSize := range symmetricDiffSizes {
		b.Run(fmt.Sprintf("DiffSize=%d", symmetricDiffSize), func(b *testing.B) {
			file, err := os.Create(fmt.Sprintf("egh_success_rate_vs_total_bits_diff_size_%d_set_inside_set.csv", symmetricDiffSize))
			if err != nil {
				b.Fatalf("Error creating file: %v", err)
			}
			defer file.Close()

			writer := csv.NewWriter(file)
			defer writer.Flush()

			writer.Write([]string{"Total Bits Transmitted", "Success Probability"})

			results := make(chan []Result, numTrials)
			var wg sync.WaitGroup
			wg.Add(numTrials)

			globalSeed := time.Now().UnixNano()

			for i := 0; i < numTrials; i++ {
				go func(trialNum int) {
					defer wg.Done()
					trialSeed := globalSeed + int64(trialNum) + rand.Int63()
					rng := rand.New(rand.NewSource(trialSeed))
					trialResults := runTrialSuccessRateVsTotalCells(trialNum+1, universeSize, symmetricDiffSize, rng)
					results <- trialResults
				}(i)
			}

			go func() {
				wg.Wait()
				close(results)
			}()

			var allResults [][]Result
			maxLength := 0

			for trialResult := range results {
				allResults = append(allResults, trialResult)
				if len(trialResult) > maxLength {
					maxLength = len(trialResult)
				}
			}

			avgResults := make([]Result, maxLength)

			for _, results := range allResults {
				for j := 0; j < maxLength; j++ {
					if j < len(results) {
						avgResults[j].SuccessRate += results[j].SuccessRate
						avgResults[j].TotalCells += results[j].TotalCells
					} else {
						avgResults[j].SuccessRate += 1.0
						avgResults[j].TotalCells += results[len(results)-1].TotalCells
					}
				}
			}

			for j := 0; j < maxLength; j++ {
				avgResults[j].SuccessRate /= float64(numTrials)
				avgResults[j].TotalCells = int(math.Ceil(float64(avgResults[j].TotalCells) / float64(numTrials)))
			}

			writer.Write([]string{"0", "0.0000"})

			for _, avgResult := range avgResults {
				writer.Write([]string{
					fmt.Sprintf("%d", avgResult.TotalCells*cellSizeInBits),
					fmt.Sprintf("%.4f", avgResult.SuccessRate),
				})
			}
		})
	}
}
