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

type MappingType string

const (
	EGH MappingType = "egh"
	OLS MappingType = "ols"
)

// runTrial simulates a reconciliation trial for benchmarking.
func runTrialTotalBitsVsDiffSize(trialNumber int,
	universeSize int,
	symmetricDiffSize int,
	mappingType MappingType,
	rng *rand.Rand) uint64 {

	// For Debugging
	// rng.Seed(0)

	// For superset assumption
	// Bob's set will include all elements from 1 to universeSize.
	bob := make([]*uint256.Int, 0, universeSize)
	for i := 1; i <= universeSize; i++ {
		bob = append(bob, uint256.NewInt(uint64((i))))
	}

	// Alice's set will include universeSize - symmetricDiffSize elements.
	alice := make([]*uint256.Int, 0, universeSize-symmetricDiffSize)

	// Randomly choose indices from Bob's set to include in Alice's set.
	chosenIndices := rng.Perm(universeSize)[:universeSize-symmetricDiffSize]
	for _, idx := range chosenIndices {
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

	transmittedBits := uint64(0)

	for {
		ibfAlice.AddSymbols(alice)

		// Start - Simulation of communication //////////////////////////////

		ibfAliceBytes, err := ibfAlice.Serialize()

		transmittedBits += uint64(len(ibfAliceBytes)) * 8

		if err != nil {
			panic(err)
		}

		receivedCells.Deserialize(ibfAliceBytes)

		// End - Simulation of communication ////////////////////////////////

		ibfBob.AddSymbols(bob)

		ibfDiff := ibfBob.Subtract(receivedCells)
		bobWithoutAlice, _, ok := ibfDiff.Decode()

		if ok == false {
			continue
		}

		if len(bobWithoutAlice) == symmetricDiffSize {
			break
		}
	}

	fmt.Printf("Trial %d for %s method, Symmetric Difference len: %d with %d bits\n",
		trialNumber, mappingType, symmetricDiffSize, transmittedBits)
	log.Printf("Trial %d for %s method, Symmetric Difference len: %d with %d bits\n",
		trialNumber, mappingType, symmetricDiffSize, transmittedBits)

	return transmittedBits
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

	cwd, err := os.Getwd()
	if err != nil {
		log.Fatalf("Failed to get current working directory: %v", err)
	}

	universeSize := int(math.Pow(10, 6))
	// For Debugging
	numTrials := 1
	// numTrials := 10
	mappingTypes := []MappingType{EGH, OLS}

	for _, mappingType := range mappingTypes {
		// Create a CSV file for each mapping type
		filename := fmt.Sprintf("%s_total_bits_vs_diff_size_set_inside_set.csv",
			string(mappingType))

		filePath := filepath.Join(cwd, "results", filename)

		file, err := os.Create(filePath)
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
					var totalBitsTransmitted uint64

					var wg sync.WaitGroup
					wg.Add(numTrials)

					for i := 0; i < numTrials; i++ {
						go func(trialNum int) {
							defer wg.Done()
							rng := rand.New(rand.NewSource(time.Now().UnixNano() + int64(trialNum)))
							result := runTrialTotalBitsVsDiffSize(
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
						totalBitsTransmitted += result
					}

					averageFloatBitsTransmitted := float64(totalBitsTransmitted) / float64(numTrials)
					averageBitsTransmitted := int(math.Ceil(averageFloatBitsTransmitted))

					writer.Write([]string{
						fmt.Sprintf("%d", bench.symmetricDiffSize),
						fmt.Sprintf("%d", averageBitsTransmitted),
					})
				})
		}
	}
}
