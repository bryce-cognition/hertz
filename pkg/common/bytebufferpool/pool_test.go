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

package bytebufferpool

import (
	"fmt"
	"math/rand"
	"sync"
	"sync/atomic"
	"testing"
	"time"
)

func TestIndex(t *testing.T) {
	testIndex(t, 0, 0)
	testIndex(t, 1, 0)

	testIndex(t, minSize-1, 0)
	testIndex(t, minSize, 0)
	testIndex(t, minSize+1, 1)

	testIndex(t, 2*minSize-1, 1)
	testIndex(t, 2*minSize, 1)
	testIndex(t, 2*minSize+1, 2)

	testIndex(t, maxSize-1, steps-1)
	testIndex(t, maxSize, steps-1)
	testIndex(t, maxSize+1, steps-1)
}

func testIndex(t *testing.T, n, expectedIdx int) {
	idx := index(n)
	if idx != expectedIdx {
		t.Fatalf("unexpected idx for n=%d: %d. Expecting %d", n, idx, expectedIdx)
	}
}

func TestPoolCalibrate(t *testing.T) {
	for i := 0; i < steps*calibrateCallsThreshold; i++ {
		n := 1004
		if i%15 == 0 {
			n = rand.Intn(15234)
		}
		testGetPut(t, n)
	}
}

func TestPoolVariousSizesSerial(t *testing.T) {
	testPoolVariousSizes(t)
}

func TestPoolVariousSizesConcurrent(t *testing.T) {
	concurrency := 5
	ch := make(chan struct{})
	for i := 0; i < concurrency; i++ {
		go func() {
			testPoolVariousSizes(t)
			ch <- struct{}{}
		}()
	}
	for i := 0; i < concurrency; i++ {
		select {
		case <-ch:
		case <-time.After(10 * time.Second):
			t.Fatalf("timeout")
		}
	}
}

func testPoolVariousSizes(t *testing.T) {
	for i := 0; i < steps+1; i++ {
		n := 1 << uint32(i)

		testGetPut(t, n)
		testGetPut(t, n+1)
		testGetPut(t, n-1)

		for j := 0; j < 10; j++ {
			testGetPut(t, j+n)
		}
	}
}

func testGetPut(t *testing.T, n int) {
	bb := Get()
	if len(bb.B) > 0 {
		t.Fatalf("non-empty byte buffer returned from acquire")
	}
	bb.B = allocNBytes(bb.B, n)
	Put(bb)
}

func allocNBytes(dst []byte, n int) []byte {
	diff := n - cap(dst)
	if diff <= 0 {
		return dst[:n]
	}
	return append(dst, make([]byte, diff)...)
}

