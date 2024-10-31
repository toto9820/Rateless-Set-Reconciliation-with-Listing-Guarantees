package certainsync_test

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

	. "github.com/toto9820/Rateless-Set-Reconciliation-with-Listing-Guarantees/certainsync"
)

// Define a struct to hold both SuccessRate and TotalBits
type Result struct {
	SuccessRate float64
	TotalBits   int
}

// runTrialSuccessRateVsTotalBits simulates a reconciliation trial
// and computes the success rate vs. total bits transmitted.
func runTrialSuccessRateVsTotalBits(trialNumber int,
	universeSize int,
	symmetricDiffSize int,
	mappingType MappingType,
	rng *rand.Rand) []Result {
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

	// Initialize the results to store success rate vs. total bits
	results := []Result{}
	totalBits := 0
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
				TotalBits:   int(ibfDiff.GetTransmittedBitsSize()),
			})

			// Stop when success rate reaches 1.0
			if successRate == 1.0 {
				totalBits = int(ibfDiff.GetTransmittedBitsSize())
				break
			}
		}
	}

	fmt.Printf("Trial %d with universe size %d for CertainSync IBLT, Symmetric Difference len: %d, Total Bits: %d\n", trialNumber, universeSize, symmetricDiffSize, totalBits)
	log.Printf("Trial %d with universe size %d for CertainSync IBLT, Symmetric Difference len: %d, Total Bits: %d\n", trialNumber, universeSize, symmetricDiffSize, totalBits)

	return results
}

// BenchmarkSuccessRateVsTotalBits benchmarks the
// reconciliation process with fixed universe size and
// varying symmetric difference sizes.
func BenchmarkSuccessRateVsTotalBits(b *testing.B) {
	symmetricDiffSizes := []int{1, 3, 30, 100, 300, 1000}
	universeSize := int(math.Pow(10, 6))
	numTrials := 10

	mappingTypes := []MappingType{EGH, OLS}

	for _, mappingType := range mappingTypes {
		for _, symmetricDiffSize := range symmetricDiffSizes {
			b.Run(fmt.Sprintf("DiffSize=%d", symmetricDiffSize), func(b *testing.B) {
				file, err := os.Create(fmt.Sprintf("%s_success_rate_vs_total_bits_diff_size_%d_set_inside_set.csv", string(mappingType), symmetricDiffSize))
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
						trialResults := runTrialSuccessRateVsTotalBits(trialNum+1, universeSize, symmetricDiffSize, mappingType, rng)
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
							avgResults[j].TotalBits += results[j].TotalBits
						} else {
							avgResults[j].SuccessRate += 1.0
							avgResults[j].TotalBits += results[len(results)-1].TotalBits
						}
					}
				}

				for j := 0; j < maxLength; j++ {
					avgResults[j].SuccessRate /= float64(numTrials)
					avgResults[j].TotalBits = int(math.Ceil(float64(avgResults[j].TotalBits) / float64(numTrials)))
				}

				writer.Write([]string{"0", "0.0000"})

				for _, avgResult := range avgResults {
					writer.Write([]string{
						fmt.Sprintf("%d", avgResult.TotalBits),
						fmt.Sprintf("%.4f", avgResult.SuccessRate),
					})
				}
			})
		}
	}
}