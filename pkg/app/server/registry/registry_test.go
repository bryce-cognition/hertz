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

package registry

import (
	"errors"
	"net"
	"testing"

	"github.com/cloudwego/hertz/pkg/common/test/assert"
)

func TestNoopRegistry(t *testing.T) {
	reg := noopRegistry{}
	assert.Nil(t, reg.Deregister(&Info{}))
	assert.Nil(t, reg.Register(&Info{}))
}

func TestInfoStruct(t *testing.T) {
	addr, _ := net.ResolveTCPAddr("tcp", "localhost:8080")
	info := &Info{
		ServiceName: "test-service",
		Addr:        addr,
		Weight:      DefaultWeight,
		Tags:        map[string]string{"key": "value"},
	}

	assert.DeepEqual(t, "test-service", info.ServiceName)
	assert.DeepEqual(t, addr, info.Addr)
	assert.DeepEqual(t, DefaultWeight, info.Weight)
	assert.DeepEqual(t, map[string]string{"key": "value"}, info.Tags)
}

type mockRegistry struct {
	registerCalled   bool
	deregisterCalled bool
}

func (m *mockRegistry) Register(info *Info) error {
	m.registerCalled = true
	return nil
}

func (m *mockRegistry) Deregister(info *Info) error {
	m.deregisterCalled = true
	return nil
}

func TestRegistryInterface(t *testing.T) {
	mock := &mockRegistry{}
	info := &Info{ServiceName: "test-service"}

	err := mock.Register(info)
	assert.Nil(t, err)
	assert.True(t, mock.registerCalled)

	err = mock.Deregister(info)
	assert.Nil(t, err)
	assert.True(t, mock.deregisterCalled)
}

type errorMockRegistry struct{}

func (e *errorMockRegistry) Register(info *Info) error {
	return errors.New("register error")
}

func (e *errorMockRegistry) Deregister(info *Info) error {
	return errors.New("deregister error")
}

func TestRegistryInterfaceErrors(t *testing.T) {
	mock := &errorMockRegistry{}
	info := &Info{ServiceName: "test-service"}

	err := mock.Register(info)
	assert.NotNil(t, err)
	assert.DeepEqual(t, "register error", err.Error())

	err = mock.Deregister(info)
	assert.NotNil(t, err)
	assert.DeepEqual(t, "deregister error", err.Error())
}
