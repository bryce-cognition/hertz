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

package adaptor

import (
	"strings"
	"testing"

	"github.com/cloudwego/hertz/pkg/common/test/assert"
	"github.com/cloudwego/hertz/pkg/protocol"
	"github.com/cloudwego/hertz/pkg/protocol/consts"
)

func TestCompatResponse_Header(t *testing.T) {
	resp := &protocol.Response{}
	cr := &CompatResponse{h: resp}

	// Test initial header creation
	header := cr.Header()
	assert.NotNil(t, header)
	assert.DeepEqual(t, 0, len(header))

	// Test header reuse
	header["Test-Key"] = []string{"Test-Value"}
	newHeader := cr.Header()
	assert.DeepEqual(t, header, newHeader)
}

func TestCompatResponse_Write(t *testing.T) {
	resp := &protocol.Response{}
	cr := &CompatResponse{h: resp}

	// Test writing without calling WriteHeader first
	n, err := cr.Write([]byte("test"))
	assert.Nil(t, err)
	assert.DeepEqual(t, 4, n)
	assert.DeepEqual(t, consts.StatusOK, resp.Header.StatusCode())
	assert.DeepEqual(t, []byte("test"), resp.Body())

	// Test writing after calling WriteHeader
	cr.WriteHeader(consts.StatusCreated)
	n, err = cr.Write([]byte(" body"))
	assert.Nil(t, err)
	assert.DeepEqual(t, 5, n)
	assert.DeepEqual(t, consts.StatusCreated, resp.Header.StatusCode())
	assert.DeepEqual(t, []byte("test body"), resp.Body())
}

func TestCompatResponse_WriteHeaderAndStatusCode(t *testing.T) {
	resp := &protocol.Response{}
	cr := &CompatResponse{h: resp}

	// Test setting headers and status code
	cr.Header().Set("Content-Type", "text/plain")
	cr.Header().Set("X-Custom-Header", "test")
	cr.WriteHeader(consts.StatusNotFound)

	assert.DeepEqual(t, consts.StatusNotFound, resp.Header.StatusCode())
	assert.DeepEqual(t, "text/plain", string(resp.Header.Peek("Content-Type")))
	assert.DeepEqual(t, "test", string(resp.Header.Peek("X-Custom-Header")))

	// Test that calling WriteHeader again doesn't change the status code
	cr.WriteHeader(consts.StatusOK)
	assert.DeepEqual(t, consts.StatusNotFound, resp.Header.StatusCode())

	// Test that setting headers after WriteHeader doesn't affect the response
	cr.Header().Set("X-Late-Header", "late-value")
	assert.DeepEqual(t, "", string(resp.Header.Peek("X-Late-Header")))
}

func TestCompatResponse_SetCookie(t *testing.T) {
	resp := &protocol.Response{}
	cr := &CompatResponse{h: resp}

	cookieStr := "session=123; Path=/; HttpOnly"
	cr.Header().Set("Set-Cookie", cookieStr)
	cr.WriteHeader(consts.StatusOK)

	setCookie := resp.Header.Peek("Set-Cookie")
	assert.NotNil(t, setCookie)
	assert.DeepEqual(t, cookieStr, string(setCookie))

	// Verify that multiple cookies are handled correctly
	cr.Header().Add("Set-Cookie", "user=john; Secure")
	cr.WriteHeader(consts.StatusOK) // This should not affect the already set cookies

	setCookies := resp.Header.PeekAll("Set-Cookie")
	assert.DeepEqual(t, 2, len(setCookies))
	assert.DeepEqual(t, cookieStr, string(setCookies[0]))
	assert.DeepEqual(t, "user=john; Secure", string(setCookies[1]))
}

func TestGetCompatResponseWriter(t *testing.T) {
	resp := &protocol.Response{}
	resp.Header.Set("X-Existing-Header", "value")
	cookie := &protocol.Cookie{}
	cookie.SetKey("session")
	cookie.SetValue("123")
	resp.Header.SetCookie(cookie)

	writer := GetCompatResponseWriter(resp)
	assert.NotNil(t, writer)

	compatResp, ok := writer.(*CompatResponse)
	assert.True(t, ok)
	assert.NotNil(t, compatResp.h)
	assert.NotNil(t, compatResp.header)

	// Check if existing headers are transferred
	assert.DeepEqual(t, "value", compatResp.Header().Get("X-Existing-Header"))

	// Check if SetCookie is properly handled
	cookies := compatResp.Header()["Set-Cookie"]
	assert.DeepEqual(t, 1, len(cookies))
	assert.True(t, strings.Contains(cookies[0], "session=123"))

	// Check if NoDefaultContentType is set
	assert.True(t, resp.Header.NoDefaultContentType())

	// Test writing to the response
	_, err := writer.Write([]byte("test body"))
	assert.Nil(t, err)
	assert.DeepEqual(t, []byte("test body"), resp.Body())

	// Test setting a new header before WriteHeader
	writer.Header().Set("X-New-Header", "new value")
	writer.WriteHeader(consts.StatusOK)

	// Verify that the new header is set in the underlying response
	assert.DeepEqual(t, "new value", resp.Header.Get("X-New-Header"))

	// Test setting a header after WriteHeader (should not be reflected)
	writer.Header().Set("X-Late-Header", "late value")
	assert.DeepEqual(t, "", resp.Header.Get("X-Late-Header"))
}

