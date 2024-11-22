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

// runTrialAdditionalCellsVsDiffSize simulates a
// reconciliation trial for benchmarking.
func runTrialAdditionalBitsVsDiffSize(trialNumber int,
	universeSize int,
	symmetricDiffSizes []int,
	mappingType MappingType,
	rng *rand.Rand) []uint64 {
	// For superset assumption
	// Bob's set will include all elements from 1 to universeSize.
	bob := make([]*uint256.Int, 0, universeSize)
	for i := 1; i <= universeSize; i++ {
		bob = append(bob, uint256.NewInt(uint64((i))))
	}

	// Sort the symmetric difference sizes
	sort.Ints(symmetricDiffSizes)

	maxSymmetricDiffSize := symmetricDiffSizes[len(symmetricDiffSizes)-1]

	// Alice's set will include universeSize - symmetricDiffSize elements.
	alice := make([]*uint256.Int, 0, universeSize-maxSymmetricDiffSize)

	// Randomly choose indices from Bob's set to include in Alice's set.
	chosenIndices := rng.Perm(universeSize)[:universeSize-maxSymmetricDiffSize] // Random permutation.

	// fmt.Println("Trial:", trialNumber, "Chosen Indices:", chosenIndices)

	for _, idx := range chosenIndices {
		// idx is within 0 to universeSize-1
		alice = append(alice, bob[idx])
	}

	// Sort Alice's set.
	sort.Slice(alice, func(i, j int) bool {
		return alice[i].Cmp(alice[j]) == -1
	})

	var ibfAlice, ibfBob, receivedCells *InvertibleBloomFilter

	switch mappingType {
	case EGH:
		ibfAlice = NewIBF(uint256.NewInt(uint64(universeSize)), &EGHMapping{})
		ibfBob = NewIBF(uint256.NewInt(uint64(universeSize)), &EGHMapping{})
		receivedCells = NewIBF(uint256.NewInt(uint64(universeSize)), &EGHMapping{})
	case OLS:
		olsMapping := OLSMapping{
			Order: uint64(math.Ceil(math.Sqrt(float64(universeSize)))),
		}
		ibfAlice = NewIBF(uint256.NewInt(uint64(universeSize)), &olsMapping)
		ibfBob = NewIBF(uint256.NewInt(uint64(universeSize)), &olsMapping)
		receivedCells = NewIBF(uint256.NewInt(uint64(universeSize)), &olsMapping)
	}

	// Prepare a results list for storing the number of cells for each symmetric difference size
	results := make([]uint64, len(symmetricDiffSizes))
	idx := 0
	curSymmetricDiffSize := 0
	transmittedBits := uint64(0)
	prevTransmittedBits := uint64(0)

	for {
		ibfAlice.AddSymbols(alice)
		ibfBob.AddSymbols(bob)

		// Start - Simulation of communication //////////////////////////////

		ibfAliceBytes, err := ibfAlice.Serialize()

		transmittedBits += uint64(len(ibfAliceBytes)) * 8

		if err != nil {
			panic(err)
		}

		receivedCells.Deserialize(ibfAliceBytes)

		// End - Simulation of communication ////////////////////////////////

		// Subtract the two IBFs and Decode the result to find the differences
		ibfDiff := ibfBob.Subtract(receivedCells)
		bobWithoutAlice, _, _ := ibfDiff.Decode()

		if len(bobWithoutAlice) > 0 {
			curSymmetricDiffSize = len(bobWithoutAlice)

			for (idx < len(symmetricDiffSizes)) &&
				(curSymmetricDiffSize >= symmetricDiffSizes[idx]) {
				results[idx] = transmittedBits - prevTransmittedBits
				prevTransmittedBits = transmittedBits

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

	cwd, err := os.Getwd()
	if err != nil {
		log.Fatalf("Failed to get current working directory: %v", err)
	}

	mappingTypes := []MappingType{EGH, OLS}

	var symmetricDiffSizes []int
	for i := 0; i <= int(math.Log10(float64(maxSymmetricDiffSize))); i++ {
		symmetricDiffSizes = append(symmetricDiffSizes, int(math.Pow(10, float64(i))))
	}

	for _, mappingType := range mappingTypes {
		filename := fmt.Sprintf("%s_additional_bits_vs_diff_size_set_inside_set.csv", string(mappingType))

		filePath := filepath.Join(cwd, "results", filename)

		file, err := os.Create(filePath)
		if err != nil {
			b.Fatalf("Error creating file for %s: %v", mappingType, err)
		}

		if err != nil {
			b.Fatalf("Error creating file: %v", err)
		}
		defer file.Close()

		writer := csv.NewWriter(file)
		defer writer.Flush()

		writer.Write([]string{"Symmetric Diff Size", "Additional Bits Transmitted"})

		numTrials := 10
		universeSize := int(math.Pow(10, 6))

		aggregatedAdditionalBits := make([]int64, len(symmetricDiffSizes))
		globalSeed := time.Now().UnixNano()

		b.Run(fmt.Sprintf("Universe=%d, Max Diff=%d", universeSize, symmetricDiffSizes[len(symmetricDiffSizes)-1]), func(b *testing.B) {
			results := make(chan []uint64, numTrials)

			var wg sync.WaitGroup
			wg.Add(numTrials)

			for i := 0; i < numTrials; i++ {
				go func(trialNum int) {
					defer wg.Done()
					trialSeed := globalSeed + int64(trialNum) + rand.Int63()
					rng := rand.New(rand.NewSource(trialSeed))
					trialResults := runTrialAdditionalBitsVsDiffSize(trialNum+1, universeSize, symmetricDiffSizes, mappingType, rng)
					results <- trialResults
				}(i)
			}

			go func() {
				wg.Wait()
				close(results)
			}()

			for trialResult := range results {
				for idx, additionalCells := range trialResult {
					aggregatedAdditionalBits[idx] += int64(additionalCells)
				}
			}
		})

		for idx, totalAdditionalBits := range aggregatedAdditionalBits {
			avgAdditionalBits := int(math.Ceil(float64(totalAdditionalBits) / float64(numTrials)))
			writer.Write([]string{
				fmt.Sprintf("%d", symmetricDiffSizes[idx]),
				fmt.Sprintf("%d", avgAdditionalBits),
			})
		}
	}
}
