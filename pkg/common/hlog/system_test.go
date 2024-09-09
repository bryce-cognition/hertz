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

package hlog

import (
	"context"
	"log"
	"os"
	"testing"

	"github.com/stretchr/testify/assert"
)

func initTestSysLogger() {
	sysLogger = &systemLogger{
		&defaultLogger{
			stdlog: log.New(os.Stderr, "", 0),
			depth:  4,
		},
		systemLogPrefix,
	}
}

func TestSysLogger(t *testing.T) {
	initTestSysLogger()
	var w byteSliceWriter
	SetOutput(&w)

	sysLogger.Trace("trace work")
	sysLogger.Debug("received work order")
	sysLogger.Info("starting work")
	sysLogger.Notice("something happens in work")
	sysLogger.Warn("work may fail")
	sysLogger.Error("work failed")

	assert.Equal(t, "[Trace] HERTZ: trace work\n"+
		"[Debug] HERTZ: received work order\n"+
		"[Info] HERTZ: starting work\n"+
		"[Notice] HERTZ: something happens in work\n"+
		"[Warn] HERTZ: work may fail\n"+
		"[Error] HERTZ: work failed\n", string(w.b))
}

func TestSysFormatLogger(t *testing.T) {
	initTestSysLogger()
	var w byteSliceWriter
	SetOutput(&w)

	work := "work"
	sysLogger.Tracef("trace %s", work)
	sysLogger.Debugf("received %s order", work)
	sysLogger.Infof("starting %s", work)
	sysLogger.Noticef("something happens in %s", work)
	sysLogger.Warnf("%s may fail", work)
	sysLogger.Errorf("%s failed", work)

	assert.Equal(t, "[Trace] HERTZ: trace work\n"+
		"[Debug] HERTZ: received work order\n"+
		"[Info] HERTZ: starting work\n"+
		"[Notice] HERTZ: something happens in work\n"+
		"[Warn] HERTZ: work may fail\n"+
		"[Error] HERTZ: work failed\n", string(w.b))
}

func TestSysCtxLogger(t *testing.T) {
	initTestSysLogger()
	var w byteSliceWriter
	SetOutput(&w)

	ctx := context.Background()
	work := "work"
	sysLogger.CtxTracef(ctx, "trace %s", work)
	sysLogger.CtxDebugf(ctx, "received %s order", work)
	sysLogger.CtxInfof(ctx, "starting %s", work)
	sysLogger.CtxNoticef(ctx, "something happens in %s", work)
	sysLogger.CtxWarnf(ctx, "%s may fail", work)
	sysLogger.CtxErrorf(ctx, "%s failed", work)

	assert.Equal(t, "[Trace] HERTZ: trace work\n"+
		"[Debug] HERTZ: received work order\n"+
		"[Info] HERTZ: starting work\n"+
		"[Notice] HERTZ: something happens in work\n"+
		"[Warn] HERTZ: work may fail\n"+
		"[Error] HERTZ: work failed\n", string(w.b))
}

func TestSetSilentMode(t *testing.T) {
	initTestSysLogger()
	var w byteSliceWriter
	SetOutput(&w)

	// Test with silent mode off
	SetSilentMode(false)
	sysLogger.Errorf(EngineErrorFormat, "test error")
	assert.Contains(t, string(w.b), "[Error] HERTZ: Error=test error, remoteAddr=%!s(MISSING)")

	// Clear the buffer
	w = byteSliceWriter{}

	// Test with silent mode on
	SetSilentMode(true)
	sysLogger.Errorf(EngineErrorFormat, "test error")
	assert.Equal(t, "", string(w.b))

	// Test non-engine error with silent mode on
	sysLogger.Errorf("Non-engine error: %s", "test")
	assert.Contains(t, string(w.b), "[Error] HERTZ: Non-engine error: test")
}

func TestSetOutput(t *testing.T) {
	initTestSysLogger()
	var w1, w2 byteSliceWriter

	// Test with first output
	SetOutput(&w1)
	sysLogger.Info("test output 1")
	assert.Contains(t, string(w1.b), "HERTZ: test output 1")
	assert.Equal(t, "", string(w2.b))

	// Test with second output
	SetOutput(&w2)
	sysLogger.Info("test output 2")
	assert.NotContains(t, string(w1.b), "test output 2")
	assert.Contains(t, string(w2.b), "HERTZ: test output 2")
}

func TestSetLevelSystem(t *testing.T) {
	initTestSysLogger()
	var w byteSliceWriter
	SetOutput(&w)

	// Test Info level
	SetLevel(LevelInfo)
	sysLogger.Debug("debug message")
	sysLogger.Info("info message")
	assert.NotContains(t, string(w.b), "debug message")
	assert.Contains(t, string(w.b), "info message")

	// Clear the buffer
	w.b = w.b[:0]

	// Test Debug level
	SetLevel(LevelDebug)
	sysLogger.Debug("debug message")
	sysLogger.Info("info message")
	assert.Contains(t, string(w.b), "debug message")
	assert.Contains(t, string(w.b), "info message")
}

func TestErrorLoggingWithSilentMode(t *testing.T) {
	initTestSysLogger()
	var w byteSliceWriter
	SetOutput(&w)

	// Test error logging with silent mode off
	SetSilentMode(false)
	sysLogger.Errorf(EngineErrorFormat, "engine error")
	sysLogger.Errorf("non-engine error: %s", "test")
	assert.Contains(t, string(w.b), "[Error] HERTZ: Error=engine error, remoteAddr=%!s(MISSING)")
	assert.Contains(t, string(w.b), "[Error] HERTZ: non-engine error: test")

	// Clear the buffer
	w = byteSliceWriter{}

	// Test error logging with silent mode on
	SetSilentMode(true)
	sysLogger.Errorf(EngineErrorFormat, "engine error")
	sysLogger.Errorf("non-engine error: %s", "test")
	assert.NotContains(t, string(w.b), "Error=engine error")
	assert.Contains(t, string(w.b), "[Error] HERTZ: non-engine error: test")
}
