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

// runTrial simulates a reconciliation trial for benchmarking.
func runTrialTotalCellsVsUniverseSize(trialNumber int,
	universeSize int,
	symmetricDiffSize int,
	mappingType MappingType,
	rng *rand.Rand) uint64 {
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

	initialCells := uint64(1000)
	var ibfAlice, ibfBob *InvertibleBloomFilter

	switch mappingType {
	case EGH:
		ibfAlice = NewIBF(initialCells, "uint64", &EGHMapping{})
		ibfBob = NewIBF(initialCells, "uint64", &EGHMapping{})
	case OLS:
		olsMapping := OLSMapping{
			Order: uint64(math.Sqrt(float64(universeSize))),
		}
		ibfAlice = NewIBF(initialCells, "uint64", &olsMapping)
		ibfBob = NewIBF(initialCells, "uint64", &olsMapping)
	}

	cost := uint64(0)

	for {
		ibfAlice.AddSymbols(alice)
		ibfBob.AddSymbols(bob)

		// Subtract the two IBFs and Decode the result to find the differences
		ibfDiff := ibfBob.Subtract(ibfAlice)
		bobWithoutAlice, ok := ibfDiff.Decode()

		if ok == false {
			continue
		}

		if len(bobWithoutAlice) == symmetricDiffSize {
			cost = ibfDiff.Size
			break
		}
	}

	fmt.Printf("Trial %d for EGH method for CertainSync IBLT, Symmetric Difference len: %d with %d cells", trialNumber, symmetricDiffSize, cost)
	fmt.Println()

	log.Printf("Trial %d for EGH method for CertainSync IBLT, Symmetric Difference len: %d with %d cells\n", trialNumber, symmetricDiffSize, cost)

	// Return number of coded symbols transmitted
	return cost
}

// BenchmarkTotalBitsVsUniverseSize benchmarks the reconciliation
// process with fixed symmetric difference sizes and varying universe sizes.
func BenchmarkTotalBitsVsUniverseSize(b *testing.B) {
	symmetricDiffSizes := []int{1, 3, 30, 90}
	universeSizes := []int{
		int(math.Pow(10, 3)), // 1,000
		int(math.Pow(10, 4)), // 10,000
		int(math.Pow(10, 5)), // 100,000
		int(math.Pow(10, 6)), // 1,000,000
		int(math.Pow(10, 7)), // 10,000,000
	}

	cellSizeInBits := 64 * 3
	numTrials := 10
	mappingTypes := []MappingType{EGH, OLS}

	for _, mappingType := range mappingTypes {
		for _, symmetricDiffSize := range symmetricDiffSizes {
			// Prepare a CSV file to store the results for the current symmetric difference size.
			file, err := os.Create(fmt.Sprintf("%s_total_bits_vs_universe_size_for_diff_size_%d_set_inside_set.csv", string(mappingType), symmetricDiffSize))
			if err != nil {
				b.Fatalf("Error creating file: %v", err)
			}
			defer file.Close()

			writer := csv.NewWriter(file)
			defer writer.Flush()

			// Write the header row to the CSV file.
			writer.Write([]string{"Universe Size", "Total Bits Transmitted"})

			for _, universeSize := range universeSizes {
				b.Run(fmt.Sprintf("DiffSize=%d, Universe=%d", symmetricDiffSize, universeSize), func(b *testing.B) {
					results := make(chan uint64, numTrials)
					var totalCellsTransmitted uint64

					// Create a wait group to synchronize goroutines
					var wg sync.WaitGroup
					wg.Add(numTrials)

					// Run trials concurrently
					for i := 0; i < numTrials; i++ {
						go func(trialNum int) {
							defer wg.Done()
							// Create a local random number generator with a time-based seed
							rng := rand.New(rand.NewSource(time.Now().UnixNano() + int64(trialNum)))
							result := runTrialTotalCellsVsUniverseSize(trialNum+1, universeSize, symmetricDiffSize, mappingType, rng)
							results <- result
						}(i)
					}

					// Close the results channel when all goroutines are done
					go func() {
						wg.Wait()
						close(results)
					}()

					// Collect results
					for result := range results {
						totalCellsTransmitted += result
					}

					averageFloatCellsTransmitted := float64(totalCellsTransmitted) / float64(numTrials)
					averageCellsTransmitted := int(math.Ceil(averageFloatCellsTransmitted))

					// Write the result to the CSV file.
					writer.Write([]string{
						fmt.Sprintf("%d", universeSize),
						fmt.Sprintf("%d", averageCellsTransmitted*cellSizeInBits),
					})
				})
			}

			// Flush the data to the file.
			writer.Flush()
		}
	}
}
