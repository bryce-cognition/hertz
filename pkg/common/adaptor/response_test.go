/*
 * Copyright 2023 CloudWeGo Authors
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

package adaptor

import (
	"net/http"
	"testing"

	"github.com/cloudwego/hertz/pkg/common/test/assert"
	"github.com/cloudwego/hertz/pkg/protocol"
	"github.com/cloudwego/hertz/pkg/protocol/consts"
)

func TestGetCompatResponseWriter(t *testing.T) {
	resp := &protocol.Response{}
	compatResp := GetCompatResponseWriter(resp).(*compatResponse)

	assert.DeepEqual(t, resp, compatResp.h)
	assert.NotNil(t, compatResp.header)
	assert.False(t, compatResp.writeHeader)
}

func TestCompatResponse_Header(t *testing.T) {
	resp := &protocol.Response{}
	compatResp := GetCompatResponseWriter(resp).(*compatResponse)

	// Test getting headers when none are set (header should be initialized)
	header := compatResp.Header()
	assert.NotNil(t, header)
	assert.DeepEqual(t, 0, len(header))

	// Test that subsequent calls to Header() return the same instance
	header2 := compatResp.Header()
	assert.DeepEqual(t, header, header2)

	// Test setting a single header
	header.Set("Test-Key", "Test-Value")
	assert.DeepEqual(t, "Test-Value", header.Get("Test-Key"))

	// Test setting multiple headers
	header.Set("Another-Key", "Another-Value")
	header.Add("Multi-Key", "Value1")
	header.Add("Multi-Key", "Value2")
	assert.DeepEqual(t, "Another-Value", header.Get("Another-Key"))
	assert.DeepEqual(t, []string{"Value1", "Value2"}, header["Multi-Key"])

	// Test overwriting existing header
	header.Set("Test-Key", "New-Value")
	assert.DeepEqual(t, "New-Value", header.Get("Test-Key"))

	// Test handling of special headers
	header.Set("Content-Type", "application/json")
	assert.DeepEqual(t, "application/json", header.Get("Content-Type"))

	// Verify that the underlying response headers are not modified before WriteHeader is called
	assert.DeepEqual(t, 0, resp.Header.Len())

	// Call WriteHeader to sync headers with the underlying response
	compatResp.WriteHeader(consts.StatusOK)

	// Verify all headers are present in the underlying response
	assert.DeepEqual(t, "New-Value", string(resp.Header.Peek("Test-Key")))
	assert.DeepEqual(t, "Another-Value", string(resp.Header.Peek("Another-Key")))
	assert.DeepEqual(t, "Value1", string(resp.Header.Peek("Multi-Key")))
	assert.DeepEqual(t, "application/json", string(resp.Header.Peek("Content-Type")))

	// Test calling WriteHeader multiple times
	compatResp.WriteHeader(consts.StatusNotFound)
	assert.DeepEqual(t, consts.StatusNotFound, resp.StatusCode())

	// Verify that headers are not duplicated when WriteHeader is called multiple times
	assert.DeepEqual(t, 1, len(resp.Header.PeekAll("Test-Key")))
}

func TestCompatResponse_Write(t *testing.T) {
	resp := &protocol.Response{}
	compatResp := GetCompatResponseWriter(resp).(*compatResponse)

	testBody := []byte("Test body")
	n, err := compatResp.Write(testBody)

	assert.Nil(t, err)
	assert.DeepEqual(t, len(testBody), n)
	assert.DeepEqual(t, testBody, resp.Body())
	assert.True(t, compatResp.writeHeader)
	assert.DeepEqual(t, consts.StatusOK, resp.StatusCode())
}

func TestCompatResponse_WriteHeaderInResponse(t *testing.T) {
	resp := &protocol.Response{}
	compatResp := GetCompatResponseWriter(resp).(*compatResponse)

	compatResp.Header().Set("Test-Key", "Test-Value")
	compatResp.Header().Set(consts.HeaderSetCookie, "test-cookie=value")
	compatResp.WriteHeader(consts.StatusCreated)

	assert.True(t, compatResp.writeHeader)
	assert.DeepEqual(t, consts.StatusCreated, resp.StatusCode())
	assert.DeepEqual(t, "Test-Value", string(resp.Header.Peek("Test-Key")))

	// Check if the cookie was set correctly
	cookie := protocol.AcquireCookie()
	resp.Header.VisitAllCookie(func(key, value []byte) {
		cookie.Parse(string(value))
	})
	assert.DeepEqual(t, "test-cookie", string(cookie.Key()))
	assert.DeepEqual(t, "value", string(cookie.Value()))
}

func TestCompatResponse_WriteHeaderTwice(t *testing.T) {
	resp := &protocol.Response{}
	compatResp := GetCompatResponseWriter(resp).(*compatResponse)

	compatResp.WriteHeader(consts.StatusCreated)
	compatResp.WriteHeader(consts.StatusOK)

	// The status code should change after the second call to WriteHeader
	assert.DeepEqual(t, consts.StatusOK, resp.StatusCode())
}

func TestCompatResponse_HeaderAlreadyInitialized(t *testing.T) {
	resp := &protocol.Response{}
	compatResp := GetCompatResponseWriter(resp).(*compatResponse)

	// Verify that the header is initially an empty map
	assert.DeepEqual(t, http.Header{}, compatResp.Header())

	// Initialize the header
	initialHeader := compatResp.Header()
	assert.NotNil(t, initialHeader)
	assert.DeepEqual(t, http.Header{}, initialHeader)
	initialHeader.Set("Initial-Key", "Initial-Value")

	// Call Header() again
	header := compatResp.Header()

	// Verify that the header is the same instance
	assert.DeepEqual(t, initialHeader, header)

	// Verify that the initial value is still present
	assert.DeepEqual(t, "Initial-Value", header.Get("Initial-Key"))

	// Add a new header
	header.Set("New-Key", "New-Value")

	// Verify both headers are present
	assert.DeepEqual(t, "Initial-Value", header.Get("Initial-Key"))
	assert.DeepEqual(t, "New-Value", header.Get("New-Key"))

	// Verify the underlying response has no headers yet
	assert.DeepEqual(t, 0, resp.Header.Len())

	// Call WriteHeader to sync headers with the underlying response
	compatResp.WriteHeader(consts.StatusOK)

	// Verify headers are now present in the underlying response
	assert.DeepEqual(t, "Initial-Value", string(resp.Header.Peek("Initial-Key")))
	assert.DeepEqual(t, "New-Value", string(resp.Header.Peek("New-Key")))
}

func TestCompatResponse_HeaderInitiallyNil(t *testing.T) {
	resp := &protocol.Response{}
	compatResp := GetCompatResponseWriter(resp).(*compatResponse)

	// Ensure the header is initially nil
	compatResp.header = nil

	// Call Header() to initialize the header
	header := compatResp.Header()

	// Verify that the header is not nil and is an empty map
	assert.NotNil(t, header)
	assert.DeepEqual(t, http.Header{}, header)

	// Add a header
	header.Set("Test-Key", "Test-Value")

	// Verify the header was added
	assert.DeepEqual(t, "Test-Value", header.Get("Test-Key"))

	// Call Header() again
	header2 := compatResp.Header()

	// Verify that it returns the same instance
	assert.DeepEqual(t, header, header2)

	// Verify the underlying response has no headers yet
	assert.DeepEqual(t, 0, resp.Header.Len())

	// Call WriteHeader to sync headers with the underlying response
	compatResp.WriteHeader(consts.StatusOK)

	// Verify header is now present in the underlying response
	assert.DeepEqual(t, "Test-Value", string(resp.Header.Peek("Test-Key")))
}
