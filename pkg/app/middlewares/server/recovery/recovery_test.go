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

package recovery

import (
	"context"
	"fmt"
	"strings"
	"testing"

	"github.com/cloudwego/hertz/pkg/app"
	"github.com/cloudwego/hertz/pkg/common/test/assert"
	"github.com/cloudwego/hertz/pkg/protocol/consts"
)

func TestRecovery(t *testing.T) {
	ctx := app.NewContext(0)
	var hc app.HandlersChain
	hc = append(hc, func(c context.Context, ctx *app.RequestContext) {
		fmt.Println("this is test")
		panic("test")
	})
	ctx.SetHandlers(hc)

	Recovery()(context.Background(), ctx)

	if ctx.Response.StatusCode() != 500 {
		t.Fatalf("unexpected %v. Expecting %v", ctx.Response.StatusCode(), 500)
	}
}

func TestWithRecoveryHandler(t *testing.T) {
	ctx := app.NewContext(0)
	var hc app.HandlersChain
	hc = append(hc, func(c context.Context, ctx *app.RequestContext) {
		fmt.Println("this is test")
		panic("test")
	})
	ctx.SetHandlers(hc)

	Recovery(WithRecoveryHandler(newRecoveryHandler))(context.Background(), ctx)

	if ctx.Response.StatusCode() != consts.StatusNotImplemented {
		t.Fatalf("unexpected %v. Expecting %v", ctx.Response.StatusCode(), 501)
	}
	assert.DeepEqual(t, "{\"msg\":\"test\"}", string(ctx.Response.Body()))
}

func TestRecoveryWithDifferentPanicTypes(t *testing.T) {
	testCases := []struct {
		name        string
		panicValue  interface{}
		expectedMsg string
	}{
		{"string panic", "test panic", "test panic"},
		{"error panic", fmt.Errorf("test error"), "test error"},
		{"integer panic", 42, "42"},
		{"struct panic", struct{ msg string }{"custom panic"}, "{custom panic}"},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			ctx := app.NewContext(0)
			var hc app.HandlersChain
			hc = append(hc, func(c context.Context, ctx *app.RequestContext) {
				panic(tc.panicValue)
			})
			ctx.SetHandlers(hc)

			var capturedPanic interface{}
			customRecoveryHandler := func(c context.Context, ctx *app.RequestContext, err interface{}, stack []byte) {
				capturedPanic = err
				ctx.AbortWithStatus(500)
				ctx.String(500, fmt.Sprintf("%v", err))
			}

			Recovery(WithRecoveryHandler(customRecoveryHandler))(context.Background(), ctx)

			assert.DeepEqual(t, 500, ctx.Response.StatusCode())
			assert.DeepEqual(t, tc.panicValue, capturedPanic)
			responseBody := string(ctx.Response.Body())
			assert.True(t, strings.Contains(responseBody, tc.expectedMsg))
		})
	}
}

func TestRecoveryWithNestedHandlerChain(t *testing.T) {
	ctx := app.NewContext(0)
	var hc app.HandlersChain
	hc = append(hc, func(c context.Context, ctx *app.RequestContext) {
		ctx.Next(c)
	})
	hc = append(hc, func(c context.Context, ctx *app.RequestContext) {
		panic("nested panic")
	})
	ctx.SetHandlers(hc)

	var capturedPanic interface{}
	customRecoveryHandler := func(c context.Context, ctx *app.RequestContext, err interface{}, stack []byte) {
		capturedPanic = err
		ctx.AbortWithStatus(500)
		ctx.String(500, fmt.Sprintf("%v", err))
	}

	Recovery(WithRecoveryHandler(customRecoveryHandler))(context.Background(), ctx)

	assert.DeepEqual(t, 500, ctx.Response.StatusCode())
	assert.DeepEqual(t, "nested panic", capturedPanic)
	responseBody := string(ctx.Response.Body())
	assert.True(t, strings.Contains(responseBody, "nested panic"))
}

func TestRecoveryStackTraceGeneration(t *testing.T) {
	ctx := app.NewContext(0)
	var hc app.HandlersChain
	hc = append(hc, func(c context.Context, ctx *app.RequestContext) {
		panic("stack trace test")
	})
	ctx.SetHandlers(hc)

	var stackTrace []byte
	customRecoveryHandler := func(c context.Context, ctx *app.RequestContext, err interface{}, stack []byte) {
		stackTrace = stack
		ctx.AbortWithStatus(500)
	}

	Recovery(WithRecoveryHandler(customRecoveryHandler))(context.Background(), ctx)

	assert.DeepEqual(t, 500, ctx.Response.StatusCode())
	assert.NotNil(t, stackTrace)
	assert.True(t, strings.Contains(string(stackTrace), "stack trace test"))
	assert.True(t, strings.Contains(string(stackTrace), "recovery_test.go"))
}
