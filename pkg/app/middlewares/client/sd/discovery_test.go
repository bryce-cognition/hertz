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

package sd

import (
	"context"
	"errors"
	"sync"
	"testing"

	"github.com/cloudwego/hertz/pkg/app/client/discovery"
	"github.com/cloudwego/hertz/pkg/common/config"
	"github.com/cloudwego/hertz/pkg/common/test/assert"
	"github.com/cloudwego/hertz/pkg/protocol"
)



// MockLoadbalancer implements the Loadbalancer interface for testing
type MockLoadbalancer struct {
	mu         sync.Mutex
	pickCalled int
}

func (m *MockLoadbalancer) Pick(result discovery.Result) discovery.Instance {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.pickCalled++
	if len(result.Instances) > 0 {
		return result.Instances[0]
	}
	return nil
}

func (m *MockLoadbalancer) Rebalance(discovery.Result) {}

func (m *MockLoadbalancer) Delete(string) {}

func (m *MockLoadbalancer) Name() string {
	return "MockLoadbalancer"
}

func TestDiscovery(t *testing.T) {
	inss := []discovery.Instance{
		discovery.NewInstance("tcp", "127.0.0.1:8888", 10, nil),
		discovery.NewInstance("tcp", "127.0.0.1:8889", 10, nil),
	}
	r := &discovery.SynthesizedResolver{
		TargetFunc: func(ctx context.Context, target *discovery.TargetInfo) string {
			return target.Host
		},
		ResolveFunc: func(ctx context.Context, key string) (discovery.Result, error) {
			return discovery.Result{CacheKey: "svc1", Instances: inss}, nil
		},
		NameFunc: func() string { return t.Name() },
	}

	t.Run("Normal case", func(t *testing.T) {
		mw := Discovery(r)
		checkMdw := func(ctx context.Context, req *protocol.Request, resp *protocol.Response) (err error) {
			t.Log(string(req.Host()))
			assert.Assert(t, string(req.Host()) == "127.0.0.1:8888" || string(req.Host()) == "127.0.0.1:8889")
			return nil
		}
		for i := 0; i < 10; i++ {
			req := &protocol.Request{}
			resp := &protocol.Response{}
			req.Options().Apply([]config.RequestOption{config.WithSD(true)})
			req.SetRequestURI("http://service_name")
			err := mw(checkMdw)(context.Background(), req, resp)
			assert.Nil(t, err)
		}
	})

	t.Run("Error handling when GetInstance fails", func(t *testing.T) {
		errResolver := &discovery.SynthesizedResolver{
			ResolveFunc: func(ctx context.Context, key string) (discovery.Result, error) {
				return discovery.Result{}, errors.New("resolve error")
			},
		}
		mw := Discovery(errResolver)
		req := &protocol.Request{}
		resp := &protocol.Response{}
		req.Options().Apply([]config.RequestOption{config.WithSD(true)})
		req.SetRequestURI("http://service_name")
		err := mw(func(ctx context.Context, req *protocol.Request, resp *protocol.Response) error {
			return nil
		})(context.Background(), req, resp)
		assert.NotNil(t, err)
		assert.DeepEqual(t, "resolve error", err.Error())
	})

	t.Run("Different ServiceDiscoveryOptions", func(t *testing.T) {
		customBalancer := &MockLoadbalancer{}
		customOption := ServiceDiscoveryOption{F: func(opts *ServiceDiscoveryOptions) {
			opts.Balancer = customBalancer
		}}
		mw := Discovery(r, customOption)
		req := &protocol.Request{}
		resp := &protocol.Response{}
		req.Options().Apply([]config.RequestOption{config.WithSD(true)})
		req.SetRequestURI("http://service_name")
		err := mw(func(ctx context.Context, req *protocol.Request, resp *protocol.Response) error {
			return nil
		})(context.Background(), req, resp)
		assert.Nil(t, err)
		assert.DeepEqual(t, 1, customBalancer.pickCalled)
	})

	t.Run("Behavior when req.Options().IsSD() is false", func(t *testing.T) {
		mw := Discovery(r)
		req := &protocol.Request{}
		resp := &protocol.Response{}
		req.SetRequestURI("http://service_name")
		req.SetHost("original_host")
		err := mw(func(ctx context.Context, req *protocol.Request, resp *protocol.Response) error {
			assert.DeepEqual(t, "original_host", string(req.Host()))
			return nil
		})(context.Background(), req, resp)
		assert.Nil(t, err)
	})
}