func TestCompatResponse_MultipleWriteHeader(t *testing.T) {
	resp := &protocol.Response{}
	cr := &CompatResponse{h: resp}

	cr.WriteHeader(consts.StatusOK)
	assert.DeepEqual(t, consts.StatusOK, resp.Header.StatusCode())

	// Subsequent WriteHeader calls should not change the status code
	cr.WriteHeader(consts.StatusNotFound)
	assert.DeepEqual(t, consts.StatusOK, resp.Header.StatusCode())

	// Test writing to body after multiple WriteHeader calls
	n, err := cr.Write([]byte("test body"))
	assert.Nil(t, err)
	assert.DeepEqual(t, 9, n)
	assert.DeepEqual(t, []byte("test body"), resp.Body())

	// Ensure status code remains unchanged
	assert.DeepEqual(t, consts.StatusOK, resp.Header.StatusCode())

	// Test that even after writing, additional WriteHeader calls don't change the status
	cr.WriteHeader(consts.StatusInternalServerError)
	assert.DeepEqual(t, consts.StatusOK, resp.Header.StatusCode())
}

func TestCompatResponse_HeaderManipulation(t *testing.T) {
	resp := &protocol.Response{}
	cr := &CompatResponse{h: resp}

	// Test adding multiple values for the same header
	cr.Header().Add("X-Multi-Header", "value1")
	cr.Header().Add("X-Multi-Header", "value2")
	cr.WriteHeader(consts.StatusOK)

	multiHeader := resp.Header.PeekAll("X-Multi-Header")
	assert.DeepEqual(t, 2, len(multiHeader))
	assert.DeepEqual(t, "value1", string(multiHeader[0]))
	assert.DeepEqual(t, "value2", string(multiHeader[1]))

	// Test deleting a header after WriteHeader (should not affect the response)
	cr.Header().Del("X-Multi-Header")
	assert.DeepEqual(t, 2, len(resp.Header.PeekAll("X-Multi-Header")))

	// Test setting a new header after WriteHeader (should not be reflected in the response)
	cr.Header().Set("X-Late-Header", "late-value")
	assert.DeepEqual(t, "", string(resp.Header.Peek("X-Late-Header")))

	// Test that manipulating headers before WriteHeader works as expected
	newResp := &protocol.Response{}
	newCr := &CompatResponse{h: newResp}
	newCr.Header().Set("X-Before-Write", "before-value")
	newCr.WriteHeader(consts.StatusOK)
	assert.DeepEqual(t, "before-value", string(newResp.Header.Peek("X-Before-Write")))
}

func TestGetCompatResponseWriter_EdgeCases(t *testing.T) {
	resp := &protocol.Response{}
	resp.Header.Set("X-Existing-Header", "value1")
	resp.Header.Add("X-Existing-Header", "value2")

	writer := GetCompatResponseWriter(resp)
	compatResp, _ := writer.(*CompatResponse)

	// Check if multiple values for the same header are preserved
	existingHeader := compatResp.Header()["X-Existing-Header"]
	assert.DeepEqual(t, 2, len(existingHeader))
	assert.DeepEqual(t, "value1", existingHeader[0])
	assert.DeepEqual(t, "value2", existingHeader[1])

	// Test handling of special headers
	resp.Header.SetContentLength(100)
	writer = GetCompatResponseWriter(resp)
	compatResp, _ = writer.(*CompatResponse)
	assert.DeepEqual(t, "", compatResp.Header().Get("Content-Length"))

	// Verify Content-Length is not transferred to CompatResponse
	assert.DeepEqual(t, 0, len(compatResp.Header()["Content-Length"]))

	// Test NoDefaultContentType is set
	assert.True(t, resp.Header.NoDefaultContentType())

	// Test that Set-Cookie headers are properly handled
	cookie := &protocol.Cookie{}
	cookie.SetKey("session")
	cookie.SetValue("123")
	resp.Header.SetCookie(cookie)
	writer = GetCompatResponseWriter(resp)
	compatResp, _ = writer.(*CompatResponse)
	assert.DeepEqual(t, 1, len(compatResp.Header()["Set-Cookie"]))
	assert.True(t, strings.Contains(compatResp.Header()["Set-Cookie"][0], "session=123"))
}
