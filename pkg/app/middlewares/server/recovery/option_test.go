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
	"io"
	"strings"
	"testing"

	"github.com/cloudwego/hertz/pkg/app"
	"github.com/cloudwego/hertz/pkg/common/hlog"
	"github.com/cloudwego/hertz/pkg/common/test/assert"
	"github.com/cloudwego/hertz/pkg/common/utils"
	"github.com/cloudwego/hertz/pkg/protocol/consts"
)

func TestDefaultOption(t *testing.T) {
	opts := newOptions()
	assert.DeepEqual(t, fmt.Sprintf("%p", defaultRecoveryHandler), fmt.Sprintf("%p", opts.recoveryHandler))
}

func newRecoveryHandler(c context.Context, ctx *app.RequestContext, err interface{}, stack []byte) {
	hlog.SystemLogger().CtxErrorf(c, "[New Recovery] panic recovered:\n%s\n%s\n",
		err, stack)
	ctx.JSON(consts.StatusNotImplemented, utils.H{"msg": err.(string)})
}

func TestOption(t *testing.T) {
	opts := newOptions(WithRecoveryHandler(newRecoveryHandler))
	assert.DeepEqual(t, fmt.Sprintf("%p", newRecoveryHandler), fmt.Sprintf("%p", opts.recoveryHandler))
}

func TestDefaultRecoveryHandler(t *testing.T) {
	ctx := app.NewContext(16)
	err := "test error"
	stack := []byte("test stack")

	// Create a mock logger to capture the log output
	mockLogger := &mockSystemLogger{}
	hlog.SetSystemLogger(mockLogger)

	defaultRecoveryHandler(context.Background(), ctx, err, stack)

	// Check if the error is logged correctly
	assert.True(t, strings.Contains(mockLogger.lastErrorMessage, "[Recovery] err=test error"))
	assert.True(t, strings.Contains(mockLogger.lastErrorMessage, "stack=test stack"))

	// Check if the status code is set correctly
	assert.DeepEqual(t, consts.StatusInternalServerError, ctx.Response.StatusCode())
}

func TestNewOptionsWithMultipleOptions(t *testing.T) {
	customHandler := func(c context.Context, ctx *app.RequestContext, err interface{}, stack []byte) {
		ctx.JSON(consts.StatusServiceUnavailable, utils.H{"error": err.(string)})
	}

	opts := newOptions(
		WithRecoveryHandler(customHandler),
		// Add more options here if needed
	)

	// Verify that the custom handler is set
	assert.DeepEqual(t, fmt.Sprintf("%p", customHandler), fmt.Sprintf("%p", opts.recoveryHandler))

	// Test the custom handler
	ctx := app.NewContext(16)
	err := "custom error"
	stack := []byte("custom stack")

	opts.recoveryHandler(context.Background(), ctx, err, stack)

	// Check if the status code is set correctly by the custom handler
	assert.DeepEqual(t, consts.StatusServiceUnavailable, ctx.Response.StatusCode())

	// Check if the response body contains the expected error message
	assert.DeepEqual(t, `{"error":"custom error"}`, string(ctx.Response.Body()))
}

// mockSystemLogger is a mock implementation of hlog.FullLogger
type mockSystemLogger struct {
	lastErrorMessage string
}

func (m *mockSystemLogger) CtxErrorf(ctx context.Context, format string, v ...interface{}) {
	m.lastErrorMessage = fmt.Sprintf(format, v...)
}

// Implement other methods of hlog.FullLogger interface with empty bodies
func (m *mockSystemLogger) Trace(v ...interface{})                                        {}
func (m *mockSystemLogger) Debug(v ...interface{})                                        {}
func (m *mockSystemLogger) Info(v ...interface{})                                         {}
func (m *mockSystemLogger) Notice(v ...interface{})                                       {}
func (m *mockSystemLogger) Warn(v ...interface{})                                         {}
func (m *mockSystemLogger) Error(v ...interface{})                                        {}
func (m *mockSystemLogger) Fatal(v ...interface{})                                        {}
func (m *mockSystemLogger) Tracef(format string, v ...interface{})                        {}
func (m *mockSystemLogger) Debugf(format string, v ...interface{})                        {}
func (m *mockSystemLogger) Infof(format string, v ...interface{})                         {}
func (m *mockSystemLogger) Noticef(format string, v ...interface{})                       {}
func (m *mockSystemLogger) Warnf(format string, v ...interface{})                         {}
func (m *mockSystemLogger) Errorf(format string, v ...interface{})                        {}
func (m *mockSystemLogger) Fatalf(format string, v ...interface{})                        {}
func (m *mockSystemLogger) CtxTracef(ctx context.Context, format string, v ...interface{})   {}
func (m *mockSystemLogger) CtxDebugf(ctx context.Context, format string, v ...interface{})   {}
func (m *mockSystemLogger) CtxInfof(ctx context.Context, format string, v ...interface{})    {}
func (m *mockSystemLogger) CtxNoticef(ctx context.Context, format string, v ...interface{})  {}
func (m *mockSystemLogger) CtxWarnf(ctx context.Context, format string, v ...interface{})    {}
func (m *mockSystemLogger) CtxFatalf(ctx context.Context, format string, v ...interface{})   {}
func (m *mockSystemLogger) SetLevel(level hlog.Level)                                        {}
func (m *mockSystemLogger) SetOutput(w io.Writer)                                            {}
