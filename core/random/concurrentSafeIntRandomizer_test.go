package random

import (
	"fmt"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

func TestConcurrentSafeIntRandomizer_IntnConcurrent(t *testing.T) {
	csir := &ConcurrentSafeIntRandomizer{}

	defer func() {
		r := recover()
		if r != nil {
			assert.Fail(t, fmt.Sprintf("randomizer fail: %v", t))
		}
	}()

	maxIterations := 100
	for i := 0; i < maxIterations; i++ {
		go func(idx int) {
			for {
				_, _ = csir.Intn(idx + 1)
				time.Sleep(time.Millisecond)
			}
		}(i)
	}

	fmt.Println("Waiting 1 second...")
	time.Sleep(time.Second * 1)
}

func TestConcurrentSafeIntRandomizer_IntnInvalidShouldReturnZero(t *testing.T) {
	t.Parallel()

	csir := &ConcurrentSafeIntRandomizer{}
	assert.False(t, csir.IsInterfaceNil())

	res, err := csir.Intn(-1)
	assert.NoError(t, err)
	assert.Equal(t, 0, res)

	res, err = csir.Intn(0)
	assert.NoError(t, err)
	assert.Equal(t, 0, res)

	// n=1 returns 0 trivially because crypto/rand.Int(1) = 0.
	res, err = csir.Intn(1)
	assert.NoError(t, err)
	assert.Equal(t, 0, res)
}

func TestConcurrentSafeIntRandomizer_IntnShouldWork(t *testing.T) {
	t.Parallel()

	csir := &ConcurrentSafeIntRandomizer{}
	assert.False(t, csir.IsInterfaceNil())

	maxValue := 70
	res, err := csir.Intn(maxValue)

	assert.NoError(t, err)
	assert.True(t, res >= 0, fmt.Sprintf("0 comparison, generated %d, max %d", res, maxValue))
	assert.True(t, res < maxValue, fmt.Sprintf("max comparison, generated %d, max %d", res, maxValue))
}
