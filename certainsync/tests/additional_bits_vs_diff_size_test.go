package certainsync_test

import (
	"encoding/csv"
	"fmt"
	"log"
	"math"
	"math/rand"
	"os"
	"path/filepath"
	"sync"
	"testing"
	"time"
)

// BenchmarkReconciliation benchmarks the reconciliation
// process with a fixed universe size and different
// symmetric difference sizes.
func BenchmarkAdditionalBitsVsDiffSize(b *testing.B) {
	benches := []struct {
		symmetricDiffSize int
	}{
		{1},
		{10},
		{100},
		{1000},
		{10000},
	}

	// benches := []struct {
	// 	symmetricDiffSize int
	// }{
	// 	{1},
	// 	{10},
	// }

	cwd, err := os.Getwd()
	if err != nil {
		log.Fatalf("Failed to get current working directory: %v", err)
	}

	// numTrials := 10
	numTrials := 1
	// numTrials := 10
	numTrials := 1
	universeSize := int(math.Pow(10, 6))

	// mappingTypes := []MappingType{EGH, OLS}
	mappingTypes := []MappingType{OLS}
	// mappingTypes := []MappingType{EGH, OLS}
	mappingTypes := []MappingType{OLS}

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

		var maxSymmetricDiffSize int
		if mappingType == OLS {
			maxSymmetricDiffSize = int(math.Ceil(math.Sqrt(float64(universeSize))))
		} else {
			maxSymmetricDiffSize = benches[len(benches)-1].symmetricDiffSize
		}

		prevSymmtericDiffSize := 0
		prevAvgBitsTransmitted := 0

		// Add first row of 0,0.
		writer.Write([]string{
			fmt.Sprintf("%d", prevSymmtericDiffSize),
			fmt.Sprintf("%d", prevAvgBitsTransmitted),
		})

		for _, bench := range benches {
			if bench.symmetricDiffSize > maxSymmetricDiffSize {
				continue
			}

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

					avgFloatBitsTransmitted := float64(totalBitsTransmitted) / float64(numTrials)
					avgBitsTransmitted := int(math.Ceil(avgFloatBitsTransmitted))
					avgAdditionalBitsTransmitted := int(math.Ceil(float64(avgBitsTransmitted-prevAvgBitsTransmitted) / float64(bench.symmetricDiffSize-prevSymmtericDiffSize)))

					prevSymmtericDiffSize = bench.symmetricDiffSize
					prevAvgBitsTransmitted = avgBitsTransmitted

					writer.Write([]string{
						fmt.Sprintf("%d", bench.symmetricDiffSize),
						fmt.Sprintf("%d", avgAdditionalBitsTransmitted),
					})
				})
		}
	}
}
