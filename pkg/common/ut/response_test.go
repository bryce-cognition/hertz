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
 */

package ut

import (
	"testing"

	"github.com/cloudwego/hertz/pkg/common/test/assert"
	"github.com/cloudwego/hertz/pkg/protocol/consts"
)

func TestResult(t *testing.T) {
	r := new(ResponseRecorder)
	ret := r.Result()
	assert.DeepEqual(t, consts.StatusOK, ret.StatusCode())
}

func TestFlush(t *testing.T) {
	r := new(ResponseRecorder)
	r.Flush()
	ret := r.Result()
	assert.DeepEqual(t, consts.StatusOK, ret.StatusCode())
}

func TestWriterHeader(t *testing.T) {
	r := NewRecorder()
	r.WriteHeader(consts.StatusCreated)
	r.WriteHeader(consts.StatusOK)
	ret := r.Result()
	assert.DeepEqual(t, consts.StatusCreated, ret.StatusCode())
}

func TestWriteString(t *testing.T) {
	r := NewRecorder()
	r.WriteString("hello")
	ret := r.Result()
	assert.DeepEqual(t, "hello", string(ret.Body()))
}

func TestWrite(t *testing.T) {
	r := NewRecorder()
	r.Write([]byte("hello"))
	ret := r.Result()
	assert.DeepEqual(t, "hello", string(ret.Body()))
}

func TestHeader(t *testing.T) {
	r := NewRecorder()
	r.Header().Set("Content-Type", "application/json")
	r.Header().Set("X-Custom-Header", "test-value")
	ret := r.Result()
	assert.DeepEqual(t, "application/json", string(ret.Header.Get("Content-Type")))
	assert.DeepEqual(t, "test-value", string(ret.Header.Get("X-Custom-Header")))
}

func TestMultipleWrites(t *testing.T) {
	r := NewRecorder()
	r.Write([]byte("Hello"))
	r.Write([]byte(" "))
	r.Write([]byte("World"))
	ret := r.Result()
	assert.DeepEqual(t, "Hello World", string(ret.Body()))
}

func TestLargePayload(t *testing.T) {
	r := NewRecorder()
	largePayload := make([]byte, 1024*1024) // 1MB payload
	for i := range largePayload {
		largePayload[i] = byte(i % 256)
	}
	r.Write(largePayload)
	ret := r.Result()
	assert.DeepEqual(t, largePayload, ret.Body())
	assert.DeepEqual(t, len(largePayload), ret.Header.ContentLength())
}

func TestDifferentStatusCodes(t *testing.T) {
	statusCodes := []int{
		consts.StatusOK,
		consts.StatusCreated,
		consts.StatusAccepted,
		consts.StatusNoContent,
		consts.StatusBadRequest,
		consts.StatusInternalServerError,
	}

	for _, code := range statusCodes {
		r := NewRecorder()
		r.WriteHeader(code)
		ret := r.Result()
		assert.DeepEqual(t, code, ret.StatusCode())
	}
}
