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
	"strings"
	"testing"

	"github.com/cloudwego/hertz/pkg/common/test/assert"
)

func TestDefaultAndSysLogger(t *testing.T) {
	defaultLog := DefaultLogger()
	systemLog := SystemLogger()

	assert.DeepEqual(t, logger, defaultLog)
	assert.DeepEqual(t, sysLogger, systemLog)
	assert.NotEqual(t, logger, systemLog)
	assert.NotEqual(t, sysLogger, defaultLog)
}

func TestSetLogger(t *testing.T) {
	setLog := &defaultLogger{
		stdlog: log.New(os.Stderr, "", log.LstdFlags|log.Lshortfile|log.Lmicroseconds),
		depth:  6,
	}
	setSysLog := &systemLogger{
		setLog,
		systemLogPrefix,
	}

	assert.NotEqual(t, logger, setLog)
	assert.NotEqual(t, sysLogger, setSysLog)
	SetLogger(setLog)
	assert.DeepEqual(t, logger, setLog)
	assert.DeepEqual(t, sysLogger, setSysLog)
}

func TestSetOutput(t *testing.T) {
	var w byteSliceWriter
	SetOutput(&w)
	Info("test output")
	assert.True(t, len(w.b) > 0 && string(w.b)[len(w.b)-12:] == "test output\n")
}

func TestLevelToString(t *testing.T) {
	assert.DeepEqual(t, "[Trace] ", LevelTrace.toString())
	assert.DeepEqual(t, "[Debug] ", LevelDebug.toString())
	assert.DeepEqual(t, "[Info] ", LevelInfo.toString())
	assert.DeepEqual(t, "[Notice] ", LevelNotice.toString())
	assert.DeepEqual(t, "[Warn] ", LevelWarn.toString())
	assert.DeepEqual(t, "[Error] ", LevelError.toString())
	assert.DeepEqual(t, "[Fatal] ", LevelFatal.toString())

	// Test edge case
	invalidLevel := Level(100)
	assert.DeepEqual(t, "[?100] ", invalidLevel.toString())
}

func TestFullLoggerInterface(t *testing.T) {
	var w byteSliceWriter
	SetOutput(&w)

	// Test all methods of FullLogger interface
	Trace("trace")
	Debug("debug")
	Info("info")
	Notice("notice")
	Warn("warn")
	Error("error")

	Tracef("trace %s", "format")
	Debugf("debug %s", "format")
	Infof("info %s", "format")
	Noticef("notice %s", "format")
	Warnf("warn %s", "format")
	Errorf("error %s", "format")

	ctx := context.Background()
	CtxTracef(ctx, "ctx trace %s", "format")
	CtxDebugf(ctx, "ctx debug %s", "format")
	CtxInfof(ctx, "ctx info %s", "format")
	CtxNoticef(ctx, "ctx notice %s", "format")
	CtxWarnf(ctx, "ctx warn %s", "format")
	CtxErrorf(ctx, "ctx error %s", "format")

	SetLevel(LevelInfo)

	output := string(w.b)
	assert.True(t, strings.Contains(output, "[Trace] trace"))
	assert.True(t, strings.Contains(output, "[Debug] debug"))
	assert.True(t, strings.Contains(output, "[Info] info"))
	assert.True(t, strings.Contains(output, "[Notice] notice"))
	assert.True(t, strings.Contains(output, "[Warn] warn"))
	assert.True(t, strings.Contains(output, "[Error] error"))
	assert.True(t, strings.Contains(output, "[Trace] trace format"))
	assert.True(t, strings.Contains(output, "[Debug] debug format"))
	assert.True(t, strings.Contains(output, "[Info] info format"))
	assert.True(t, strings.Contains(output, "[Notice] notice format"))
	assert.True(t, strings.Contains(output, "[Warn] warn format"))
	assert.True(t, strings.Contains(output, "[Error] error format"))
	assert.True(t, strings.Contains(output, "[Trace] ctx trace format"))
	assert.True(t, strings.Contains(output, "[Debug] ctx debug format"))
	assert.True(t, strings.Contains(output, "[Info] ctx info format"))
	assert.True(t, strings.Contains(output, "[Notice] ctx notice format"))
	assert.True(t, strings.Contains(output, "[Warn] ctx warn format"))
	assert.True(t, strings.Contains(output, "[Error] ctx error format"))
}
