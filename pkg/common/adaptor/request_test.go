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
	"bytes"
	"context"
	"fmt"
	"io/ioutil"
	"net/http"
	"net/url"
	"strings"
	"testing"
	"time"

	"github.com/cloudwego/hertz/pkg/app"
	"github.com/cloudwego/hertz/pkg/app/server"
	"github.com/cloudwego/hertz/pkg/common/test/assert"
	"github.com/cloudwego/hertz/pkg/protocol"
	"github.com/cloudwego/hertz/pkg/protocol/consts"
)

func TestCompatResponse_WriteHeader(t *testing.T) {
	var testHeader http.Header
	var testBody string
	testUrl1 := "http://127.0.0.1:9000/test1"
	testUrl2 := "http://127.0.0.1:9000/test2"
	testStatusCode := 299
	testCookieValue := "cookie"

	testHeader = make(map[string][]string)
	testHeader["Key1"] = []string{"value1"}
	testHeader["Key2"] = []string{"value2", "value22"}
	testHeader["Key3"] = []string{"value3", "value33", "value333"}
	testHeader[consts.HeaderSetCookie] = []string{testCookieValue}

	testBody = "test body"

	h := server.New(server.WithHostPorts("127.0.0.1:9000"))
	h.POST("/test1", func(c context.Context, ctx *app.RequestContext) {
		req, _ := GetCompatRequest(&ctx.Request)
		resp := GetCompatResponseWriter(&ctx.Response)
		handlerAndCheck(t, resp, req, testHeader, testBody, testStatusCode)
	})

	h.POST("/test2", func(c context.Context, ctx *app.RequestContext) {
		req, _ := GetCompatRequest(&ctx.Request)
		resp := GetCompatResponseWriter(&ctx.Response)
		handlerAndCheck(t, resp, req, testHeader, testBody)
	})

	go h.Spin()
	time.Sleep(200 * time.Millisecond)

	makeACall(t, http.MethodPost, testUrl1, testHeader, testBody, testStatusCode, []byte(testCookieValue))
	makeACall(t, http.MethodPost, testUrl2, testHeader, testBody, consts.StatusOK, []byte(testCookieValue))
}

func makeACall(t *testing.T, method, url string, header http.Header, body string, expectStatusCode int, expectCookieValue []byte) {
	client := http.Client{}
	req, _ := http.NewRequest(method, url, strings.NewReader(body))
	req.Header = header
	resp, err := client.Do(req)
	if err != nil {
		t.Fatalf("make a call error: %s", err)
	}

	respHeader := resp.Header

	for k, v := range header {
		for i := 0; i < len(v); i++ {
			if respHeader[k][i] != v[i] {
				t.Fatalf("Header error: want %s=%s, got %s=%s", respHeader[k], respHeader[k][i], respHeader[k], v[i])
			}
		}
	}

	b, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		t.Fatalf("Read body error: %s", err)
	}
	assert.DeepEqual(t, body, string(b))
	assert.DeepEqual(t, expectStatusCode, resp.StatusCode)

	// Parse out the cookie to verify it is correct
	cookie := protocol.Cookie{}
	_ = cookie.Parse(header[consts.HeaderSetCookie][0])
	assert.DeepEqual(t, expectCookieValue, cookie.Value())
}

