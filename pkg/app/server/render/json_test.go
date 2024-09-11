/*
 * Copyright 2022 CloudWeGo Authors
 *
 * Licensed under the Apache License, Version 2.0 (the "License");
 * you may not use this file except in compliance with the License.
 * You may obtain a copy of the License at
 *
 *  http://www.apache.org/licenses/LICENSE-2.0
 *
 * Unless required by applicable law or agreed to in writing, software
 * distributed under the License is distributed on an "AS IS" BASIS,
 * WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
 * See the License for the specific language governing permissions and
 * limitations under the License.
 *
 * The MIT License (MIT)
 *
 * Copyright (c) 2014 Manuel Martínez-Almeida
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

 * This file may have been modified by CloudWeGo authors. All CloudWeGo
 * Modifications are Copyright 2022 CloudWeGo Authors.
 */

package render

import (
	"fmt"
	"strings"
	"testing"

	"github.com/cloudwego/hertz/pkg/protocol"
	"github.com/stretchr/testify/assert"
)

func Test_ResetStdJSONMarshal(t *testing.T) {
	table := map[string]string{
		"testA": "hello",
		"B":     "world",
	}
	ResetStdJSONMarshal()
	jsonBytes, err := jsonMarshalFunc(table)
	if err != nil {
		t.Fatal(err)
	}
	if !strings.Contains(string(jsonBytes), "\"B\":\"world\"") || !strings.Contains(string(jsonBytes), "\"testA\":\"hello\"") {
		t.Fatal("marshal struct is not equal to the string")
	}
}

func Test_DefaultJSONMarshal(t *testing.T) {
	table := map[string]string{
		"testA": "hello",
		"B":     "world",
	}
	jsonBytes, err := jsonMarshalFunc(table)
	if err != nil {
		t.Fatal(err)
	}
	if !strings.Contains(string(jsonBytes), "\"B\":\"world\"") || !strings.Contains(string(jsonBytes), "\"testA\":\"hello\"") {
		t.Fatal("marshal struct is not equal to the string")
	}
}

func TestPureJSON_Render(t *testing.T) {
	data := map[string]interface{}{
		"foo": "bar",
		"html": "<h1>Hello</h1>",
	}
	pureJSON := PureJSON{Data: data}
	resp := &protocol.Response{}

	err := pureJSON.Render(resp)
	assert.NoError(t, err)
	assert.Contains(t, string(resp.Body()), `"foo":"bar"`)
	assert.Contains(t, string(resp.Body()), `"html":"<h1>Hello</h1>"`)

	// Test error case
	pureJSON.Data = make(chan int) // Unencodable type
	err = pureJSON.Render(resp)
	assert.Error(t, err)
}

func TestPureJSON_WriteContentType(t *testing.T) {
	pureJSON := PureJSON{}
	resp := &protocol.Response{}
	pureJSON.WriteContentType(resp)
	assert.Equal(t, "application/json; charset=utf-8", string(resp.Header.ContentType()))
}

func TestIndentedJSON_Render(t *testing.T) {
	t.Run("Valid JSON", func(t *testing.T) {
		data := map[string]string{
			"foo": "bar",
			"baz": "qux",
		}
		indentedJSON := IndentedJSON{Data: data}
		resp := &protocol.Response{}

		err := indentedJSON.Render(resp)
		assert.NoError(t, err)
		assert.Contains(t, string(resp.Body()), "{\n    \"baz\": \"qux\",\n    \"foo\": \"bar\"\n}")
	})

	t.Run("Deeply nested JSON", func(t *testing.T) {
		data := map[string]interface{}{
			"level1": map[string]interface{}{
				"level2": map[string]interface{}{
					"level3": map[string]interface{}{
						"key": "value",
					},
				},
			},
		}
		indentedJSON := IndentedJSON{Data: data}
		resp := &protocol.Response{}

		err := indentedJSON.Render(resp)
		assert.NoError(t, err)
		assert.Contains(t, string(resp.Body()), "{\n    \"level1\": {\n        \"level2\": {\n            \"level3\": {\n                \"key\": \"value\"\n            }\n        }\n    }\n}")
	})

	t.Run("Custom jsonMarshalFunc", func(t *testing.T) {
		oldMarshalFunc := jsonMarshalFunc
		defer func() { jsonMarshalFunc = oldMarshalFunc }()

		jsonMarshalFunc = func(v interface{}) ([]byte, error) {
			return []byte(`{"custom":"marshaler"}`), nil
		}

		indentedJSON := IndentedJSON{Data: "anything"}
		resp := &protocol.Response{}

		err := indentedJSON.Render(resp)
		assert.NoError(t, err)
		assert.Equal(t, "{\n    \"custom\": \"marshaler\"\n}", string(resp.Body()))
	})

	t.Run("Large JSON payload", func(t *testing.T) {
		largeData := make(map[string]int)
		for i := 0; i < 1000; i++ {
			largeData[fmt.Sprintf("key%d", i)] = i
		}

		indentedJSON := IndentedJSON{Data: largeData}
		resp := &protocol.Response{}

		err := indentedJSON.Render(resp)
		assert.NoError(t, err)
		assert.Greater(t, len(resp.Body()), 10000)
		assert.Contains(t, string(resp.Body()), "    \"key999\": 999")
	})

	t.Run("Unencodable type", func(t *testing.T) {
		indentedJSON := IndentedJSON{Data: make(chan int)}
		resp := &protocol.Response{}

		err := indentedJSON.Render(resp)
		assert.Error(t, err)
	})

	t.Run("Invalid JSON for indentation", func(t *testing.T) {
		indentedJSON := IndentedJSON{Data: map[string]interface{}{"invalid": func() {}}}
		resp := &protocol.Response{}

		err := indentedJSON.Render(resp)
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "json: unsupported type: func()")
	})

	t.Run("Error during JSON indentation", func(t *testing.T) {
		oldMarshalFunc := jsonMarshalFunc
		defer func() { jsonMarshalFunc = oldMarshalFunc }()

		jsonMarshalFunc = func(v interface{}) ([]byte, error) {
			return []byte(`{"key": "value`), nil // Invalid JSON that will cause indentation error
		}

		indentedJSON := IndentedJSON{Data: "anything"}
		resp := &protocol.Response{}

		err := indentedJSON.Render(resp)
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "unexpected end of JSON input")
	})
}

