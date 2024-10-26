package riblt_with_certainty_test

import (
	"encoding/csv"
	"fmt"
	"math"
	"math/rand"
	"os"
	"sort"
	"sync"
	"testing"
	"time"

	. "github.com/toto9820/Rateless-Set-Reconciliation-with-Listing-Guarantees/riblt_with_certainty"
)

// runTrialAdditionalCellsVsDiffSize simulates a reconciliation trial for benchmarking.
func runTrialAdditionalCellsVsDiffSize(trialNumber int,
	universeSize int,
	symmetricDiffSizes []int,
	mappingType MappingType,
	rng *rand.Rand) []int {
	// For superset assumption
	// Bob's set will include all elements from 1 to universeSize.
	bob := make([]Symbol, 0, universeSize)
	for i := 1; i <= universeSize; i++ {
		bob = append(bob, Uint64Symbol(i))
	}

	// Sort the symmetric difference sizes
	sort.Ints(symmetricDiffSizes)

	maxSymmetricDiffSize := symmetricDiffSizes[len(symmetricDiffSizes)-1]

	// Alice's set will include universeSize - symmetricDiffSize elements.
	alice := make([]Symbol, 0, universeSize-maxSymmetricDiffSize)

	// Randomly choose indices from Bob's set to include in Alice's set.
	chosenIndices := rng.Perm(universeSize)[:universeSize-maxSymmetricDiffSize] // Random permutation.

	// fmt.Println("Trial:", trialNumber, "Chosen Indices:", chosenIndices)

	for _, idx := range chosenIndices {
		// idx is within 0 to universeSize-1
		alice = append(alice, bob[idx])
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
		bobWithoutAlice, _ := ibfDiff.Decode()

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
	maxSymmetricDiffSize := 10000

	mappingTypes := []MappingType{EGH, OLS}

	var symmetricDiffSizes []int
	for i := 0; i <= int(math.Log10(float64(maxSymmetricDiffSize))); i++ {
		symmetricDiffSizes = append(symmetricDiffSizes, int(math.Pow(10, float64(i))))
	}

	for _, mappingType := range mappingTypes {
		file, err := os.Create(fmt.Sprintf("%s_additional_bits_vs_diff_size_set_inside_set.csv", string(mappingType)))
		if err != nil {
			b.Fatalf("Error creating file: %v", err)
		}
		defer file.Close()

		writer := csv.NewWriter(file)
		defer writer.Flush()

		writer.Write([]string{"Symmetric Diff Size", "Additional Bits Transmitted"})

		numTrials := 10
		cellSizeInBits := 64 * 3
		universeSize := int(math.Pow(10, 6))

		aggregatedAdditionalCells := make([]int64, len(symmetricDiffSizes))
		globalSeed := time.Now().UnixNano()

		b.Run(fmt.Sprintf("Universe=%d, Max Diff=%d", universeSize, symmetricDiffSizes[len(symmetricDiffSizes)-1]), func(b *testing.B) {
			results := make(chan []int, numTrials)

			var wg sync.WaitGroup
			wg.Add(numTrials)

			for i := 0; i < numTrials; i++ {
				go func(trialNum int) {
					defer wg.Done()
					trialSeed := globalSeed + int64(trialNum) + rand.Int63()
					rng := rand.New(rand.NewSource(trialSeed))
					trialResults := runTrialAdditionalCellsVsDiffSize(trialNum+1, universeSize, symmetricDiffSizes, mappingType, rng)
					results <- trialResults
				}(i)
			}

			go func() {
				wg.Wait()
				close(results)
			}()

			for trialResult := range results {
				for idx, additionalCells := range trialResult {
					aggregatedAdditionalCells[idx] += int64(additionalCells)
				}
			}
		})

		for idx, totalAdditionalCells := range aggregatedAdditionalCells {
			avgAdditionalCells := int(math.Ceil(float64(totalAdditionalCells) / float64(numTrials)))
			writer.Write([]string{
				fmt.Sprintf("%d", symmetricDiffSizes[idx]),
				fmt.Sprintf("%d", avgAdditionalCells*cellSizeInBits),
			})
		}
	}
}
