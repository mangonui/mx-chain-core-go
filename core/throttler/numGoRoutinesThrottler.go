package throttler

import (
	"sync/atomic"

	"github.com/multiversx/mx-chain-core-go/core"
)

// NumGoRoutinesThrottler can limit the number of go routines launched
type NumGoRoutinesThrottler struct {
	max     int32
	counter int32
}

// NewNumGoRoutinesThrottler creates a new num go routine throttler instance
func NewNumGoRoutinesThrottler(max int32) (*NumGoRoutinesThrottler, error) {
	if max <= 0 {
		return nil, core.ErrNotPositiveValue
	}

	return &NumGoRoutinesThrottler{
		max: max,
	}, nil
}

// CanProcess returns true if current counter is less than max
func (ngrt *NumGoRoutinesThrottler) CanProcess() bool {
	valCounter := atomic.LoadInt32(&ngrt.counter)

	return valCounter < ngrt.max
}

// StartProcessing will increment current counter
func (ngrt *NumGoRoutinesThrottler) StartProcessing() {
	atomic.AddInt32(&ngrt.counter, 1)
}

// TryStartProcessing increments the counter only if capacity is available.
func (ngrt *NumGoRoutinesThrottler) TryStartProcessing() bool {
	for {
		current := atomic.LoadInt32(&ngrt.counter)
		if current >= ngrt.max {
			return false
		}
		if atomic.CompareAndSwapInt32(&ngrt.counter, current, current+1) {
			return true
		}
	}
}

// EndProcessing will decrement current counter
func (ngrt *NumGoRoutinesThrottler) EndProcessing() {
	atomic.AddInt32(&ngrt.counter, -1)
}

// IsInterfaceNil returns true if there is no value under the interface
func (ngrt *NumGoRoutinesThrottler) IsInterfaceNil() bool {
	return ngrt == nil
}
