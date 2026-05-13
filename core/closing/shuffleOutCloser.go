package closing

import (
	"context"
	"fmt"
	"time"

	"github.com/multiversx/mx-chain-core-go/core"
	"github.com/multiversx/mx-chain-core-go/core/check"
	"github.com/multiversx/mx-chain-core-go/core/random"
	"github.com/multiversx/mx-chain-core-go/data/endProcess"
)

const minDuration = time.Second

type shuffleOutCloser struct {
	minWaitDuration time.Duration
	maxWaitDuration time.Duration
	signalChan      chan endProcess.ArgEndProcess
	randomizer      IntRandomizer
	log             core.Logger
	ctx             context.Context
	cancelFunc      func()
}

// NewShuffleOutCloser creates a shuffle out component that is able to trigger a node restart and cancel that request if necessarily
func NewShuffleOutCloser(
	minWaitDuration time.Duration,
	maxWaitDuration time.Duration,
	signalChan chan endProcess.ArgEndProcess,
	log core.Logger,
) (*shuffleOutCloser, error) {

	if minWaitDuration < minDuration {
		return nil, fmt.Errorf("%w for minWaitDuration", core.ErrInvalidValue)
	}
	if maxWaitDuration < minDuration {
		return nil, fmt.Errorf("%w for maxWaitDuration", core.ErrInvalidValue)
	}
	if minWaitDuration > maxWaitDuration {
		return nil, fmt.Errorf("%w, minWaitDuration > maxWaitDuration", core.ErrInvalidValue)
	}
	if signalChan == nil {
		return nil, core.ErrNilSignalChan
	}
	if check.IfNil(log) {
		return nil, core.ErrNilLogger
	}

	soc := &shuffleOutCloser{
		minWaitDuration: minWaitDuration,
		maxWaitDuration: maxWaitDuration,
		signalChan:      signalChan,
		randomizer:      &random.ConcurrentSafeIntRandomizer{},
		log:             log,
	}
	soc.ctx, soc.cancelFunc = context.WithCancel(context.Background())

	return soc, nil
}

// EndOfProcessingHandler will be called each time a delayed end of processing is needed
func (soc *shuffleOutCloser) EndOfProcessingHandler(event endProcess.ArgEndProcess) error {
	go soc.writeOnChanDelayed(event)

	return nil
}

func (soc *shuffleOutCloser) writeOnChanDelayed(event endProcess.ArgEndProcess) {
	delta := soc.maxWaitDuration - soc.minWaitDuration

	// ISSUE-045: on entropy failure, fall back to the deterministic
	// midpoint of the wait window. The shutdown-delay path is not
	// security-critical (it just spreads node-restart timing to avoid
	// thundering-herd reconnects); a deterministic wait is correct
	// behaviour, just less effective at the herd-spreading goal.
	// Logging the error makes the deployment-side cause visible
	// (entropy starvation in the container) without crashing the
	// shutdown path.
	randDurationBeforeStop, err := soc.randomizer.Intn(int(delta))
	if err != nil {
		soc.log.Warn("shuffleOutCloser: entropy failure on shutdown-delay randomization, "+
			"falling back to deterministic midpoint",
			"error", err)
		randDurationBeforeStop = int(delta) / 2
	}
	timeToWait := soc.minWaitDuration + time.Duration(randDurationBeforeStop)

	soc.log.Info("the application will stop in",
		"waiting time", fmt.Sprintf("%v", timeToWait),
		"description", event.Description,
		"reason", event.Reason)

	select {
	case <-time.After(timeToWait):
	case <-soc.ctx.Done():
		soc.log.Debug("canceled the application stop go routine")
		return
	}

	soc.log.Info("the application will stop now after",
		"waiting time", fmt.Sprintf("%v", timeToWait),
		"description", event.Description,
		"reason", event.Reason,
	)

	soc.signalChan <- event
}

// Close cancels the channel write
func (soc *shuffleOutCloser) Close() error {
	soc.cancelFunc()

	return nil
}

// IsInterfaceNil returns true if there is no value under the interface
func (soc *shuffleOutCloser) IsInterfaceNil() bool {
	return soc == nil
}
