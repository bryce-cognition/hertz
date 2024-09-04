// Copyright 2023 CloudWeGo Authors
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//     http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

package network

import (
	"context"
	"testing"
	"time"

	"github.com/cloudwego/hertz/pkg/common/test/assert"
)

type mockTransporter struct {
	closeCalled     bool
	shutdownCalled  bool
	listenAndServed bool
}

func (m *mockTransporter) Close() error {
	m.closeCalled = true
	return nil
}

func (m *mockTransporter) Shutdown(ctx context.Context) error {
	m.shutdownCalled = true
	return nil
}

func (m *mockTransporter) ListenAndServe(onData OnData) error {
	m.listenAndServed = true
	return nil
}

func TestTransporter(t *testing.T) {
	m := &mockTransporter{}

	t.Run("Close", func(t *testing.T) {
		err := m.Close()
		assert.Nil(t, err)
		assert.True(t, m.closeCalled)
	})

	t.Run("Shutdown", func(t *testing.T) {
		ctx, cancel := context.WithTimeout(context.Background(), time.Second)
		defer cancel()
		err := m.Shutdown(ctx)
		assert.Nil(t, err)
		assert.True(t, m.shutdownCalled)
	})

	t.Run("ListenAndServe", func(t *testing.T) {
		err := m.ListenAndServe(func(ctx context.Context, conn interface{}) error {
			return nil
		})
		assert.Nil(t, err)
		assert.True(t, m.listenAndServed)
	})
}