// handlerAndCheck is designed to handle the program and check the header
//
// "..." is used in the type of statusCode, which is a syntactic sugar in Go.
// In this way, the statusCode can be made an optional parameter,
// and there is no need to pass in some meaningless numbers to judge some special cases.
func handlerAndCheck(t *testing.T, writer http.ResponseWriter, request *http.Request, wantHeader http.Header, wantBody string, statusCode ...int) {
	reqHeader := request.Header
	for k, v := range wantHeader {
		if reqHeader[k] == nil {
			t.Fatalf("Header error: want %s=%s, got %s=nil", reqHeader[k], reqHeader[k][0], reqHeader[k])
		}
		if reqHeader[k][0] != v[0] {
			t.Fatalf("Header error: want %s=%s, got %s=%s", reqHeader[k], reqHeader[k][0], reqHeader[k], v[0])
		}
	}

	body, err := ioutil.ReadAll(request.Body)
	if err != nil {
		t.Fatalf("Read body error: %s", err)
	}
	assert.DeepEqual(t, wantBody, string(body))

	respHeader := writer.Header()
	for k, v := range reqHeader {
		respHeader[k] = v
	}

	// When the incoming status code is nil, the execution of this code is skipped
	// and the status code is set to 200
	if statusCode != nil {
		writer.WriteHeader(statusCode[0])
	}

	_, err = writer.Write([]byte("test"))
	if err != nil {
		t.Fatalf("Write body error: %s", err)
	}
	_, err = writer.Write([]byte(" body"))
	if err != nil {
		t.Fatalf("Write body error: %s", err)
	}
}

func TestCopyToHertzRequest(t *testing.T) {
	req := http.Request{
		Method:     "GET",
		RequestURI: "/test",
		URL: &url.URL{
			Scheme: "http",
			Host:   "test.com",
		},
		Proto:  "HTTP/1.1",
		Header: http.Header{},
	}
	req.Header.Set("key1", "value1")
	req.Header.Add("key2", "value2")
	req.Header.Add("key2", "value22")
	hertzReq := protocol.Request{}
	err := CopyToHertzRequest(&req, &hertzReq)
	assert.Nil(t, err)
	assert.DeepEqual(t, req.Method, string(hertzReq.Method()))
	assert.DeepEqual(t, req.RequestURI, string(hertzReq.Path()))
	assert.DeepEqual(t, req.Proto, hertzReq.Header.GetProtocol())
	assert.DeepEqual(t, req.Header.Get("key1"), hertzReq.Header.Get("key1"))
	valueSlice := make([]string, 0, 2)
	hertzReq.Header.VisitAllCustomHeader(func(key, value []byte) {
		if strings.ToLower(string(key)) == "key2" {
			valueSlice = append(valueSlice, string(value))
		}
	})

	assert.DeepEqual(t, req.Header.Values("key2"), valueSlice)

	assert.DeepEqual(t, 3, hertzReq.Header.Len())
}

func TestGetCompatRequestErrorHandling(t *testing.T) {
	// Test with invalid method
	invalidReq := &protocol.Request{}
	invalidReq.SetMethod("INVALID_METHOD")
	httpReq, err := GetCompatRequest(invalidReq)
	assert.Nil(t, err)
	assert.NotNil(t, httpReq)
	assert.DeepEqual(t, "INVALID_METHOD", httpReq.Method)

	// Test with empty method
	emptyMethodReq := &protocol.Request{}
	httpReq, err = GetCompatRequest(emptyMethodReq)
	assert.Nil(t, err)
	assert.NotNil(t, httpReq)
	assert.DeepEqual(t, "GET", httpReq.Method) // Default method when method is empty

	// Test with nil request
	httpReq, err = GetCompatRequest(nil)
	assert.NotNil(t, err)
	assert.Nil(t, httpReq)
	assert.DeepEqual(t, "nil request", err.Error())
}