func TestConcurrentCalibration(t *testing.T) {
	testCases := []struct {
		name        string
		concurrency int
		bufferSizes []int
		iterations  int
	}{
		{"MixedBuffers", 10, []int{1, 64, 512, 4096, minSize - 1, minSize, minSize + 1, maxSize - 1, maxSize, maxSize + 1}, calibrateCallsThreshold},
		{"LargeBuffers", 5, []int{1 << 20, 1 << 22, 1 << 24}, calibrateCallsThreshold},
		{"SmallBuffers", 15, []int{1, 2, 4, 8, 16, 32}, calibrateCallsThreshold},
		{"ExtremeBuffers", 5, []int{0, 1, maxSize - 1, maxSize, maxSize + 1, 1 << 30}, calibrateCallsThreshold},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			p := &Pool{}
			var wg sync.WaitGroup

			// Simulate concurrent usage
			for i := 0; i < tc.concurrency; i++ {
				wg.Add(1)
				go func() {
					defer wg.Done()
					for j := 0; j < tc.iterations; j++ {
						bb := p.Get()
						bb.B = allocNBytes(bb.B, tc.bufferSizes[rand.Intn(len(tc.bufferSizes))])
						p.Put(bb)
					}
				}()
			}

			wg.Wait()

			// Wait for calibration to complete
			waitForCalibration(t, p)

			// Check if calibration has occurred and values are set correctly
			newDefaultSize := atomic.LoadUint64(&p.defaultSize)
			newMaxSize := atomic.LoadUint64(&p.maxSize)

			t.Logf("Calibration results: defaultSize=%d, maxSize=%d", newDefaultSize, newMaxSize)

			if newDefaultSize == 0 || newMaxSize == 0 {
				t.Errorf("Expected defaultSize and maxSize to be set after calibration")
			}

			if newDefaultSize > newMaxSize {
				t.Errorf("Expected defaultSize (%d) to be <= maxSize (%d)", newDefaultSize, newMaxSize)
			}

			if newDefaultSize < minSize || newDefaultSize > maxSize {
				t.Errorf("Expected defaultSize (%d) to be between %d and %d", newDefaultSize, minSize, maxSize)
			}

			// Verify that the pool returns buffers of the correct size
			for i := 0; i < 100; i++ {
				bb := p.Get()
				if uint64(cap(bb.B)) != newDefaultSize {
					t.Errorf("Expected Get() to return buffer with capacity %d, but got %d", newDefaultSize, cap(bb.B))
				}
				p.Put(bb)
			}

			// Verify calls array is reset after calibration
			callsSum := uint64(0)
			for i := 0; i < steps; i++ {
				calls := atomic.LoadUint64(&p.calls[i])
				callsSum += calls
				t.Logf("Calls for size %d: %d", minSize<<i, calls)
			}
			if callsSum != 0 {
				t.Errorf("Expected calls array to be reset after calibration, but sum is %d", callsSum)
			}

			// Verify index function
			for _, size := range tc.bufferSizes {
				idx := index(size)
				if idx < 0 || idx >= steps {
					t.Errorf("Invalid index %d for size %d", idx, size)
				}
			}

			// Test stability of calibration results
			initialDefaultSize := newDefaultSize
			initialMaxSize := newMaxSize

			// Perform additional operations to potentially trigger recalibration
			for i := 0; i < calibrateCallsThreshold; i++ {
				bb := p.Get()
				bb.B = allocNBytes(bb.B, tc.bufferSizes[rand.Intn(len(tc.bufferSizes))])
				p.Put(bb)
			}

			waitForCalibration(t, p)

			finalDefaultSize := atomic.LoadUint64(&p.defaultSize)
			finalMaxSize := atomic.LoadUint64(&p.maxSize)

			if finalDefaultSize != initialDefaultSize || finalMaxSize != initialMaxSize {
				t.Errorf("Calibration results changed unexpectedly. Initial: (%d, %d), Final: (%d, %d)",
					initialDefaultSize, initialMaxSize, finalDefaultSize, finalMaxSize)
			}
		})
	}

	// Test calibration when it's already in progress
	t.Run("CalibrationInProgress", func(t *testing.T) {
		p := &Pool{}
		atomic.StoreUint64(&p.calibrating, 1) // Set calibrating flag

		var wg sync.WaitGroup
		for i := 0; i < 5; i++ {
			wg.Add(1)
			go func() {
				defer wg.Done()
				for j := 0; j < calibrateCallsThreshold / 5; j++ {
					bb := p.Get()
					bb.B = allocNBytes(bb.B, rand.Intn(maxSize))
					p.Put(bb)
				}
			}()
		}

		wg.Wait()

		// Verify that calibration values didn't change
		if p.defaultSize != 0 || p.maxSize != 0 {
			t.Errorf("Calibration occurred when it shouldn't have: defaultSize=%d, maxSize=%d", p.defaultSize, p.maxSize)
		}

		// Reset calibrating flag and run one more time to ensure calibration occurs
		atomic.StoreUint64(&p.calibrating, 0)
		bb := p.Get()
		bb.B = allocNBytes(bb.B, maxSize)
		p.Put(bb)

		waitForCalibration(t, p)

		if p.defaultSize == 0 || p.maxSize == 0 {
			t.Errorf("Calibration did not occur after resetting calibrating flag")
		}
	})

	// Test concurrent calibration attempts
	t.Run("ConcurrentCalibrationAttempts", func(t *testing.T) {
		p := &Pool{}
		var wg sync.WaitGroup
		concurrency := 5

		for i := 0; i < concurrency; i++ {
			wg.Add(1)
			go func() {
				defer wg.Done()
				p.calibrate()
			}()
		}

		wg.Wait()

		if atomic.LoadUint64(&p.calibrating) != 0 {
			t.Errorf("Calibration flag should be reset to 0 after concurrent calibration attempts")
		}

		if p.defaultSize == 0 || p.maxSize == 0 {
			t.Errorf("Calibration did not occur during concurrent calibration attempts")
		}
	})

	// Test callSizes sorting and percentile calculation
	t.Run("CallSizesSortingAndPercentile", func(t *testing.T) {
		p := &Pool{}

		// Set up mock call counts
		mockCalls := []uint64{1000, 100, 10000, 500, 5000}
		for i, calls := range mockCalls {
			atomic.StoreUint64(&p.calls[i], calls)
		}

		p.calibrate()

		// Check if defaultSize and maxSize are set correctly
		if p.defaultSize == 0 || p.maxSize == 0 {
			t.Errorf("Calibration failed to set defaultSize or maxSize")
		}

		// Verify that defaultSize is set to the size with the highest call count
		expectedDefaultSize := minSize << 2 // Index 2 has the highest call count (10000)
		if p.defaultSize != uint64(expectedDefaultSize) {
			t.Errorf("Expected defaultSize to be %d, but got %d", expectedDefaultSize, p.defaultSize)
		}

		// Verify that maxSize is set correctly based on the 95th percentile
		expectedMaxSize := minSize << 4 // Index 4 is the 95th percentile
		if p.maxSize != uint64(expectedMaxSize) {
			t.Errorf("Expected maxSize to be %d, but got %d", expectedMaxSize, p.maxSize)
		}
	})
}

