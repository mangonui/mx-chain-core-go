package random

import (
	"crypto/rand"
	"fmt"
	"math/big"
)

// ConcurrentSafeIntRandomizer implements both
// `mx-chain-core-go/core/closing.IntRandomizer` and
// `mx-chain-go/dataRetriever.IntRandomizer` and can be accessed
// concurrently. ISSUE-045: the interface signature is
// `Intn(n int) (int, error)` so callers can react to entropy-source
// failures (e.g. /dev/urandom unavailable in containers, FIPS-mode
// kernel restrictions) without crashing the process.
type ConcurrentSafeIntRandomizer struct {
}

// Intn returns an int in [0, n) interval. Uses crypto/rand.Int so the
// distribution is uniform (no modulo bias) and any failure from the
// entropy source is surfaced explicitly instead of silently returning
// a deterministic value derived from an all-zero buffer.
//
// For n <= 0, returns (0, nil) — preserves the historical contract
// callers rely on (e.g. randomizer.Intn(len(emptySlice)) returning 0
// without erroring).
func (csir *ConcurrentSafeIntRandomizer) Intn(n int) (int, error) {
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
