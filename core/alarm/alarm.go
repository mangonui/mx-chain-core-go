package alarm

import (
	"context"
	"sync"
	"time"
)

type eventType int

const (
	add eventType = iota
	cancel
	reset
)

const timeoutNoAlarm = time.Second * 100000
const toleranceExpiry = time.Millisecond * 10

type alarmEvent struct {
	alarmID string
	event   eventType
	alarm   *alarmItem
}

type alarmItem struct {
	initialDuration   time.Duration
	remainingDuration time.Duration
	callback          func(string)
}

type alarmScheduler struct {
	cancelFunc         context.CancelFunc
	scheduledAlarms    map[string]*alarmItem
	event              chan alarmEvent
	mutScheduledAlarms sync.RWMutex
}

// NewAlarmScheduler creates a new alarm scheduler instance and starts it's process loop
func NewAlarmScheduler() *alarmScheduler {
	as := &alarmScheduler{
		cancelFunc:      nil,
		scheduledAlarms: make(map[string]*alarmItem),
		event:           make(chan alarmEvent),
	}
	ctx, cancelFunc := context.WithCancel(context.Background())
	as.cancelFunc = cancelFunc

	go as.startProcessLoop(ctx)

	return as
}

// Add adds a new alarm to the alarm scheduler
func (as *alarmScheduler) Add(callback func(alarmID string), duration time.Duration, alarmID string) {
	alarm := &alarmItem{
		initialDuration:   duration,
		remainingDuration: duration,
		callback:          callback,
	}

	evt := alarmEvent{
		alarmID: alarmID,
		alarm:   alarm,
		event:   add,
	}

	as.event <- evt
}

// Cancel cancels a scheduled alarm.
// The cancel event is always sent into the event loop and the loop decides
// whether the alarm exists, eliminating the TOCTOU window between the prior
// map pre-check and the channel send. handleCancel safely no-ops via map
// delete when the alarm is absent.
func (as *alarmScheduler) Cancel(alarmID string) {
	evt := alarmEvent{
		alarmID: alarmID,
		alarm:   nil,
		event:   cancel,
	}

	as.event <- evt
}

func (as *alarmScheduler) startProcessLoop(ctx context.Context) {
	waitTime := timeoutNoAlarm
	var startTime time.Time

	for {
		startTime = time.Now()

		select {
		case <-ctx.Done():
			return
		case evt := <-as.event:
			elapsedTime := time.Since(startTime)
			waitTime = as.handleEvent(evt, elapsedTime)

		case <-time.After(waitTime):
			waitTime = as.updateAlarms(waitTime)
		}
	}
}

func (as *alarmScheduler) handleEvent(evt alarmEvent, elapsedSinceLastUpdate time.Duration) time.Duration {
	var waitTime time.Duration
	switch evt.event {
	case add:
		waitTime = as.handleAdd(elapsedSinceLastUpdate, evt.alarm, evt.alarmID)
	case cancel:
		waitTime = as.handleCancel(elapsedSinceLastUpdate, evt.alarmID)
	case reset:
		waitTime = as.handleReset(elapsedSinceLastUpdate, evt.alarmID)
	default:
		waitTime = as.updateAlarms(elapsedSinceLastUpdate)
	}

	return waitTime
}

func (as *alarmScheduler) handleAdd(
	elapsedSinceLastUpdate time.Duration,
	alarm *alarmItem,
	alarmID string,
) time.Duration {
	waitTime := as.updateAlarms(elapsedSinceLastUpdate)

	as.mutScheduledAlarms.Lock()
	as.scheduledAlarms[alarmID] = alarm
	as.mutScheduledAlarms.Unlock()

	if waitTime > alarm.remainingDuration {
		waitTime = alarm.remainingDuration
	}

	return waitTime
}

func (as *alarmScheduler) handleCancel(elapsedSinceLastUpdate time.Duration, alarmID string) time.Duration {
	as.mutScheduledAlarms.Lock()
	delete(as.scheduledAlarms, alarmID)
	as.mutScheduledAlarms.Unlock()

	return as.updateAlarms(elapsedSinceLastUpdate)
}

// handleReset restarts the alarm's countdown to its initial duration.
// The alarm's remainingDuration is pre-inflated by elapsedSinceLastUpdate so
// that the subsequent updateAlarms decrement leaves it at exactly
// initialDuration — mirroring handleAdd's "insert after updateAlarms" trick.
// If the alarm is not present the reset is a no-op.
func (as *alarmScheduler) handleReset(elapsedSinceLastUpdate time.Duration, alarmID string) time.Duration {
	as.mutScheduledAlarms.Lock()
	if alarm, ok := as.scheduledAlarms[alarmID]; ok {
		alarm.remainingDuration = alarm.initialDuration + elapsedSinceLastUpdate
	}
	as.mutScheduledAlarms.Unlock()

	return as.updateAlarms(elapsedSinceLastUpdate)
}

// updateAlarms updates the remaining duration for all alarms and returns the remaining minimum duration
func (as *alarmScheduler) updateAlarms(elapsed time.Duration) time.Duration {
	minDuration := timeoutNoAlarm

	as.mutScheduledAlarms.Lock()
	defer as.mutScheduledAlarms.Unlock()

	for alarmID, alarm := range as.scheduledAlarms {
		if alarm.remainingDuration <= elapsed+toleranceExpiry {
			go alarm.callback(alarmID)
			delete(as.scheduledAlarms, alarmID)
		} else {
			alarm.remainingDuration -= elapsed
			if minDuration > alarm.remainingDuration {
				minDuration = alarm.remainingDuration
			}
		}
	}

	return minDuration
}

// Close closes the alarm scheduler stopping the process loop
func (as *alarmScheduler) Close() {
	as.cancelFunc()
}

// Reset resets the alarm with the given id.
// The reset event is processed atomically inside the event loop, so the
// (alarmID -> alarmItem) lookup and the timer restart happen under the same
// serialization point as Add and Cancel. This eliminates the previous
// TOCTOU race where the alarm could expire (or be replaced) between the
// outside-the-loop map read and the channel sends.
func (as *alarmScheduler) Reset(alarmID string) {
	evt := alarmEvent{
		alarmID: alarmID,
		alarm:   nil,
		event:   reset,
	}

	as.event <- evt
}

// IsInterfaceNil returns true if interface is nil
func (as *alarmScheduler) IsInterfaceNil() bool {
	return as == nil
}
