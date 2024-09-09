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
	"math/rand"
	"sync"
	"sync/atomic"
	"testing"
	"time"

	"github.com/cloudwego/hertz/pkg/common/test/assert"
)

	"github.com/cloudwego/hertz/pkg/common/test/assert"
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

func TestPoolGet(t *testing.T) {
	p := &Pool{}
	b := p.Get()
	assert.NotNil(t, b)
	assert.Assert(t, len(b.B) == 0, "Get() returned non-empty buffer: %d", len(b.B))
}

func TestPoolPut(t *testing.T) {
	p := &Pool{}
	b := p.Get()
	b.B = append(b.B, []byte("test")...)
	p.Put(b)

	// Check if the buffer is reused
	b2 := p.Get()
	assert.Assert(t, len(b2.B) == 0, "Put() didn't reset buffer: %d", len(b2.B))
}

func TestPoolConcurrent(t *testing.T) {
	p := &Pool{}
	concurrency := 10
	iterations := 5000 // Increased iterations to trigger calibration

	var wg sync.WaitGroup
	wg.Add(concurrency)

	for i := 0; i < concurrency; i++ {
		go func() {
			defer wg.Done()
			for j := 0; j < iterations; j++ {
				b := p.Get()
				b.B = append(b.B, []byte("test")...)
				p.Put(b)
				if j%100 == 0 {
					time.Sleep(time.Millisecond) // Add small delay to allow calibration
				}
			}
		}()
	}

	wg.Wait()

	// Wait for calibration to complete
	for atomic.LoadUint64(&p.calibrating) != 0 {
		time.Sleep(time.Millisecond)
	}

	// Check if calibration occurred after concurrent usage
	assert.Assert(t, atomic.LoadUint64(&p.defaultSize) != 0, "Calibration didn't update defaultSize after concurrent usage")
	assert.Assert(t, atomic.LoadUint64(&p.maxSize) != 0, "Calibration didn't update maxSize after concurrent usage")
}

func TestPoolVariousSizes(t *testing.T) {
	p := &Pool{}
	sizes := []int{64, 128, 256, 512, 1024, 2048, 4096}

	for _, size := range sizes {
		b := p.Get()
		b.B = make([]byte, size)
		p.Put(b)
	}

	// Get a buffer after putting various sizes
	b := p.Get()
	assert.Assert(t, cap(b.B) >= sizes[0], "Expected minimum capacity of %d, got %d", sizes[0], cap(b.B))
	p.Put(b)
}

func TestPoolEdgeCases(t *testing.T) {
	p := &Pool{}

	// Test with very small buffer
	b := p.Get()
	b.B = append(b.B, []byte("a")...)
	p.Put(b)

	// Test with very large buffer
	b = p.Get()
	b.B = make([]byte, 1<<20) // 1MB
	p.Put(b)

	// Test multiple Get() and Put() operations
	for i := 0; i < 1000; i++ {
		b := p.Get()
		b.B = append(b.B, []byte("test")...)
		p.Put(b)
	}

	// Verify that the pool is still functioning correctly
	b = p.Get()
	assert.NotNil(t, b)
	assert.Assert(t, len(b.B) == 0, "Get() returned non-empty buffer after multiple operations")
}

func TestPoolCalibration(t *testing.T) {
	p := &Pool{}

	// Trigger calibration
	for i := 0; i < calibrateCallsThreshold+1; i++ {
		b := p.Get()
		b.B = make([]byte, i%maxSize)
		p.Put(b)
	}

	// Wait for calibration to complete
	for atomic.LoadUint64(&p.calibrating) != 0 {
		time.Sleep(time.Millisecond)
	}

	assert.Assert(t, atomic.LoadUint64(&p.defaultSize) != 0, "Calibration didn't update defaultSize")
	assert.Assert(t, atomic.LoadUint64(&p.maxSize) != 0, "Calibration didn't update maxSize")
	assert.Assert(t, atomic.LoadUint64(&p.maxSize) >= atomic.LoadUint64(&p.defaultSize), "maxSize should be >= defaultSize")
}
