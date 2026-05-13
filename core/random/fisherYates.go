package random

// intRandomizer is the package-local view of the public IntRandomizer
// interfaces (defined in mx-chain-core-go/core/closing and
// mx-chain-go/dataRetriever). ISSUE-045 changed Intn to return
// `(int, error)` so entropy-source failures can be handled rather
// than panicking; this local interface mirrors that.
type intRandomizer interface {
	Intn(n int) (int, error)
}

// FisherYatesShuffle will shuffle the provided indexes slice based on a provided randomizer.
//
// ISSUE-045: returns an error if the randomizer fails on any of the
// (len(indexes) - 1) Intn calls. On error the partially-shuffled slice
// is returned alongside the error so callers can fall back to it
// (deterministic but non-empty) if they choose, or discard it if the
// distribution matters.
func FisherYatesShuffle(indexes []int, randomizer intRandomizer) ([]int, error) {
	newIndexes := make([]int, len(indexes))
	copy(newIndexes, indexes)

	for i := len(newIndexes) - 1; i > 0; i-- {
		j, err := randomizer.Intn(i + 1)
		if err != nil {
			return newIndexes, err
		}
		newIndexes[i], newIndexes[j] = newIndexes[j], newIndexes[i]
	}

	return newIndexes, nil
}
