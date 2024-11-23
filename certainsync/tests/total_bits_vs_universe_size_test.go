package certainsync_test

import (
	"encoding/csv"
	"fmt"
	"log"
	"math"
	"math/rand"
	"os"
	"path/filepath"
	"sort"
	"sync"
	"testing"
	"time"

	"github.com/holiman/uint256"
	. "github.com/toto9820/Rateless-Set-Reconciliation-with-Listing-Guarantees/certainsync"
)

// runTrial simulates a reconciliation trial for benchmarking.
func runTrialTotalBitsVsUniverseSize(trialNumber int,
	universeSize int,
	symmetricDiffSize int,
	mappingType MappingType,
	rng *rand.Rand) uint64 {
	// For superset assumption
	// Bob's set will include all elements from 1 to universeSize.
	bob := make([]*uint256.Int, 0, universeSize)
	for i := 1; i <= universeSize; i++ {
		bob = append(bob, uint256.NewInt(uint64((i))))
	}

	// Alice's set will include universeSize - symmetricDiffSize elements.
	alice := make([]*uint256.Int, 0, universeSize-symmetricDiffSize)

	// Randomly choose indices from Bob's set to include in Alice's set.
	chosenIndices := rng.Perm(universeSize)[:universeSize-symmetricDiffSize] // Random permutation.
	for _, idx := range chosenIndices {
		alice = append(alice, bob[idx]) // idx is within 0 to universeSize-1
	}

	// Sort Alice's set.
	sort.Slice(alice, func(i, j int) bool {
		return alice[i].Cmp(alice[j]) == -1
	})

	var ibfAlice, ibfBob *InvertibleBloomFilter

	switch mappingType {
	case EGH:
		ibfAlice = NewIBF(uint256.NewInt(uint64(universeSize)), &EGHMapping{})
		ibfBob = NewIBF(uint256.NewInt(uint64(universeSize)), &EGHMapping{})
	case OLS:
		olsMapping := OLSMapping{
			Order: uint64(math.Ceil(math.Sqrt(float64(universeSize)))),
		}
		ibfAlice = NewIBF(uint256.NewInt(uint64(universeSize)), &olsMapping)
		ibfBob = NewIBF(uint256.NewInt(uint64(universeSize)), &olsMapping)
	}

	cost := uint64(0)

	for {
		ibfAlice.AddSymbols(alice)

		cost = ibfAlice.GetTransmittedBitsSize()

		ibfBob.AddSymbols(bob)

		// Subtract the two IBFs and Decode the result to find the differences
		ibfDiff := ibfBob.Subtract(ibfAlice)
		bobWithoutAlice, _, ok := ibfDiff.Decode()

		if ok == false {
			continue
		}

		if len(bobWithoutAlice) == symmetricDiffSize {
			break
		}
	}

	fmt.Printf("Trial %d for %s method for CertainSync IBLT, Symmetric Difference len: %d with %d bits", trialNumber, mappingType, symmetricDiffSize, cost)
	fmt.Println()

	log.Printf("Trial %d for %s method for CertainSync IBLT, Symmetric Difference len: %d with %d bits\n", trialNumber, mappingType, symmetricDiffSize, cost)

	// Return number of bits transmitted
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

	cwd, err := os.Getwd()
	if err != nil {
		log.Fatalf("Failed to get current working directory: %v", err)
	}

	numTrials := 10

	mappingTypes := []MappingType{EGH, OLS}

	for _, mappingType := range mappingTypes {
		for _, symmetricDiffSize := range symmetricDiffSizes {
			// Prepare a CSV file to store the results for the current symmetric difference size.
			filename := fmt.Sprintf("%s_total_bits_vs_universe_size_for_diff_size_%d_set_inside_set.csv", string(mappingType), symmetricDiffSize)
			filePath := filepath.Join(cwd, "results", filename)

			file, err := os.Create(filePath)
			if err != nil {
				b.Fatalf("Error creating file for %s: %v", mappingType, err)
			}
			defer file.Close()

			writer := csv.NewWriter(file)
			defer writer.Flush()

			// Write the header row to the CSV file.
			writer.Write([]string{"Universe Size", "Total Bits Transmitted"})

			b.Run(fmt.Sprintf("MappingType=%s_DiffSize=%d", mappingType, symmetricDiffSize),
				func(b *testing.B) {
					for _, universeSize := range universeSizes {
						results := make(chan uint64, numTrials)
						var totalBitsTransmitted uint64

						var wg sync.WaitGroup
						wg.Add(numTrials)

						// Run trials concurrently
						for i := 0; i < numTrials; i++ {
							go func(trialNum int) {
								defer wg.Done()
								rng := rand.New(rand.NewSource(time.Now().UnixNano() + int64(trialNum)))
								result := runTrialTotalBitsVsUniverseSize(
									trialNum+1, universeSize, symmetricDiffSize, mappingType, rng)
								results <- result
							}(i)
						}

						// Close results channel when all goroutines are done
						go func() {
							wg.Wait()
							close(results)
						}()

						// Collect results
						for result := range results {
							totalBitsTransmitted += result
						}

						averageFloatBitsTransmitted := float64(totalBitsTransmitted) / float64(numTrials)
						averageBitsTransmitted := int(math.Ceil(averageFloatBitsTransmitted))

						// Write the result to the CSV file
						writer.Write([]string{
							fmt.Sprintf("%d", universeSize),
							fmt.Sprintf("%d", averageBitsTransmitted),
						})
						writer.Flush()
					}
				})

			file.Close()
		}
	}
}