func waitForCalibration(t *testing.T, p *Pool) {
	start := time.Now()
	for time.Since(start) < 5*time.Second {
		if atomic.LoadUint64(&p.calibrating) == 0 {
			return
		}
		time.Sleep(10 * time.Millisecond)
	}
	t.Errorf("Calibration did not complete within the expected time")
}

func TestGetWithDifferentDefaultSizes(t *testing.T) {
	testSizes := []uint64{0, 64, 1024, 8192, maxSize, maxSize + 1}
	for _, size := range testSizes {
		t.Run(fmt.Sprintf("DefaultSize_%d", size), func(t *testing.T) {
			p := &Pool{}
			atomic.StoreUint64(&p.defaultSize, size)
			actualDefaultSize := atomic.LoadUint64(&p.defaultSize)
			t.Logf("Input defaultSize: %d, Actual defaultSize in Pool: %d", size, actualDefaultSize)

			bb := p.Get()
			expectedSize := size
			if size == 0 {
				expectedSize = minSize
			} else if size > maxSize {
				expectedSize = maxSize
			}
			actualCapacity := uint64(cap(bb.B))
			t.Logf("Expected capacity: %d, Actual capacity: %d", expectedSize, actualCapacity)

			if actualCapacity != expectedSize {
				t.Errorf("For defaultSize %d: Expected capacity %d, but got %d (actual defaultSize: %d)",
					size, expectedSize, actualCapacity, actualDefaultSize)
			}
			p.Put(bb)

			// Test multiple Get() calls
			for i := 0; i < 5; i++ {
				bb := p.Get()
				actualCapacity := uint64(cap(bb.B))
				t.Logf("Iteration %d - Expected capacity: %d, Actual capacity: %d", i, expectedSize, actualCapacity)

				if actualCapacity != expectedSize {
					t.Errorf("For defaultSize %d (iteration %d): Expected capacity %d, but got %d (actual defaultSize: %d)",
						size, i, expectedSize, actualCapacity, actualDefaultSize)
				}
				p.Put(bb)
			}

			// Verify final state of the pool
			finalDefaultSize := atomic.LoadUint64(&p.defaultSize)
			t.Logf("Final defaultSize in Pool: %d", finalDefaultSize)
			if finalDefaultSize != actualDefaultSize {
				t.Errorf("Pool defaultSize changed unexpectedly. Initial: %d, Final: %d",
					actualDefaultSize, finalDefaultSize)
			}
		})
	}
}

func TestPutEdgeCases(t *testing.T) {
	p := &Pool{maxSize: 1024}

	// Test with maxSize of 0
	p.maxSize = 0
	bb := p.Get()
	bb.B = make([]byte, 2048)
	p.Put(bb)

	// Test with very large buffer capacity
	p.maxSize = 1024
	bb = p.Get()
	bb.B = make([]byte, 4096)
	p.Put(bb)

	// Test with buffer size equal to maxSize
	bb = p.Get()
	bb.B = make([]byte, 1024)
	p.Put(bb)

	// Verify that Put doesn't panic in these edge cases
}

func TestIndexExpanded(t *testing.T) {
	testCases := []struct {
		input    int
		expected int
	}{
		{0, 0},
		{1, 0},
		{63, 0},
		{64, 0},
		{65, 1},
		{127, 1},
		{128, 1},
		{129, 2},
		{maxSize - 1, steps - 1},
		{maxSize, steps - 1},
		{maxSize + 1, steps - 1},
		{1 << 30, steps - 1}, // Test with a very large number
	}

	for _, tc := range testCases {
		result := index(tc.input)
		if result != tc.expected {
			t.Errorf("index(%d) = %d, expected %d", tc.input, result, tc.expected)
		}
	}
}

func TestByteBufferReset(t *testing.T) {
	bb := Get()
	bb.WriteString("test data")

	if len(bb.B) == 0 {
		t.Error("ByteBuffer should contain data before reset")
	}

	bb.Reset()

	if len(bb.B) != 0 {
		t.Error("ByteBuffer should be empty after reset")
	}

	if cap(bb.B) == 0 {
		t.Error("ByteBuffer capacity should not be zero after reset")
	}

	Put(bb)
}
