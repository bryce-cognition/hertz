package render

import (
	"testing"

	"github.com/cloudwego/hertz/pkg/common/test/assert"
	"github.com/cloudwego/hertz/pkg/protocol"
)

func TestStringRender(t *testing.T) {
	tests := []struct {
		name     string
		s        String
		expected string
	}{
		{
			name:     "Empty Format and Data",
			s:        String{Format: "", Data: nil},
			expected: "",
		},
		{
			name:     "Format with no placeholders",
			s:        String{Format: "Hello, World!", Data: nil},
			expected: "Hello, World!",
		},
		{
			name:     "Format with placeholders and matching Data",
			s:        String{Format: "Hello, %s! You are %d years old.", Data: []interface{}{"Alice", 30}},
			expected: "Hello, Alice! You are 30 years old.",
		},
		{
			name:     "Format with placeholders and empty Data",
			s:        String{Format: "Hello, %s!", Data: []interface{}{}},
			expected: "Hello, %s!",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			resp := &protocol.Response{}
			err := tt.s.Render(resp)
			assert.Nil(t, err)
			assert.DeepEqual(t, tt.expected, string(resp.Body()))
		})
	}
}

func TestStringWriteContentType(t *testing.T) {
	s := String{}
	resp := &protocol.Response{}
	s.WriteContentType(resp)
	assert.DeepEqual(t, []byte(plainContentType), resp.Header.Peek("Content-Type"))
}
