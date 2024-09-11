package binding

import (
	stdJson "encoding/json"
	"reflect"
	"testing"
	"time"

	"github.com/cloudwego/hertz/pkg/common/test/assert"
	hJson "github.com/cloudwego/hertz/pkg/common/json"
	"github.com/cloudwego/hertz/pkg/protocol"
	"github.com/cloudwego/hertz/pkg/route/param"
)

func TestNewBindConfig(t *testing.T) {
	config := NewBindConfig()
	assert.DeepEqual(t, false, config.LooseZeroMode)
	assert.DeepEqual(t, false, config.DisableDefaultTag)
	assert.DeepEqual(t, false, config.DisableStructFieldResolve)
	assert.DeepEqual(t, false, config.EnableDecoderUseNumber)
	assert.DeepEqual(t, false, config.EnableDecoderDisallowUnknownFields)
	assert.NotNil(t, config.TypeUnmarshalFuncs)
	assert.NotNil(t, config.Validator)
}

func TestRegTypeUnmarshalConfig(t *testing.T) {
	config := NewBindConfig()
	type CustomType struct{}
	err := config.RegTypeUnmarshal(reflect.TypeOf(CustomType{}), func(req *protocol.Request, params param.Params, text string) (reflect.Value, error) {
		return reflect.Value{}, nil
	})
	assert.Nil(t, err)

	// Test registering basic type (should fail)
	err = config.RegTypeUnmarshal(reflect.TypeOf(""), nil)
	assert.NotNil(t, err)

	// Test registering pointer type (should fail)
	err = config.RegTypeUnmarshal(reflect.TypeOf(&CustomType{}), nil)
	assert.NotNil(t, err)
}

func TestMustRegTypeUnmarshal(t *testing.T) {
	config := NewBindConfig()
	type CustomType struct{}

	notPanicked := true
	func() {
		defer func() {
			if r := recover(); r != nil {
				notPanicked = false
			}
		}()
		config.MustRegTypeUnmarshal(reflect.TypeOf(CustomType{}), func(req *protocol.Request, params param.Params, text string) (reflect.Value, error) {
			return reflect.Value{}, nil
		})
	}()
	assert.True(t, notPanicked)

	// Test registering invalid type (should panic)
	panicked := false
	func() {
		defer func() {
			if r := recover(); r != nil {
				panicked = true
			}
		}()
		config.MustRegTypeUnmarshal(reflect.TypeOf(""), nil)
	}()
	assert.True(t, panicked)
}

func TestUseThirdPartyJSONUnmarshaler(t *testing.T) {
	config := NewBindConfig()
	customUnmarshaler := func(data []byte, v interface{}) error {
		return nil
	}
	config.UseThirdPartyJSONUnmarshaler(customUnmarshaler)
	assert.DeepEqual(t, reflect.ValueOf(customUnmarshaler).Pointer(), reflect.ValueOf(hJson.Unmarshal).Pointer())
}

func TestUseStdJSONUnmarshaler(t *testing.T) {
	config := NewBindConfig()
	config.UseStdJSONUnmarshaler()
	assert.DeepEqual(t, reflect.ValueOf(stdJson.Unmarshal).Pointer(), reflect.ValueOf(hJson.Unmarshal).Pointer())
}

func TestInitTypeUnmarshal(t *testing.T) {
	config := NewBindConfig()
	config.initTypeUnmarshal()

	// Test if time.Time is registered
	timeType := reflect.TypeOf(time.Time{})
	_, exists := config.TypeUnmarshalFuncs[timeType]
	assert.True(t, exists)

	// Test the registered function
	req := &protocol.Request{}
	params := param.Params{}
	text := "2023-05-18T10:00:00Z"

	value, err := config.TypeUnmarshalFuncs[timeType](req, params, text)
	assert.Nil(t, err)

	parsedTime, ok := value.Interface().(time.Time)
	assert.True(t, ok)
	assert.DeepEqual(t, "2023-05-18 10:00:00 +0000 UTC", parsedTime.String())

	// Test with invalid time string
	_, err = config.TypeUnmarshalFuncs[timeType](req, params, "invalid-time")
	assert.NotNil(t, err)
}
