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

package config

import (
	"testing"
	"time"

	"github.com/cloudwego/hertz/pkg/app/client/retry"
	"github.com/cloudwego/hertz/pkg/common/test/assert"
	"github.com/cloudwego/hertz/pkg/protocol/consts"
)

// TestDefaultClientOptions test client options with default values
func TestDefaultClientOptions(t *testing.T) {
	options := NewClientOptions([]ClientOption{})

	assert.DeepEqual(t, consts.DefaultDialTimeout, options.DialTimeout)
	assert.DeepEqual(t, consts.DefaultMaxConnsPerHost, options.MaxConnsPerHost)
	assert.DeepEqual(t, consts.DefaultMaxIdleConnDuration, options.MaxIdleConnDuration)
	assert.True(t, options.KeepAlive)
	assert.DeepEqual(t, time.Second*5, options.ObservationInterval)
	assert.Nil(t, options.TLSConfig)
	assert.False(t, options.ResponseBodyStream)
	assert.DeepEqual(t, "", options.Name)
	assert.False(t, options.NoDefaultUserAgentHeader)
	assert.Nil(t, options.Dialer)
	assert.False(t, options.DialDualStack)
	assert.DeepEqual(t, time.Duration(0), options.WriteTimeout)
	assert.DeepEqual(t, 0, options.MaxResponseBodySize)
	assert.False(t, options.DisableHeaderNamesNormalizing)
	assert.False(t, options.DisablePathNormalizing)
	assert.Nil(t, options.RetryConfig)
	assert.Nil(t, options.HostClientStateObserve)
	assert.Nil(t, options.HostClientConfigHook)
}

// TestCustomClientOptions test client options with custom values
func TestCustomClientOptions(t *testing.T) {
	options := NewClientOptions([]ClientOption{})

	options.Apply([]ClientOption{
		{
			F: func(o *ClientOptions) {
				o.DialTimeout = 2 * time.Second
			},
		},
	})
	assert.DeepEqual(t, 2*time.Second, options.DialTimeout)
}

// TestNewClientOptionsWithCustomValues tests NewClientOptions with custom values
func TestNewClientOptionsWithCustomValues(t *testing.T) {
	customOptions := []ClientOption{
		{
			F: func(o *ClientOptions) {
				o.ReadTimeout = 5 * time.Second
				o.WriteTimeout = 10 * time.Second
				o.MaxResponseBodySize = 1024 * 1024 // 1MB
				o.MaxConnsPerHost = 100
				o.DisableHeaderNamesNormalizing = true
			},
		},
	}

	options := NewClientOptions(customOptions)

	assert.DeepEqual(t, 5*time.Second, options.ReadTimeout)
	assert.DeepEqual(t, 10*time.Second, options.WriteTimeout)
	assert.DeepEqual(t, 1024*1024, options.MaxResponseBodySize)
	assert.DeepEqual(t, 100, options.MaxConnsPerHost)
	assert.True(t, options.DisableHeaderNamesNormalizing)
}

// TestApplyMultipleClientOptions tests Apply method with multiple ClientOption inputs
func TestApplyMultipleClientOptions(t *testing.T) {
	options := NewClientOptions([]ClientOption{})

	options.Apply([]ClientOption{
		{
			F: func(o *ClientOptions) {
				o.MaxConnWaitTimeout = 3 * time.Second
			},
		},
		{
			F: func(o *ClientOptions) {
				o.DisableHeaderNamesNormalizing = true
			},
		},
		{
			F: func(o *ClientOptions) {
				o.DisablePathNormalizing = true
			},
		},
	})

	assert.DeepEqual(t, 3*time.Second, options.MaxConnWaitTimeout)
	assert.DeepEqual(t, true, options.DisableHeaderNamesNormalizing)
	assert.DeepEqual(t, true, options.DisablePathNormalizing)
}

// TestClientOptionsRetryConfig tests the RetryConfig field
func TestClientOptionsRetryConfig(t *testing.T) {
	retryConfig := &retry.Config{
		MaxAttemptTimes: 3,
		Delay:           100 * time.Millisecond,
	}
	options := NewClientOptions([]ClientOption{
		{
			F: func(o *ClientOptions) {
				o.RetryConfig = retryConfig
			},
		},
	})

	assert.DeepEqual(t, retryConfig, options.RetryConfig)
}

// TestClientOptionsHostClientStateObserve tests the HostClientStateObserve field
func TestClientOptionsHostClientStateObserve(t *testing.T) {
	var observedState ConnPoolState
	stateObserveFunc := func(state HostClientState) {
		observedState = state.ConnPoolState()
	}

	options := NewClientOptions([]ClientOption{
		{
			F: func(o *ClientOptions) {
				o.HostClientStateObserve = stateObserveFunc
			},
		},
	})

	assert.NotNil(t, options.HostClientStateObserve)

	// Simulate calling the HostClientStateObserve function
	mockState := &mockHostClientState{
		connPoolState: ConnPoolState{PoolConnNum: 5, TotalConnNum: 10},
	}
	options.HostClientStateObserve(mockState)

	assert.DeepEqual(t, ConnPoolState{PoolConnNum: 5, TotalConnNum: 10}, observedState)
}

// TestClientOptionsHostClientConfigHook tests the HostClientConfigHook field
func TestClientOptionsHostClientConfigHook(t *testing.T) {
	hookCalled := false
	configHook := func(hc interface{}) error {
		hookCalled = true
		return nil
	}

	options := NewClientOptions([]ClientOption{
		{
			F: func(o *ClientOptions) {
				o.HostClientConfigHook = configHook
			},
		},
	})

	assert.NotNil(t, options.HostClientConfigHook)

	// Simulate calling the HostClientConfigHook function
	err := options.HostClientConfigHook(nil)
	assert.Nil(t, err)
	assert.True(t, hookCalled)
}

// mockHostClientState is a mock implementation of HostClientState for testing
type mockHostClientState struct {
	connPoolState ConnPoolState
}

func (m *mockHostClientState) ConnPoolState() ConnPoolState {
	return m.connPoolState
}
