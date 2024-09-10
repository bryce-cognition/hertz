package loadbalance

import (
	"testing"
	"time"

	"github.com/cloudwego/hertz/pkg/common/test/assert"
)

func TestOptionsCheck(t *testing.T) {
	tests := []struct {
		name     string
		options  Options
		expected Options
	}{
		{
			name:     "Default values",
			options:  Options{},
			expected: Options{RefreshInterval: DefaultRefreshInterval, ExpireInterval: DefaultExpireInterval},
		},
		{
			name:     "Custom valid values",
			options:  Options{RefreshInterval: 10 * time.Second, ExpireInterval: 30 * time.Second},
			expected: Options{RefreshInterval: 10 * time.Second, ExpireInterval: 30 * time.Second},
		},
		{
			name:     "Invalid RefreshInterval",
			options:  Options{RefreshInterval: -1 * time.Second, ExpireInterval: 30 * time.Second},
			expected: Options{RefreshInterval: DefaultRefreshInterval, ExpireInterval: 30 * time.Second},
		},
		{
			name:     "Invalid ExpireInterval",
			options:  Options{RefreshInterval: 10 * time.Second, ExpireInterval: -1 * time.Second},
			expected: Options{RefreshInterval: 10 * time.Second, ExpireInterval: DefaultExpireInterval},
		},
		{
			name:     "Both intervals invalid",
			options:  Options{RefreshInterval: -1 * time.Second, ExpireInterval: -1 * time.Second},
			expected: Options{RefreshInterval: DefaultRefreshInterval, ExpireInterval: DefaultExpireInterval},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tt.options.Check()
			assert.DeepEqual(t, tt.expected, tt.options)
		})
	}
}
