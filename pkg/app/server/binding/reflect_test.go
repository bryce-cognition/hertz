/*
 * Copyright 2023 CloudWeGo Authors
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

package binding

import (
	"reflect"
	"testing"

	"github.com/cloudwego/hertz/pkg/common/test/assert"
)

type foo struct {
	f1 string
}

func TestReflect_TypeID(t *testing.T) {
	_, intType := valueAndTypeID(int(1))
	_, uintType := valueAndTypeID(uint(1))
	_, shouldBeIntType := valueAndTypeID(int(1))
	assert.DeepEqual(t, intType, shouldBeIntType)
	assert.NotEqual(t, intType, uintType)

	foo1 := foo{f1: "1"}
	foo2 := foo{f1: "2"}
	_, foo1Type := valueAndTypeID(foo1)
	_, foo2Type := valueAndTypeID(foo2)
	_, foo2PointerType := valueAndTypeID(&foo2)
	_, foo1PointerType := valueAndTypeID(&foo1)
	assert.DeepEqual(t, foo1Type, foo2Type)
	assert.NotEqual(t, foo1Type, foo2PointerType)
	assert.DeepEqual(t, foo1PointerType, foo2PointerType)
}

func TestReflect_CheckPointer(t *testing.T) {
	testCases := []struct {
		name      string
		input     interface{}
		expectErr bool
	}{
		{"Non-pointer", foo{}, true},
		{"Valid pointer", &foo{}, false},
		{"Nil pointer", (*foo)(nil), true},
		{"Double pointer", new(*foo), false},
		{"Nil double pointer", (**foo)(nil), true},
		{"Integer", 42, true},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			val := reflect.ValueOf(tc.input)
			err := checkPointer(val)
			if tc.expectErr {
				assert.NotNil(t, err)
			} else {
				assert.Nil(t, err)
			}
		})
	}
}

func TestReflect_DereferPointer(t *testing.T) {
	testCases := []struct {
		name         string
		input        interface{}
		expectedName string
	}{
		{"Four-level pointer", new(***foo), "foo"},
		{"Three-level pointer", new(**foo), "foo"},
		{"Two-level pointer", new(*foo), "foo"},
		{"One-level pointer", new(foo), "foo"},
		{"Non-pointer", foo{}, "foo"},
		{"Integer pointer", new(int), "int"},
		{"String pointer", new(string), "string"},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			val := reflect.ValueOf(tc.input)
			rt := dereferPointer(val)
			assert.NotEqual(t, reflect.Ptr, rt.Kind())
			assert.DeepEqual(t, tc.expectedName, rt.Name())
		})
	}
}
