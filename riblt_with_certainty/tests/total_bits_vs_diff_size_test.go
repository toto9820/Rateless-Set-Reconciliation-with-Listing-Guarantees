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

type MappingType string

const (
	EGH MappingType = "egh"
	OLS MappingType = "ols"
)

// runTrial simulates a reconciliation trial for benchmarking.
func runTrialTotalCellsVsDiffSize(trialNumber int,
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
	chosenIndices := rng.Perm(universeSize)[:universeSize-symmetricDiffSize]
	for _, idx := range chosenIndices {
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

	cost := uint64(0)

	for {
		ibfAlice.AddSymbols(alice)
		ibfBob.AddSymbols(bob)

		ibfDiff := ibfBob.Subtract(ibfAlice)
		bobWithoutAlice, ok := ibfDiff.Decode()

		// occurrences := make(map[uint64]bool)
		// hasDuplicates := false

		// for _, value := range bobWithoutAlice {
		// 	// Check if value already exists in the occurrences map
		// 	if occurrences[uint64(value.(Uint64Symbol))] {
		// 		fmt.Printf("Duplicate found: %d\n", value)
		// 		hasDuplicates = true
		// 	} else {
		// 		occurrences[uint64(value.(Uint64Symbol))] = true
		// 	}
		// }

		// if !hasDuplicates {
		// 	fmt.Println("No duplicates found.")
		// }

		if ok == false {
			continue
		}

		if len(bobWithoutAlice) == symmetricDiffSize {
			cost = ibfDiff.Size
			break
		}
	}

	fmt.Printf("Trial %d for %s method, Symmetric Difference len: %d with %d cells\n",
		trialNumber, mappingType, symmetricDiffSize, cost)
	log.Printf("Trial %d for %s method, Symmetric Difference len: %d with %d cells\n",
		trialNumber, mappingType, symmetricDiffSize, cost)

	return cost
}

// BenchmarkReconciliation benchmarks both EGH and OLS methods
func BenchmarkTotalBitsVsDiffSize(b *testing.B) {
	benches := []struct {
		symmetricDiffSize int
	}{
		{1},
		{10},
		{100},
		{1000},
		{10000},
	}

	cellSizeInBits := 64 * 3
	universeSize := int(math.Pow(10, 6))
	// For Debugging
	// numTrials := 1
	numTrials := 10
	mappingTypes := []MappingType{EGH, OLS}

	for _, mappingType := range mappingTypes {
		// Create a CSV file for each mapping type
		filename := fmt.Sprintf("%s_total_bits_vs_diff_size_set_inside_set.csv",
			string(mappingType))
		file, err := os.Create(filename)
		if err != nil {
			b.Fatalf("Error creating file for %s: %v", mappingType, err)
		}
		defer file.Close()

		writer := csv.NewWriter(file)
		defer writer.Flush()

		// Write the header row
		writer.Write([]string{"Symmetric Diff Size", "Total Bits Transmitted"})

		for _, bench := range benches {
			b.Run(fmt.Sprintf("%s_Universe=%d_Diff=%d",
				mappingType, universeSize, bench.symmetricDiffSize),
				func(b *testing.B) {
					results := make(chan uint64, numTrials)
					var totalCellsTransmitted uint64

					var wg sync.WaitGroup
					wg.Add(numTrials)

					for i := 0; i < numTrials; i++ {
						go func(trialNum int) {
							defer wg.Done()
							rng := rand.New(rand.NewSource(time.Now().UnixNano() + int64(trialNum)))
							result := runTrialTotalCellsVsDiffSize(
								trialNum+1,
								universeSize,
								bench.symmetricDiffSize,
								mappingType,
								rng,
							)
							results <- result
						}(i)
					}

					go func() {
						wg.Wait()
						close(results)
					}()

					for result := range results {
						totalCellsTransmitted += result
					}

					averageFloatCellsTransmitted := float64(totalCellsTransmitted) / float64(numTrials)
					averageCellsTransmitted := int(math.Ceil(averageFloatCellsTransmitted))

					writer.Write([]string{
						fmt.Sprintf("%d", bench.symmetricDiffSize),
						fmt.Sprintf("%d", averageCellsTransmitted*cellSizeInBits),
					})
				})
		}
	}
}
