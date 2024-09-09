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

package ut

import (
	"bytes"
	"testing"

	"github.com/cloudwego/hertz/pkg/common/test/assert"
)

func TestCreateUtRequestContext(t *testing.T) {
	t.Run("Basic functionality", func(t *testing.T) {
		body := "1"
		method := "PUT"
		path := "/hey/dy"
		headerKey := "Connection"
		headerValue := "close"
		ctx := CreateUtRequestContext(method, path, &Body{bytes.NewBufferString(body), len(body)},
			Header{headerKey, headerValue})

		assert.DeepEqual(t, method, string(ctx.Method()))
		assert.DeepEqual(t, path, string(ctx.Path()))
		body1, err := ctx.Body()
		assert.DeepEqual(t, nil, err)
		assert.DeepEqual(t, body, string(body1))
		assert.DeepEqual(t, headerValue, string(ctx.GetHeader(headerKey)))
	})

	t.Run("Empty body", func(t *testing.T) {
		ctx := CreateUtRequestContext("GET", "/empty", &Body{bytes.NewBuffer(nil), 0})
		body, err := ctx.Body()
		assert.DeepEqual(t, nil, err)
		assert.DeepEqual(t, 0, len(body))
	})

	t.Run("Multiple headers", func(t *testing.T) {
		ctx := CreateUtRequestContext("POST", "/multi-header",
			&Body{bytes.NewBufferString("test"), 4},
			Header{"Content-Type", "application/json"},
			Header{"X-Custom", "value"})

		assert.DeepEqual(t, "application/json", string(ctx.GetHeader("Content-Type")))
		assert.DeepEqual(t, "value", string(ctx.GetHeader("X-Custom")))
	})

	t.Run("Query parameters", func(t *testing.T) {
		ctx := CreateUtRequestContext("GET", "/query?param1=value1&param2=value2", nil)
		assert.DeepEqual(t, "value1", string(ctx.QueryArgs().Peek("param1")))
		assert.DeepEqual(t, "value2", string(ctx.QueryArgs().Peek("param2")))
	})
}

func TestUtRequestContextMethods(t *testing.T) {
	ctx := CreateUtRequestContext("POST", "/test", &Body{bytes.NewBufferString("body"), 4},
		Header{"Content-Type", "application/json"})

	t.Run("SetMethod", func(t *testing.T) {
		ctx.Request.SetMethod("GET")
		assert.DeepEqual(t, "GET", string(ctx.Method()))
	})

	t.Run("SetRequestURI", func(t *testing.T) {
		ctx.Request.SetRequestURI("/new-uri")
		assert.DeepEqual(t, "/new-uri", string(ctx.URI().Path()))
	})

	t.Run("SetBody", func(t *testing.T) {
		newBody := []byte("new body")
		ctx.Request.SetBody(newBody)
		body, err := ctx.Body()
		assert.DeepEqual(t, nil, err)
		assert.DeepEqual(t, newBody, body)
	})

	t.Run("SetUserValue", func(t *testing.T) {
		ctx.Set("key", "value")
		assert.DeepEqual(t, "value", ctx.Value("key"))
	})
}
