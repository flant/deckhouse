/*
Copyright 2021 Flant CJSC

Licensed under the Apache License, Version 2.0 (the "License");
you may not use this file except in compliance with the License.
You may obtain a copy of the License at

    http://www.apache.org/licenses/LICENSE-2.0

Unless required by applicable law or agreed to in writing, software
distributed under the License is distributed on an "AS IS" BASIS,
WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
See the License for the specific language governing permissions and
limitations under the License.
*/

package util

import (
	"context"
	"sync"
	"time"
)

// DoWithTimer runs jobCb in background and waits until it is done. When timerDuration
// is passed and job is not done yet, onTimerCb is executed.
func DoWithTimer(interval time.Duration, jobCb, onTimerCb func()) {
	timer := time.NewTimer(interval)
	defer timer.Stop()

	// Start job in background
	doneCh := make(chan struct{})
	go func() {
		jobCb()
		close(doneCh)
	}()

	// Wait for closed doneCh or for timeout signal.
	for {
		select {
		case <-timer.C:
			onTimerCb()
		case <-doneCh:
			return
		}
	}
}

// ItemHandler should return 1 on fail, any other value is success
type ItemHandler func(ctx context.Context, index int, item string) int

// TimeoutHandler triggers the item handler execution times out
type TimeoutHandler func(idx int, item string)

// SequentialDoWithTimer starts itemHandler for each item in array.
// If itemHandler works more then timerDuration, then itemHandler for next
// item is started. If itemHandler returns 1, then all handlers are stopped via context.
func SequentialDoWithTimer(
	parentCtx context.Context,
	period time.Duration,
	items []string,
	handleItem ItemHandler,
	handleTimeout TimeoutHandler,
) {
	wg := sync.WaitGroup{}
	wg.Add(len(items))

	ctx, cancel := context.WithCancel(parentCtx)
	defer cancel()

	for index, item := range items {
		go doOne(ctx, cancel, &wg, period, handleItem, handleTimeout, index, item)
	}

	wg.Wait()
}

func doOne(
	// runtime
	ctx context.Context,
	cancel context.CancelFunc,
	wg *sync.WaitGroup,
	// configuration
	period time.Duration,
	handleItem ItemHandler,
	handleTimeout TimeoutHandler,
	// data
	index int,
	item string,
) {
	delayTimer := time.NewTimer(time.Duration(index) * period)
	defer func() {
		wg.Done()
		delayTimer.Stop()
	}()

	select {
	case <-ctx.Done():
		return
	case <-delayTimer.C:
	}

	DoWithTimer(period, func() {
		result := handleItem(ctx, index, item)
		if result == 1 {
			cancel()
		}
	}, func() {
		handleTimeout(index, item)
	})
}
