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

	"github.com/cloudwego/hertz/pkg/app/server/registry"
	"github.com/cloudwego/hertz/pkg/common/test/assert"
)

// TestDefaultOptions test options with default values
func TestDefaultOptions(t *testing.T) {
	options := NewOptions([]Option{})

	assert.DeepEqual(t, defaultKeepAliveTimeout, options.KeepAliveTimeout)
	assert.DeepEqual(t, defaultReadTimeout, options.ReadTimeout)
	assert.DeepEqual(t, defaultReadTimeout, options.IdleTimeout)
	assert.DeepEqual(t, time.Duration(0), options.WriteTimeout)
	assert.True(t, options.RedirectTrailingSlash)
	assert.True(t, options.RedirectTrailingSlash)
	assert.False(t, options.HandleMethodNotAllowed)
	assert.False(t, options.UseRawPath)
	assert.False(t, options.RemoveExtraSlash)
	assert.True(t, options.UnescapePathValues)
	assert.False(t, options.DisablePreParseMultipartForm)
	assert.False(t, options.SenseClientDisconnection)
	assert.DeepEqual(t, defaultNetwork, options.Network)
	assert.DeepEqual(t, defaultAddr, options.Addr)
	assert.DeepEqual(t, defaultMaxRequestBodySize, options.MaxRequestBodySize)
	assert.False(t, options.GetOnly)
	assert.False(t, options.DisableKeepalive)
	assert.False(t, options.NoDefaultServerHeader)
	assert.DeepEqual(t, defaultWaitExitTimeout, options.ExitWaitTimeout)
	assert.Nil(t, options.TLS)
	assert.DeepEqual(t, defaultReadBufferSize, options.ReadBufferSize)
	assert.False(t, options.ALPN)
	assert.False(t, options.H2C)
	assert.DeepEqual(t, []interface{}{}, options.Tracers)
	assert.DeepEqual(t, new(interface{}), options.TraceLevel)
	assert.DeepEqual(t, registry.NoopRegistry, options.Registry)
	assert.Nil(t, options.BindConfig)
	assert.Nil(t, options.ValidateConfig)
	assert.Nil(t, options.CustomBinder)
	assert.Nil(t, options.CustomValidator)
	assert.DeepEqual(t, false, options.DisableHeaderNamesNormalizing)
}

// TestApplyCustomOptions test apply options with custom values after init
func TestApplyCustomOptions(t *testing.T) {
	// Test initial options
	options := NewOptions([]Option{
		{F: func(o *Options) {
			o.Network = "unix"
			o.Addr = ":9999"
			o.MaxRequestBodySize = 8 * 1024 * 1024
			o.IdleTimeout = 5 * time.Minute
			o.RedirectTrailingSlash = false
		}},
	})

	assert.DeepEqual(t, "unix", options.Network)
	assert.DeepEqual(t, ":9999", options.Addr)
	assert.DeepEqual(t, 8*1024*1024, options.MaxRequestBodySize)
	assert.DeepEqual(t, 5*time.Minute, options.IdleTimeout)
	assert.False(t, options.RedirectTrailingSlash)

	// Test applying an empty option
	options.Apply([]Option{{}})
	assert.DeepEqual(t, "unix", options.Network) // Ensure previous options are not affected

	// Test overwriting a previously set option
	options.Apply([]Option{
		{F: func(o *Options) {
			o.Network = "tcp"
		}},
	})
	assert.DeepEqual(t, "tcp", options.Network)

	// Test applying multiple options at once
	options.Apply([]Option{
		{F: func(o *Options) {
			o.Addr = ":8080"
		}},
		{F: func(o *Options) {
			o.MaxRequestBodySize = 16 * 1024 * 1024
		}},
	})
	assert.DeepEqual(t, ":8080", options.Addr)
	assert.DeepEqual(t, 16*1024*1024, options.MaxRequestBodySize)

	// Test applying an option that doesn't change anything
	options.Apply([]Option{
		{F: func(o *Options) {}},
	})
	assert.DeepEqual(t, "tcp", options.Network)
	assert.DeepEqual(t, ":8080", options.Addr)

	// Test applying an option with a nil function
	options.Apply([]Option{{F: nil}})
	assert.DeepEqual(t, "tcp", options.Network)
	assert.DeepEqual(t, ":8080", options.Addr)
}

// TestIndividualOptions tests individual option functions
func TestIndividualOptions(t *testing.T) {
	t.Run("KeepAliveTimeout", func(t *testing.T) {
		options := NewOptions([]Option{
			{F: func(o *Options) {
				o.KeepAliveTimeout = 2 * time.Minute
			}},
		})
		assert.DeepEqual(t, 2*time.Minute, options.KeepAliveTimeout)
	})

	t.Run("ReadTimeout", func(t *testing.T) {
		options := NewOptions([]Option{
			{F: func(o *Options) {
				o.ReadTimeout = 1 * time.Minute
			}},
		})
		assert.DeepEqual(t, 1*time.Minute, options.ReadTimeout)
	})

	t.Run("WriteTimeout", func(t *testing.T) {
		options := NewOptions([]Option{
			{F: func(o *Options) {
				o.WriteTimeout = 30 * time.Second
			}},
		})
		assert.DeepEqual(t, 30*time.Second, options.WriteTimeout)
	})
}

// TestEdgeCases tests edge cases for options
func TestEdgeCases(t *testing.T) {
	t.Run("ZeroValues", func(t *testing.T) {
		options := NewOptions([]Option{
			{F: func(o *Options) {
				o.MaxRequestBodySize = 0
				o.ReadBufferSize = 0
			}},
		})
		assert.DeepEqual(t, 0, options.MaxRequestBodySize)
		assert.DeepEqual(t, 0, options.ReadBufferSize)
	})

	t.Run("NegativeValues", func(t *testing.T) {
		options := NewOptions([]Option{
			{F: func(o *Options) {
				o.MaxRequestBodySize = -1
				o.ReadBufferSize = -100
			}},
		})
		assert.DeepEqual(t, -1, options.MaxRequestBodySize)
		assert.DeepEqual(t, -100, options.ReadBufferSize)
	})

	t.Run("EmptyStrings", func(t *testing.T) {
		options := NewOptions([]Option{
			{F: func(o *Options) {
				o.Network = ""
				o.Addr = ""
			}},
		})
		assert.DeepEqual(t, "", options.Network)
		assert.DeepEqual(t, "", options.Addr)
	})
}
