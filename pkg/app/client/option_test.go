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

package client

import (
	"crypto/tls"
	"testing"
	"time"

	"github.com/cloudwego/hertz/pkg/app/client/retry"
	"github.com/cloudwego/hertz/pkg/common/config"
	"github.com/cloudwego/hertz/pkg/common/test/assert"
	"github.com/cloudwego/hertz/pkg/network"
)

// mockHostClientState implements the HostClientState interface for testing
type mockHostClientState struct{}

func (m *mockHostClientState) ConnPoolState() config.ConnPoolState {
	return config.ConnPoolState{
		PoolConnNum:  1,
		TotalConnNum: 2,
		WaitConnNum:  0,
		Addr:         "test-addr",
	}
}

func TestClientOptions(t *testing.T) {
	tlsConfig := &tls.Config{InsecureSkipVerify: true}
	customDialer := &mockDialer{}
	hostClientConfigHook := func(hc interface{}) error { return nil }
	connStateObserveFunc := func(hcs config.HostClientState) {}

	opt := config.NewClientOptions([]config.ClientOption{
		WithDialTimeout(100 * time.Millisecond),
		WithMaxConnsPerHost(128),
		WithMaxIdleConnDuration(5 * time.Second),
		WithMaxConnDuration(10 * time.Second),
		WithMaxConnWaitTimeout(5 * time.Second),
		WithKeepAlive(false),
		WithClientReadTimeout(1 * time.Second),
		WithResponseBodyStream(true),
		WithRetryConfig(
			retry.WithMaxAttemptTimes(2),
			retry.WithInitDelay(100*time.Millisecond),
			retry.WithMaxDelay(5*time.Second),
			retry.WithMaxJitter(1*time.Second),
			retry.WithDelayPolicy(retry.CombineDelay(retry.DefaultDelayPolicy, retry.FixedDelayPolicy, retry.BackOffDelayPolicy)),
		),
		WithWriteTimeout(time.Second),
		WithConnStateObserve(connStateObserveFunc, time.Second),
		WithTLSConfig(tlsConfig),
		WithDialer(customDialer),
		WithHostClientConfigHook(hostClientConfigHook),
		WithDisableHeaderNamesNormalizing(true),
		WithName("TestClient"),
		WithNoDefaultUserAgentHeader(true),
		WithDisablePathNormalizing(true),
	})

	assert.DeepEqual(t, 100*time.Millisecond, opt.DialTimeout)
	assert.DeepEqual(t, 128, opt.MaxConnsPerHost)
	assert.DeepEqual(t, 5*time.Second, opt.MaxIdleConnDuration)
	assert.DeepEqual(t, 10*time.Second, opt.MaxConnDuration)
	assert.DeepEqual(t, 5*time.Second, opt.MaxConnWaitTimeout)
	assert.DeepEqual(t, false, opt.KeepAlive)
	assert.DeepEqual(t, 1*time.Second, opt.ReadTimeout)
	assert.DeepEqual(t, 1*time.Second, opt.WriteTimeout)
	assert.DeepEqual(t, true, opt.ResponseBodyStream)
	assert.DeepEqual(t, uint(2), opt.RetryConfig.MaxAttemptTimes)
	assert.DeepEqual(t, 100*time.Millisecond, opt.RetryConfig.Delay)
	assert.DeepEqual(t, 5*time.Second, opt.RetryConfig.MaxDelay)
	assert.DeepEqual(t, 1*time.Second, opt.RetryConfig.MaxJitter)
	assert.DeepEqual(t, 1*time.Second, opt.ObservationInterval)
	assert.DeepEqual(t, tlsConfig, opt.TLSConfig)
	assert.NotNil(t, opt.Dialer)
	assert.True(t, func() bool { _, ok := opt.Dialer.(*mockDialer); return ok }())
	assert.NotNil(t, opt.HostClientConfigHook)
	assert.DeepEqual(t, true, opt.DisableHeaderNamesNormalizing)
	assert.DeepEqual(t, "TestClient", opt.Name)
	assert.DeepEqual(t, true, opt.NoDefaultUserAgentHeader)
	assert.DeepEqual(t, true, opt.DisablePathNormalizing)
	assert.NotNil(t, opt.HostClientStateObserve)

	// Test RetryConfig DelayPolicy
	for i := 0; i < 100; i++ {
		assert.DeepEqual(t, opt.RetryConfig.DelayPolicy(uint(i), nil, opt.RetryConfig), retry.CombineDelay(retry.DefaultDelayPolicy, retry.FixedDelayPolicy, retry.BackOffDelayPolicy)(uint(i), nil, opt.RetryConfig))
	}

	// Test WithDialFunc
	dialFuncCalled := false
	opt = config.NewClientOptions([]config.ClientOption{
		WithDialFunc(func(addr string) (network.Conn, error) {
			dialFuncCalled = true
			return nil, nil
		}),
	})
	assert.NotNil(t, opt.Dialer)
	_, err := opt.Dialer.DialConnection("tcp", "localhost:8080", time.Second, nil)
	assert.Nil(t, err)
	assert.True(t, dialFuncCalled)

	// Test WithConnStateObserve
	observeCalled := false
	opt = config.NewClientOptions([]config.ClientOption{
		WithConnStateObserve(func(hcs config.HostClientState) {
			observeCalled = true
		}, 2*time.Second),
	})
	assert.NotNil(t, opt.HostClientStateObserve)
	assert.DeepEqual(t, 2*time.Second, opt.ObservationInterval)

	// Create a mock implementation of HostClientState
	mockState := &mockHostClientState{}

	opt.HostClientStateObserve(mockState)
	assert.True(t, observeCalled)
}
