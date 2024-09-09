/*
 * Copyright 2022 CloudWeGo Authors
 *
 * Licensed under the Apache License, Version 2.0 (the "License");
 * you may not use this file except in compliance with the License.
 * You may obtain a copy of the License at
 *
 *     http://www.apache.org/licenses/LICENSE-2.0
 *
 * Unless required by applicable law or agreed to in writing, software
 * distributed under the License is distributed on an "AS IS" BASIS,
 * WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
 * See the License for the specific language governing permissions and
 * limitations under the License.
 *
 * The MIT License (MIT)
 *
 * Copyright (c) 2015-present Aliaksandr Valialkin, VertaMedia, Kirill Danshin, Erik Dubbelboer, FastHTTP Authors
 *
 * Permission is hereby granted, free of charge, to any person obtaining a copy
 * of this software and associated documentation files (the "Software"), to deal
 * in the Software without restriction, including without limitation the rights
 * to use, copy, modify, merge, publish, distribute, sublicense, and/or sell
 * copies of the Software, and to permit persons to whom the Software is
 * furnished to do so, subject to the following conditions:
 *
 * The above copyright notice and this permission notice shall be included in
 * all copies or substantial portions of the Software.
 *
 * THE SOFTWARE IS PROVIDED "AS IS", WITHOUT WARRANTY OF ANY KIND, EXPRESS OR
 * IMPLIED, INCLUDING BUT NOT LIMITED TO THE WARRANTIES OF MERCHANTABILITY,
 * FITNESS FOR A PARTICULAR PURPOSE AND NONINFRINGEMENT. IN NO EVENT SHALL THE
 * AUTHORS OR COPYRIGHT HOLDERS BE LIABLE FOR ANY CLAIM, DAMAGES OR OTHER
 * LIABILITY, WHETHER IN AN ACTION OF CONTRACT, TORT OR OTHERWISE, ARISING FROM,
 * OUT OF OR IN CONNECTION WITH THE SOFTWARE OR THE USE OR OTHER DEALINGS IN
 * THE SOFTWARE.
 *
 * This file may have been modified by CloudWeGo authors. All CloudWeGo
 * Modifications are Copyright 2022 CloudWeGo Authors.
 */

package timer

import (
	"sync"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

// test initTimer function
func TestTimerInitTimer(t *testing.T) {
	// test nil Timer
	var nilTimer *time.Timer
	resNilTime := initTimer(nilTimer, 2*time.Second)
	assert.NotNil(t, resNilTime, "Expecting a Timer, got nil")

	// test the panic
	panicTimer := time.NewTimer(1 * time.Second)
	resPanicTimer := wrapInitTimer(panicTimer, 2*time.Second)
	assert.Equal(t, -1, resPanicTimer, "Expecting a panic for Timer")

	// test with different timeout durations
	shortTimer := initTimer(nil, 100*time.Millisecond)
	assert.NotNil(t, shortTimer, "Expecting a Timer for short duration")
	longTimer := initTimer(nil, 5*time.Second)
	assert.NotNil(t, longTimer, "Expecting a Timer for long duration")
}

func wrapInitTimer(t *time.Timer, timeout time.Duration) (ret int) {
	defer func() {
		if err := recover(); err != nil {
			ret = -1
		}
	}()
	res := initTimer(t, timeout)
	if res != nil {
		ret = 1
	}
	return ret
}

func TestTimerStopTimer(t *testing.T) {
	normalTimer := time.NewTimer(3 * time.Second)
	stopTimer(normalTimer)
	assert.False(t, normalTimer.Stop(), "Expecting timer to be stopped")

	// Test stopping an already expired timer
	expiredTimer := time.NewTimer(1 * time.Millisecond)
	time.Sleep(2 * time.Millisecond)
	stopTimer(expiredTimer)
	assert.False(t, expiredTimer.Stop(), "Expecting expired timer to be stopped")
}

func TestTimerAcquireTimer(t *testing.T) {
	normalTimer := AcquireTimer(2 * time.Second)
	assert.NotNil(t, normalTimer, "Expecting a timer, got nil")
	ReleaseTimer(normalTimer)

	// Test acquiring multiple timers
	timer1 := AcquireTimer(1 * time.Second)
	timer2 := AcquireTimer(2 * time.Second)
	assert.NotEqual(t, timer1, timer2, "Acquired timers should be different")
	ReleaseTimer(timer1)
	ReleaseTimer(timer2)
}

func TestTimerReleaseTimer(t *testing.T) {
	normalTimer := AcquireTimer(2 * time.Second)
	ReleaseTimer(normalTimer)
	assert.False(t, normalTimer.Stop(), "Expecting the timer to be released")

	// Test releasing the same timer multiple times
	ReleaseTimer(normalTimer)
	assert.False(t, normalTimer.Stop(), "Expecting no panic when releasing the same timer multiple times")
}

func TestConcurrentTimerUsage(t *testing.T) {
	const goroutines = 100
	var wg sync.WaitGroup
	wg.Add(goroutines)

	for i := 0; i < goroutines; i++ {
		go func() {
			defer wg.Done()
			timer := AcquireTimer(time.Duration(i) * time.Millisecond)
			time.Sleep(1 * time.Millisecond)
			ReleaseTimer(timer)
		}()
	}

	wg.Wait()
	// If we reach here without deadlock or panic, the test passes
	assert.True(t, true, "Concurrent timer usage should not cause deadlock or panic")
}