func TestCopyToHertzRequestEdgeCases(t *testing.T) {
	t.Run("Nil body", func(t *testing.T) {
		httpReq, _ := http.NewRequest("GET", "http://example.com", nil)
		hertzReq := &protocol.Request{}
		err := CopyToHertzRequest(httpReq, hertzReq)
		assert.Nil(t, err)
		assert.DeepEqual(t, int64(0), hertzReq.Header.ContentLength())
		assert.Nil(t, hertzReq.Body())
		assert.DeepEqual(t, 0, len(hertzReq.Body()))
		assert.DeepEqual(t, "GET", string(hertzReq.Method()))
		assert.DeepEqual(t, "http://example.com", string(hertzReq.URI().FullURI()))
		assert.DeepEqual(t, 2, hertzReq.Header.Len()) // Host and User-Agent headers are added
	})

	t.Run("Empty headers", func(t *testing.T) {
		httpReq, _ := http.NewRequest("GET", "http://example.com", nil)
		httpReq.Header = make(http.Header)
		hertzReq := &protocol.Request{}
		err := CopyToHertzRequest(httpReq, hertzReq)
		assert.Nil(t, err)
		assert.DeepEqual(t, 2, hertzReq.Header.Len()) // Host and RequestURI headers are added
		assert.DeepEqual(t, "GET", string(hertzReq.Method()))
		assert.DeepEqual(t, "http://example.com", string(hertzReq.URI().FullURI()))
	})

	t.Run("Multiple header values", func(t *testing.T) {
		httpReq, _ := http.NewRequest("GET", "http://example.com", nil)
		httpReq.Header.Add("Multi-Value", "value1")
		httpReq.Header.Add("Multi-Value", "value2")
		hertzReq := &protocol.Request{}
		err := CopyToHertzRequest(httpReq, hertzReq)
		assert.Nil(t, err)
		assert.DeepEqual(t, []string{"value1", "value2"}, hertzReq.Header.GetAll("Multi-Value"))
		assert.DeepEqual(t, 2, hertzReq.Header.Len()) // Multi-Value and Host headers
	})

	t.Run("Non-nil body", func(t *testing.T) {
		body := bytes.NewBufferString("test body")
		httpReq, _ := http.NewRequest("POST", "http://example.com", body)
		httpReq.Header.Set("Content-Length", "9")
		httpReq.Header.Set("Content-Type", "text/plain")
		hertzReq := &protocol.Request{}
		err := CopyToHertzRequest(httpReq, hertzReq)
		assert.Nil(t, err)
		assert.DeepEqual(t, int64(9), hertzReq.Header.ContentLength())
		assert.DeepEqual(t, []byte("test body"), hertzReq.Body())
		assert.DeepEqual(t, "POST", string(hertzReq.Method()))
		assert.DeepEqual(t, "text/plain", string(hertzReq.Header.ContentType())) // Content-Type is copied
		assert.DeepEqual(t, "http://example.com", string(hertzReq.URI().FullURI()))
		assert.DeepEqual(t, 3, hertzReq.Header.Len()) // Content-Length, Content-Type, and Host
	})

	t.Run("Content-Length header set to 0", func(t *testing.T) {
		httpReq, _ := http.NewRequest("POST", "http://example.com", nil)
		httpReq.Header.Set("Content-Length", "0")
		hertzReq := &protocol.Request{}
		err := CopyToHertzRequest(httpReq, hertzReq)
		assert.Nil(t, err)
		assert.DeepEqual(t, int64(0), hertzReq.Header.ContentLength())
		assert.Nil(t, hertzReq.Body())
		assert.DeepEqual(t, 2, hertzReq.Header.Len()) // Content-Length and Host
	})

	t.Run("Invalid Content-Length header", func(t *testing.T) {
		body := bytes.NewBufferString("test body")
		httpReq, _ := http.NewRequest("POST", "http://example.com", body)
		httpReq.Header.Set("Content-Length", "invalid")
		hertzReq := &protocol.Request{}
		err := CopyToHertzRequest(httpReq, hertzReq)
		assert.Nil(t, err)
		assert.DeepEqual(t, int64(0), hertzReq.Header.ContentLength()) // 0 indicates unknown length
		assert.DeepEqual(t, []byte("test body"), hertzReq.Body())
		assert.DeepEqual(t, 2, hertzReq.Header.Len()) // Content-Length and Host
	})

	t.Run("Empty body with non-zero Content-Length", func(t *testing.T) {
		httpReq, _ := http.NewRequest("POST", "http://example.com", nil)
		httpReq.Header.Set("Content-Length", "10")
		hertzReq := &protocol.Request{}
		err := CopyToHertzRequest(httpReq, hertzReq)
		assert.Nil(t, err)
		assert.DeepEqual(t, int64(10), hertzReq.Header.ContentLength())
		assert.Nil(t, hertzReq.Body())
		assert.DeepEqual(t, 2, hertzReq.Header.Len()) // Content-Length and Host
	})

	t.Run("Custom HTTP method", func(t *testing.T) {
		httpReq, _ := http.NewRequest("CUSTOM", "http://example.com", nil)
		hertzReq := &protocol.Request{}
		err := CopyToHertzRequest(httpReq, hertzReq)
		assert.Nil(t, err)
		assert.DeepEqual(t, "CUSTOM", string(hertzReq.Method()))
	})

	t.Run("URL with query parameters", func(t *testing.T) {
		httpReq, _ := http.NewRequest("GET", "http://example.com/path?param1=value1&param2=value2", nil)
		hertzReq := &protocol.Request{}
		err := CopyToHertzRequest(httpReq, hertzReq)
		assert.Nil(t, err)
		assert.DeepEqual(t, "http://example.com/path?param1=value1&param2=value2", string(hertzReq.URI().FullURI()))
		assert.DeepEqual(t, "value1", string(hertzReq.URI().QueryArgs().Peek("param1")))
		assert.DeepEqual(t, "value2", string(hertzReq.URI().QueryArgs().Peek("param2")))
		assert.DeepEqual(t, 1, hertzReq.Header.Len()) // Host header
	})

	t.Run("Large body", func(t *testing.T) {
		largeBody := bytes.Repeat([]byte("a"), 1024*1024) // 1MB body
		httpReq, _ := http.NewRequest("POST", "http://example.com", bytes.NewReader(largeBody))
		httpReq.Header.Set("Content-Length", fmt.Sprintf("%d", len(largeBody)))
		hertzReq := &protocol.Request{}
		err := CopyToHertzRequest(httpReq, hertzReq)
		assert.Nil(t, err)
		assert.DeepEqual(t, int64(len(largeBody)), hertzReq.Header.ContentLength())
		assert.DeepEqual(t, largeBody, hertzReq.Body())
		assert.DeepEqual(t, 2, hertzReq.Header.Len()) // Content-Length and Host
	})

	t.Run("Transfer-Encoding header", func(t *testing.T) {
		httpReq, _ := http.NewRequest("POST", "http://example.com", strings.NewReader("chunked body"))
		httpReq.Header.Set("Transfer-Encoding", "chunked")
		hertzReq := &protocol.Request{}
		err := CopyToHertzRequest(httpReq, hertzReq)
		assert.Nil(t, err)
		assert.DeepEqual(t, "chunked", string(hertzReq.Header.Peek("Transfer-Encoding")))
		assert.DeepEqual(t, []byte("chunked body"), hertzReq.Body())
		assert.DeepEqual(t, 1, hertzReq.Header.Len()) // Only Host header is added
		assert.DeepEqual(t, "http://example.com", string(hertzReq.URI().FullURI()))
	})

	t.Run("Non-standard port in URL", func(t *testing.T) {
		httpReq, _ := http.NewRequest("GET", "http://example.com:8080/path", nil)
		hertzReq := &protocol.Request{}
		err := CopyToHertzRequest(httpReq, hertzReq)
		assert.Nil(t, err)
		assert.DeepEqual(t, "/path", string(hertzReq.URI().Path()))
		assert.DeepEqual(t, "example.com:8080", string(hertzReq.URI().Host()))
		assert.DeepEqual(t, "http", string(hertzReq.URI().Scheme()))
		assert.DeepEqual(t, 1, hertzReq.Header.Len()) // Only Host header
	})

	t.Run("Relative URL", func(t *testing.T) {
		httpReq, _ := http.NewRequest("GET", "/relative/path", nil)
		httpReq.Host = "example.com"
		hertzReq := &protocol.Request{}
		err := CopyToHertzRequest(httpReq, hertzReq)
		assert.Nil(t, err)
		assert.DeepEqual(t, "/relative/path", string(hertzReq.URI().Path()))
		assert.DeepEqual(t, "example.com", string(hertzReq.URI().Host()))
		assert.DeepEqual(t, "/relative/path", string(hertzReq.URI().FullURI()))
	})

	t.Run("Empty URL", func(t *testing.T) {
		httpReq, _ := http.NewRequest("GET", "", nil)
		httpReq.Host = "example.com"
		hertzReq := &protocol.Request{}
		err := CopyToHertzRequest(httpReq, hertzReq)
		assert.Nil(t, err)
		assert.DeepEqual(t, "", string(hertzReq.URI().Path()))
		assert.DeepEqual(t, "example.com", string(hertzReq.URI().Host()))
		assert.DeepEqual(t, "", string(hertzReq.URI().FullURI()))
		assert.DeepEqual(t, 1, hertzReq.Header.Len()) // Only Host header
	})

	t.Run("Custom header", func(t *testing.T) {
		httpReq, _ := http.NewRequest("GET", "http://example.com", nil)
		httpReq.Header.Set("X-Custom-Header", "custom-value")
		hertzReq := &protocol.Request{}
		err := CopyToHertzRequest(httpReq, hertzReq)
		assert.Nil(t, err)
		assert.DeepEqual(t, "custom-value", string(hertzReq.Header.Peek("X-Custom-Header")))
		assert.DeepEqual(t, 2, hertzReq.Header.Len()) // X-Custom-Header and Host
	})

	t.Run("URL with special characters", func(t *testing.T) {
		httpReq, _ := http.NewRequest("GET", "http://example.com/path%20with%20spaces?q=hello%20world", nil)
		hertzReq := &protocol.Request{}
		err := CopyToHertzRequest(httpReq, hertzReq)
		assert.Nil(t, err)
		assert.DeepEqual(t, "http://example.com/path%20with%20spaces?q=hello%20world", string(hertzReq.URI().FullURI()))
		assert.DeepEqual(t, "/path%20with%20spaces", string(hertzReq.URI().Path()))
		assert.DeepEqual(t, "hello world", string(hertzReq.URI().QueryArgs().Peek("q")))
	})

	t.Run("Body without Content-Length header", func(t *testing.T) {
		body := bytes.NewBufferString("test body")
		httpReq, _ := http.NewRequest("POST", "http://example.com", body)
		hertzReq := &protocol.Request{}
		err := CopyToHertzRequest(httpReq, hertzReq)
		assert.Nil(t, err)
		assert.DeepEqual(t, int64(0), hertzReq.Header.ContentLength()) // 0 indicates unknown length
		assert.DeepEqual(t, []byte("test body"), hertzReq.Body())
		assert.DeepEqual(t, 1, hertzReq.Header.Len()) // Only Host header
	})

	t.Run("HTTPS URL", func(t *testing.T) {
		httpReq, _ := http.NewRequest("GET", "https://example.com/secure", nil)
		hertzReq := &protocol.Request{}
		err := CopyToHertzRequest(httpReq, hertzReq)
		assert.Nil(t, err)
		assert.DeepEqual(t, "https://example.com/secure", string(hertzReq.URI().FullURI()))
		assert.DeepEqual(t, "https", string(hertzReq.URI().Scheme()))
	})

	t.Run("URL with fragment", func(t *testing.T) {
		httpReq, _ := http.NewRequest("GET", "http://example.com/page#section1", nil)
		hertzReq := &protocol.Request{}
		err := CopyToHertzRequest(httpReq, hertzReq)
		assert.Nil(t, err)
		assert.DeepEqual(t, "http://example.com/page#section1", string(hertzReq.URI().FullURI()))
		assert.DeepEqual(t, "section1", string(hertzReq.URI().Hash()))
	})

	t.Run("URL with userinfo", func(t *testing.T) {
		httpReq, _ := http.NewRequest("GET", "http://user:pass@example.com/", nil)
		hertzReq := &protocol.Request{}
		err := CopyToHertzRequest(httpReq, hertzReq)
		assert.Nil(t, err)
		assert.DeepEqual(t, "http://user:pass@example.com/", string(hertzReq.URI().FullURI()))
		assert.True(t, strings.Contains(string(hertzReq.URI().FullURI()), "user:pass@"))
	})

	t.Run("Request with nil URL", func(t *testing.T) {
		httpReq := &http.Request{Method: "GET", URL: nil}
		hertzReq := &protocol.Request{}
		err := CopyToHertzRequest(httpReq, hertzReq)
		assert.Nil(t, err)
		assert.DeepEqual(t, "", string(hertzReq.URI().FullURI()))
	})

	t.Run("Request with nil Header", func(t *testing.T) {
		httpReq, _ := http.NewRequest("GET", "http://example.com", nil)
		httpReq.Header = nil
		hertzReq := &protocol.Request{}
		err := CopyToHertzRequest(httpReq, hertzReq)

		assert.Nil(t, err)
		assert.DeepEqual(t, 1, hertzReq.Header.Len()) // Host header is still added
	})

	t.Run("Request with non-nil Body but zero Content-Length", func(t *testing.T) {
		httpReq, _ := http.NewRequest("POST", "http://example.com", strings.NewReader(""))
		httpReq.ContentLength = 0
		hertzReq := &protocol.Request{}
		err := CopyToHertzRequest(httpReq, hertzReq)
		assert.Nil(t, err)
		assert.DeepEqual(t, int64(0), hertzReq.Header.ContentLength())
		assert.NotNil(t, hertzReq.Body())
		assert.DeepEqual(t, 0, len(hertzReq.Body()))
	})

	t.Run("Request with custom protocol version", func(t *testing.T) {
		httpReq, _ := http.NewRequest("GET", "http://example.com", nil)
		httpReq.Proto = "HTTP/2.0"
		hertzReq := &protocol.Request{}
		err := CopyToHertzRequest(httpReq, hertzReq)
		assert.Nil(t, err)
		assert.DeepEqual(t, "HTTP/2.0", string(hertzReq.Header.GetProtocol()))
	})

	t.Run("Request with nil URL and non-nil Header", func(t *testing.T) {
		httpReq := &http.Request{Method: "GET", URL: nil, Header: make(http.Header)}
		httpReq.Header.Set("X-Custom-Header", "custom-value")
		hertzReq := &protocol.Request{}
		err := CopyToHertzRequest(httpReq, hertzReq)
		assert.Nil(t, err)
		assert.DeepEqual(t, "", string(hertzReq.URI().FullURI()))
		assert.DeepEqual(t, "custom-value", string(hertzReq.Header.Peek("X-Custom-Header")))
		assert.DeepEqual(t, 1, hertzReq.Header.Len()) // Only X-Custom-Header
	})

	t.Run("Request with both Transfer-Encoding and Content-Length headers", func(t *testing.T) {
		body := bytes.NewBufferString("test body")
		httpReq, _ := http.NewRequest("POST", "http://example.com", body)
		httpReq.Header.Set("Transfer-Encoding", "chunked")
		httpReq.Header.Set("Content-Length", "9")
		hertzReq := &protocol.Request{}
		err := CopyToHertzRequest(httpReq, hertzReq)
		assert.Nil(t, err)
		assert.DeepEqual(t, "chunked", string(hertzReq.Header.Peek("Transfer-Encoding")))
		assert.DeepEqual(t, int64(9), hertzReq.Header.ContentLength())
		assert.DeepEqual(t, []byte("test body"), hertzReq.Body())
		assert.DeepEqual(t, 3, hertzReq.Header.Len()) // Transfer-Encoding, Content-Length, and Host
	})

	t.Run("Request with body but zero Content-Length", func(t *testing.T) {
		body := bytes.NewBufferString("test body")
		httpReq, _ := http.NewRequest("POST", "http://example.com", body)
		httpReq.ContentLength = 0
		hertzReq := &protocol.Request{}
		err := CopyToHertzRequest(httpReq, hertzReq)
		assert.Nil(t, err)
		assert.DeepEqual(t, int64(0), hertzReq.Header.ContentLength())
		assert.DeepEqual(t, []byte("test body"), hertzReq.Body())
		assert.DeepEqual(t, 1, hertzReq.Header.Len()) // Only Host header
	})
}
