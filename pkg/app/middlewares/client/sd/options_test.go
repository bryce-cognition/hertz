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
	"testing"
	"time"

	"github.com/cloudwego/hertz/pkg/app/client/loadbalance"
	"github.com/cloudwego/hertz/pkg/common/test/assert"
)

func TestWithCustomizedAddrs(t *testing.T) {
	t.Run("Valid addresses", func(t *testing.T) {
		var options []ServiceDiscoveryOption
		options = append(options, WithCustomizedAddrs("127.0.0.1:8080", "/tmp/unix_ss"))
		opts := &ServiceDiscoveryOptions{}
		opts.Apply(options)
		assert.Assert(t, opts.Resolver.Name() == "127.0.0.1:8080,/tmp/unix_ss")
		res, err := opts.Resolver.Resolve(context.Background(), "")
		assert.Assert(t, err == nil)
		assert.Assert(t, res.Instances[0].Address().String() == "127.0.0.1:8080")
		assert.Assert(t, res.Instances[1].Address().String() == "/tmp/unix_ss")
	})

	t.Run("Valid TCP address with scheme", func(t *testing.T) {
		var options []ServiceDiscoveryOption
		options = append(options, WithCustomizedAddrs("tcp://127.0.0.1:8080"))
		opts := &ServiceDiscoveryOptions{}
		opts.Apply(options)
		assert.Assert(t, opts.Resolver.Name() == "tcp://127.0.0.1:8080")
		res, err := opts.Resolver.Resolve(context.Background(), "")
		assert.Assert(t, err == nil)
		assert.Assert(t, res.Instances[0].Address().String() == "127.0.0.1:8080")
	})

	t.Run("Valid Unix address with scheme", func(t *testing.T) {
		var options []ServiceDiscoveryOption
		options = append(options, WithCustomizedAddrs("unix:///tmp/unix_ss"))
		opts := &ServiceDiscoveryOptions{}
		opts.Apply(options)
		assert.Assert(t, opts.Resolver.Name() == "unix:///tmp/unix_ss")
		res, err := opts.Resolver.Resolve(context.Background(), "")
		assert.Assert(t, err == nil)
		assert.Assert(t, res.Instances[0].Address().String() == "/tmp/unix_ss")
	})

	t.Run("Invalid scheme", func(t *testing.T) {
		assert.Panic(t, func() {
			option := WithCustomizedAddrs("invalid://127.0.0.1:8080")
			opts := &ServiceDiscoveryOptions{}
			option.F(opts)
		})
	})

	t.Run("Invalid TCP address", func(t *testing.T) {
		assert.Panic(t, func() {
			option := WithCustomizedAddrs("127.0.0.1:invalid_port")
			opts := &ServiceDiscoveryOptions{}
			option.F(opts)
		})
	})

	t.Run("Invalid Unix address (not absolute path)", func(t *testing.T) {
		assert.Panic(t, func() {
			option := WithCustomizedAddrs("relative/path/to/socket")
			opts := &ServiceDiscoveryOptions{}
			option.F(opts)
		})
	})

	t.Run("Mixed valid and invalid addresses", func(t *testing.T) {
		assert.Panic(t, func() {
			option := WithCustomizedAddrs("127.0.0.1:8080", "invalid://address", "/tmp/unix_ss")
			opts := &ServiceDiscoveryOptions{}
			option.F(opts)
		})
	})

	t.Run("No addresses", func(t *testing.T) {
		assert.Panic(t, func() {
			option := WithCustomizedAddrs()
			opts := &ServiceDiscoveryOptions{}
			option.F(opts)
		})
	})
}

func TestWithLoadBalanceOptions(t *testing.T) {
	t.Run("Weighted Balancer with Default Options", func(t *testing.T) {
		balance := loadbalance.NewWeightedBalancer()
		var options []ServiceDiscoveryOption
		options = append(options, WithLoadBalanceOptions(balance, loadbalance.DefaultLbOpts))
		opts := &ServiceDiscoveryOptions{}
		opts.Apply(options)
		assert.Assert(t, opts.Balancer.Name() == "weight_random")
		assert.DeepEqual(t, opts.LbOpts, loadbalance.DefaultLbOpts)
	})

	t.Run("Weighted Balancer with Custom Options", func(t *testing.T) {
		balance := loadbalance.NewWeightedBalancer()
		customOpts := loadbalance.Options{
			RefreshInterval: 60 * time.Second,
			ExpireInterval:  300 * time.Second,
		}
		var options []ServiceDiscoveryOption
		options = append(options, WithLoadBalanceOptions(balance, customOpts))
		opts := &ServiceDiscoveryOptions{}
		opts.Apply(options)
		assert.Assert(t, opts.Balancer.Name() == "weight_random")
		assert.DeepEqual(t, opts.LbOpts, customOpts)
	})

	t.Run("Nil Balancer", func(t *testing.T) {
		var options []ServiceDiscoveryOption
		options = append(options, WithLoadBalanceOptions(nil, loadbalance.DefaultLbOpts))
		opts := &ServiceDiscoveryOptions{}
		opts.Apply(options)
		assert.Assert(t, opts.Balancer == nil)
		assert.DeepEqual(t, opts.LbOpts, loadbalance.DefaultLbOpts)
	})
}
