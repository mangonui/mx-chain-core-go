package random

import (
	"crypto/rand"
	"fmt"
	"math/big"
)

// ConcurrentSafeIntRandomizer implements dataRetriever.IntRandomizer and can be accessed in a concurrent manner
type ConcurrentSafeIntRandomizer struct {
}

func (csir *ConcurrentSafeIntRandomizer) Intn(n int) int {
	val, err := csir.IntnWithError(n)
	if err != nil {
		panic(err)
	}

	return val
}

// IntnWithError returns an int in [0, n) interval.
// Uses crypto/rand.Int so the distribution is uniform (no modulo bias) and any
// failure from the entropy source is surfaced explicitly instead of silently
// returning a deterministic value derived from an all-zero buffer.
func (csir *ConcurrentSafeIntRandomizer) IntnWithError(n int) (int, error) {
	if n <= 0 {
		return 0, nil
	}

	val, err := rand.Int(rand.Reader, big.NewInt(int64(n)))
	if err != nil {
		return 0, fmt.Errorf("cannot generate crypto-random int: %w", err)
	}

	return int(val.Int64()), nil
}

// IsInterfaceNil returns true if there is no value under the interface
func (csir *ConcurrentSafeIntRandomizer) IsInterfaceNil() bool {
	return csir == nil
}
