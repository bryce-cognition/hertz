package binding

import (
	"reflect"
	"testing"
	"time"

	"github.com/cloudwego/hertz/pkg/common/test/assert"
	"github.com/cloudwego/hertz/pkg/protocol"
	"github.com/cloudwego/hertz/pkg/route/param"
	hJson "github.com/cloudwego/hertz/pkg/common/json"
	stdJson "encoding/json"
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

func TestBindConfig_RegTypeUnmarshal(t *testing.T) {
	config := NewBindConfig()
	customType := reflect.TypeOf(time.Time{})
	err := config.RegTypeUnmarshal(customType, func(req *protocol.Request, params param.Params, text string) (reflect.Value, error) {
		return reflect.Value{}, nil
	})
	assert.Nil(t, err)
	assert.NotNil(t, config.TypeUnmarshalFuncs[customType])

	// Test registering basic type (should fail)
	err = config.RegTypeUnmarshal(reflect.TypeOf(""), nil)
	assert.NotNil(t, err)
	assert.DeepEqual(t, "registration type cannot be a basic type", err.Error())

	// Test registering pointer type (should fail)
	err = config.RegTypeUnmarshal(reflect.TypeOf(&struct{}{}), nil)
	assert.NotNil(t, err)
	assert.DeepEqual(t, "registration type cannot be a pointer type", err.Error())
}

func TestBindConfig_MustRegTypeUnmarshal(t *testing.T) {
	config := NewBindConfig()
	customType := reflect.TypeOf(time.Time{})
	assert.NotPanic(t, func() {
		config.MustRegTypeUnmarshal(customType, func(req *protocol.Request, params param.Params, text string) (reflect.Value, error) {
			return reflect.Value{}, nil
		})
	})
	assert.NotNil(t, config.TypeUnmarshalFuncs[customType])

	// Test registering basic type (should panic)
	assert.Panic(t, func() {
		config.MustRegTypeUnmarshal(reflect.TypeOf(""), nil)
	})

	// Test registering pointer type (should panic)
	assert.Panic(t, func() {
		config.MustRegTypeUnmarshal(reflect.TypeOf(&struct{}{}), nil)
	})
}

func TestBindConfig_UseThirdPartyJSONUnmarshaler(t *testing.T) {
	config := NewBindConfig()
	customUnmarshaler := func(data []byte, v interface{}) error {
		return nil
	}
	config.UseThirdPartyJSONUnmarshaler(customUnmarshaler)
	assert.DeepEqual(t, reflect.ValueOf(customUnmarshaler).Pointer(), reflect.ValueOf(hJson.Unmarshal).Pointer())
}

func TestBindConfig_UseStdJSONUnmarshaler(t *testing.T) {
	config := NewBindConfig()
	config.UseStdJSONUnmarshaler()
	assert.DeepEqual(t, reflect.ValueOf(stdJson.Unmarshal).Pointer(), reflect.ValueOf(hJson.Unmarshal).Pointer())
}

func TestNewValidateConfig(t *testing.T) {
	config := NewValidateConfig()
	assert.NotNil(t, config)
	assert.DeepEqual(t, "", config.ValidateTag)
	assert.Nil(t, config.ErrFactory)
}

func TestValidateConfig_MustRegValidateFunc(t *testing.T) {
	config := NewValidateConfig()
	assert.NotPanic(t, func() {
		config.MustRegValidateFunc("testFunc", func(args ...interface{}) error {
			return nil
		})
	})

	// Test registering duplicate function (should not panic)
	assert.NotPanic(t, func() {
		config.MustRegValidateFunc("testFunc", func(args ...interface{}) error {
			return nil
		}, true)
	})
}

func TestValidateConfig_SetValidatorErrorFactory(t *testing.T) {
	config := NewValidateConfig()
	customErrFactory := func(fieldSelector, msg string) error {
		return nil
	}
	config.SetValidatorErrorFactory(customErrFactory)
	assert.NotNil(t, config.ErrFactory)
	assert.DeepEqual(t, reflect.ValueOf(customErrFactory).Pointer(), reflect.ValueOf(config.ErrFactory).Pointer())
}

func TestValidateConfig_SetValidatorTag(t *testing.T) {
	config := NewValidateConfig()
	config.SetValidatorTag("customTag")
	assert.DeepEqual(t, "customTag", config.ValidateTag)
}