func TestIndentedJSON_WriteContentType(t *testing.T) {
	indentedJSON := IndentedJSON{}
	resp := &protocol.Response{}
	indentedJSON.WriteContentType(resp)
	assert.Equal(t, "application/json; charset=utf-8", string(resp.Header.ContentType()))
}

func TestJSON_Render(t *testing.T) {
	t.Run("Normal case", func(t *testing.T) {
		data := map[string]interface{}{
			"foo": "bar",
			"num": 42,
		}
		jsonRender := JSONRender{Data: data}
		resp := &protocol.Response{}

		err := jsonRender.Render(resp)
		assert.NoError(t, err)
		assert.Contains(t, string(resp.Body()), `"foo":"bar"`)
		assert.Contains(t, string(resp.Body()), `"num":42`)
	})

	t.Run("Empty data", func(t *testing.T) {
		jsonRender := JSONRender{Data: map[string]string{}}
		resp := &protocol.Response{}

		err := jsonRender.Render(resp)
		assert.NoError(t, err)
		assert.Equal(t, "{}", string(resp.Body()))
	})

	t.Run("Special characters", func(t *testing.T) {
		data := map[string]string{
			"special": "!@#$%^&*()_+{}|:<>?",
		}
		jsonRender := JSONRender{Data: data}
		resp := &protocol.Response{}

		err := jsonRender.Render(resp)
		assert.NoError(t, err)
		assert.Contains(t, string(resp.Body()), `"special":"!@#$%^\u0026*()_+{}|:\u003c\u003e?"`)
	})

	t.Run("Error case", func(t *testing.T) {
		jsonRender := JSONRender{Data: make(chan int)} // Unencodable type
		resp := &protocol.Response{}

		err := jsonRender.Render(resp)
		assert.Error(t, err)
	})

	t.Run("Nil data", func(t *testing.T) {
		jsonRender := JSONRender{Data: nil}
		resp := &protocol.Response{}

		err := jsonRender.Render(resp)
		assert.NoError(t, err)
		assert.Equal(t, "null", string(resp.Body()))
	})

	t.Run("Custom marshaler", func(t *testing.T) {
		oldMarshalFunc := jsonMarshalFunc
		defer func() { jsonMarshalFunc = oldMarshalFunc }()

		jsonMarshalFunc = func(v interface{}) ([]byte, error) {
			return []byte(`{"custom":"marshaler"}`), nil
		}

		jsonRender := JSONRender{Data: "anything"}
		resp := &protocol.Response{}

		err := jsonRender.Render(resp)
		assert.NoError(t, err)
		assert.Equal(t, `{"custom":"marshaler"}`, string(resp.Body()))
	})

	t.Run("Large data", func(t *testing.T) {
		largeData := make(map[string]int)
		for i := 0; i < 1000; i++ {
			largeData[fmt.Sprintf("key%d", i)] = i
		}

		jsonRender := JSONRender{Data: largeData}
		resp := &protocol.Response{}

		err := jsonRender.Render(resp)
		assert.NoError(t, err)
		assert.Greater(t, len(resp.Body()), 5000)
		assert.Contains(t, string(resp.Body()), `"key999":999`)
	})

	t.Run("jsonMarshalFunc error", func(t *testing.T) {
		oldMarshalFunc := jsonMarshalFunc
		defer func() { jsonMarshalFunc = oldMarshalFunc }()

		jsonMarshalFunc = func(v interface{}) ([]byte, error) {
			return nil, fmt.Errorf("mock marshal error")
		}

		jsonRender := JSONRender{Data: "anything"}
		resp := &protocol.Response{}

		err := jsonRender.Render(resp)
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "mock marshal error")
	})

	t.Run("Complex struct with nested fields", func(t *testing.T) {
		type nested struct {
			Field string `json:"field"`
		}
		type complex struct {
			Name   string  `json:"name"`
			Age    int     `json:"age"`
			Nested nested  `json:"nested"`
		}
		data := complex{
			Name: "John",
			Age:  30,
			Nested: nested{
				Field: "value",
			},
		}
		jsonRender := JSONRender{Data: data}
		resp := &protocol.Response{}

		err := jsonRender.Render(resp)
		assert.NoError(t, err)
		assert.Contains(t, string(resp.Body()), `"name":"John"`)
		assert.Contains(t, string(resp.Body()), `"age":30`)
		assert.Contains(t, string(resp.Body()), `"nested":{"field":"value"}`)
	})

	t.Run("Slice of structs", func(t *testing.T) {
		type item struct {
			ID   int    `json:"id"`
			Name string `json:"name"`
		}
		data := []item{
			{ID: 1, Name: "Item 1"},
			{ID: 2, Name: "Item 2"},
		}
		jsonRender := JSONRender{Data: data}
		resp := &protocol.Response{}

		err := jsonRender.Render(resp)
		assert.NoError(t, err)
		assert.Contains(t, string(resp.Body()), `[{"id":1,"name":"Item 1"},{"id":2,"name":"Item 2"}]`)
	})
}
