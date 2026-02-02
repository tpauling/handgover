// Copyright (c) 2025 tpauling <github@pauling.io>
//
// Permission is hereby granted, free of charge, to any person obtaining a copy of this software and associated documentation files (the “Software”), to deal in the Software without restriction, including without limitation the rights to use, copy, modify, merge, publish, distribute, sublicense, and/or sell copies of the Software, and to permit persons to whom the Software is furnished to do so, subject to the following conditions:
//
// The above copyright notice and this permission notice shall be included in all copies or substantial portions of the Software.
//
// THE SOFTWARE IS PROVIDED “AS IS”, WITHOUT WARRANTY OF ANY KIND, EXPRESS OR IMPLIED, INCLUDING BUT NOT LIMITED TO THE WARRANTIES OF MERCHANTABILITY, FITNESS FOR A PARTICULAR PURPOSE AND NONINFRINGEMENT. IN NO EVENT SHALL THE AUTHORS OR COPYRIGHT HOLDERS BE LIABLE FOR ANY CLAIM, DAMAGES OR OTHER LIABILITY, WHETHER IN AN ACTION OF CONTRACT, TORT OR OTHERWISE, ARISING FROM, OUT OF OR IN CONNECTION WITH THE SOFTWARE OR THE USE OR OTHER DEALINGS IN THE SOFTWARE.
package handgover

import (
	"encoding/json"
	"errors"
	"testing"
	"time"

	"github.com/stretchr/testify/require"
)

func TestReturnValuerAsNilWithoutError(t *testing.T) {
	var s struct {
		String string `foo:"bar"`
	}
	s.String = "hello world"

	sources := []Source{
		NewSourceWithCurrent(
			"foo",
			func(current any, field string) (Valuer, error) {
				require.Equal(t, string("hello world"), current)
				require.Equal(t, "bar", field)
				return nil, nil
			},
		),
	}
	err := From(sources).To(&s)
	require.NoError(t, err)

	require.Equal(t, "hello world", s.String)
}

func TestReturnValuerAsNilWithError(t *testing.T) {
	var s struct {
		String string `foo:"bar"`
	}
	s.String = "hello world"

	sources := []Source{
		NewSourceWithCurrent(
			"foo",
			func(current any, name string) (Valuer, error) {
				require.Equal(t, string("hello world"), current)
				require.Equal(t, "bar", name)

				return nil, errors.New("test error")
			},
		),
	}
	err := From(sources).To(&s)
	require.Error(t, err)

	var parsedErr Error
	require.True(t, errors.As(err, &parsedErr))

	require.Equal(t, "", parsedErr.Value)
	require.Equal(t, "hello world", s.String)
}

func TestFillWithNilStruct(t *testing.T) {

	sources := []Source{
		NewSource(
			"foo",
			func(name string) (Valuer, error) {
				require.Equal(t, "bar", name)
				return Value("helloworld"), nil
			},
		),
	}
	require.Error(t, From(sources).To(nil))
}

func TestFillWithNoSource(t *testing.T) {

	var (
		s struct {
			Pointer *string `foo:"bar"`
		}
		sources []Source
	)
	require.NoError(t, From(sources).To(&s))
}

func TestFillPointer(t *testing.T) {

	var s struct {
		Pointer *string `foo:"bar"`
	}

	sources := []Source{
		NewSourceWithCurrent(
			"foo",
			func(current any, field string) (Valuer, error) {
				require.Equal(t, (*string)(nil), current)
				require.Equal(t, "bar", field)
				return Value("helloworld"), nil
			},
		),
	}
	require.NoError(t, From(sources).To(&s))

	require.NotNil(t, s.Pointer)
	require.Equal(t, "helloworld", *s.Pointer)
}

func TestFillSlice(t *testing.T) {

	var s struct {
		Slice   []string         `foo:"bar"`
		Bytes   []byte           `john:"doe"`
		RawJSON *json.RawMessage `jane:"doe"`
	}

	sources := []Source{
		NewSourceWithCurrent(
			"foo",
			func(current any, field string) (Valuer, error) {
				require.Equal(t, []string(nil), current)
				require.Equal(t, "bar", field)
				return Value("hello", "world"), nil
			},
		),
		NewSourceWithCurrent(
			"john",
			func(current any, field string) (Valuer, error) {
				require.Equal(t, []byte(nil), current)
				require.Equal(t, "doe", field)
				return Value(`{ "some": "json" }`), nil
			},
		),
		NewSourceWithCurrent(
			"jane",
			func(current any, field string) (Valuer, error) {
				require.Equal(t, (*json.RawMessage)(nil), current)
				require.Equal(t, "doe", field)
				return Value(`{ "some": "json" }`), nil
			},
		),
	}

	require.NoError(t, From(sources).To(&s))

	require.Equal(t, []string{"hello", "world"}, s.Slice)
	require.Equal(t, []byte(`{ "some": "json" }`), s.Bytes)

	require.NotNil(t, s.RawJSON)
	require.Equal(t, json.RawMessage(`{ "some": "json" }`), *s.RawJSON)
}

func TestFillSliceWithInvalidValue(t *testing.T) {

	var s struct {
		Slice []int `foo:"bar"`
	}
	s.Slice = []int{1}

	sources := []Source{
		NewSourceWithCurrent(
			"foo",
			func(current any, field string) (Valuer, error) {
				require.Equal(t, []int{1}, current)
				require.Equal(t, "bar", field)
				return Value([]string{"invalid", "value"}...), nil
			},
		),
	}

	err := From(sources).To(&s)
	require.Error(t, err)

	var parsedErr Error
	require.True(t, errors.As(err, &parsedErr))

	require.Equal(t, "bar", parsedErr.Field)
	require.Equal(t, "invalid", parsedErr.Value)
	require.Error(t, parsedErr.InnerError)

	require.Equal(t, []int{1}, s.Slice)
}

func TestFillString(t *testing.T) {

	var s struct {
		String string `foo:"bar"`
	}

	sources := []Source{
		NewSourceWithCurrent(
			"foo",
			func(current any, field string) (Valuer, error) {
				require.Equal(t, string(""), current)
				require.Equal(t, "bar", field)
				return Value("helloworld"), nil
			},
		),
	}
	require.NoError(t, From(sources).To(&s))
	require.Equal(t, "helloworld", s.String)
}

func TestFillTimeDuration(t *testing.T) {

	var s struct {
		Duration time.Duration `foo:"bar"`
	}

	sources := []Source{
		NewSourceWithCurrent(
			"foo",
			func(current any, field string) (Valuer, error) {
				require.Equal(t, time.Duration(0), current)
				require.Equal(t, "bar", field)
				return Value("1h"), nil
			},
		),
	}

	require.NoError(t, From(sources).To(&s))
	require.Equal(t, time.Minute*60, s.Duration)
}

func TestFillTimeDurationWithInvalidValue(t *testing.T) {

	var s struct {
		Duration time.Duration `foo:"bar"`
	}
	s.Duration = time.Second

	sources := []Source{
		NewSourceWithCurrent(
			"foo",
			func(current any, field string) (Valuer, error) {
				require.Equal(t, time.Second, current)
				require.Equal(t, "bar", field)
				return Value("1"), nil
			},
		),
	}

	err := From(sources).To(&s)
	require.Error(t, err)

	var parsedErr Error

	require.True(t, errors.As(err, &parsedErr))
	require.Equal(t, "bar", parsedErr.Field)
	require.Equal(t, "1", parsedErr.Value)
	require.Error(t, parsedErr.InnerError)

	require.Equal(t, time.Second, s.Duration)
}

func TestFillInt(t *testing.T) {

	var s struct {
		Int int `foo:"bar"`
	}

	sources := []Source{
		NewSourceWithCurrent(
			"foo",
			func(current any, field string) (Valuer, error) {
				require.Equal(t, int(0), current)
				require.Equal(t, "bar", field)
				return Value("1"), nil
			},
		),
	}
	require.NoError(t, From(sources).To(&s))
	require.Equal(t, int(1), s.Int)
}

func TestFillIntWithInvalidValue(t *testing.T) {

	var s struct {
		Int int `foo:"bar"`
	}
	s.Int = 1

	sources := []Source{
		NewSourceWithCurrent(
			"foo",
			func(current any, field string) (Valuer, error) {
				require.Equal(t, int(1), current)
				require.Equal(t, "bar", field)
				return Value("invalid"), nil
			},
		),
	}

	err := From(sources).To(&s)
	require.Error(t, err)

	var parsedErr Error

	require.True(t, errors.As(err, &parsedErr))
	require.Equal(t, "bar", parsedErr.Field)
	require.Equal(t, "invalid", parsedErr.Value)
	require.Error(t, parsedErr.InnerError)

	require.Equal(t, 1, s.Int)
}

func TestFillInt8(t *testing.T) {

	var s struct {
		Int8 int8 `foo:"bar"`
	}
	sources := []Source{
		NewSourceWithCurrent(
			"foo",
			func(current any, field string) (Valuer, error) {
				require.Equal(t, int8(0), current)
				require.Equal(t, "bar", field)
				return Value("1"), nil
			},
		),
	}
	require.NoError(t, From(sources).To(&s))
	require.Equal(t, int8(1), s.Int8)
}

func TestFillInt8WithInvalidValue(t *testing.T) {

	var s struct {
		Int8 int8 `foo:"bar"`
	}
	s.Int8 = 1

	sources := []Source{
		NewSourceWithCurrent(
			"foo",
			func(current any, field string) (Valuer, error) {
				require.Equal(t, int8(1), current)
				require.Equal(t, "bar", field)
				return Value("invalid"), nil
			},
		),
	}

	err := From(sources).To(&s)
	require.Error(t, err)

	var parsedErr Error

	require.True(t, errors.As(err, &parsedErr))
	require.Equal(t, "bar", parsedErr.Field)
	require.Equal(t, "invalid", parsedErr.Value)
	require.Error(t, parsedErr.InnerError)
}

func TestFillInt16(t *testing.T) {

	var s struct {
		Int16 int16 `foo:"bar"`
	}

	sources := []Source{
		NewSourceWithCurrent(
			"foo",
			func(current any, field string) (Valuer, error) {
				require.Equal(t, int16(0), current)
				require.Equal(t, "bar", field)
				return Value("1"), nil
			},
		),
	}

	require.NoError(t, From(sources).To(&s))
	require.Equal(t, int16(1), s.Int16)
}

func TestFillInt16WithInvalidValue(t *testing.T) {

	var s struct {
		Int16 int16 `foo:"bar"`
	}
	s.Int16 = int16(1)

	sources := []Source{
		NewSourceWithCurrent(
			"foo",
			func(current any, field string) (Valuer, error) {
				require.Equal(t, int16(1), current)
				require.Equal(t, "bar", field)
				return Value("invalid"), nil
			},
		),
	}

	err := From(sources).To(&s)
	require.Error(t, err)

	var parsedErr Error

	require.True(t, errors.As(err, &parsedErr))
	require.Equal(t, "bar", parsedErr.Field)
	require.Equal(t, "invalid", parsedErr.Value)
	require.Error(t, parsedErr.InnerError)

	require.Equal(t, int16(1), s.Int16)
}

func TestFillInt32(t *testing.T) {

	var s struct {
		Int32 int32 `foo:"bar"`
	}

	sources := []Source{
		NewSourceWithCurrent(
			"foo",
			func(current any, field string) (Valuer, error) {
				require.Equal(t, int32(0), current)
				require.Equal(t, "bar", field)
				return Value("1"), nil
			},
		),
	}

	require.NoError(t, From(sources).To(&s))
	require.Equal(t, int32(1), s.Int32)
}

func TestFillInt32WithInvalidValue(t *testing.T) {

	var s struct {
		Int32 int32 `foo:"bar"`
	}
	s.Int32 = int32(1)

	sources := []Source{
		NewSourceWithCurrent(
			"foo",
			func(current any, field string) (Valuer, error) {
				require.Equal(t, int32(1), current)
				require.Equal(t, "bar", field)
				return Value("invalid"), nil
			},
		),
	}

	err := From(sources).To(&s)
	require.Error(t, err)

	var parsedErr Error

	require.True(t, errors.As(err, &parsedErr))
	require.Equal(t, "bar", parsedErr.Field)
	require.Equal(t, "invalid", parsedErr.Value)
	require.Error(t, parsedErr.InnerError)

	require.Equal(t, int32(1), s.Int32)
}

func TestFillInt64(t *testing.T) {

	var s struct {
		Int64 int64 `foo:"bar"`
	}

	sources := []Source{
		NewSourceWithCurrent(
			"foo",
			func(current any, field string) (Valuer, error) {
				require.Equal(t, int64(0), current)
				require.Equal(t, "bar", field)
				return Value("1"), nil
			},
		),
	}

	require.NoError(t, From(sources).To(&s))
	require.Equal(t, int64(1), s.Int64)
}

func TestFillInt64WithInvalidValue(t *testing.T) {

	var s struct {
		Int64 int64 `foo:"bar"`
	}
	s.Int64 = int64(1)

	sources := []Source{
		NewSourceWithCurrent(
			"foo",
			func(current any, field string) (Valuer, error) {
				require.Equal(t, int64(1), current)
				require.Equal(t, "bar", field)
				return Value("invalid"), nil
			},
		),
	}

	err := From(sources).To(&s)
	require.Error(t, err)

	var parsedErr Error

	require.True(t, errors.As(err, &parsedErr))
	require.Equal(t, "bar", parsedErr.Field)
	require.Equal(t, "invalid", parsedErr.Value)
	require.Error(t, parsedErr.InnerError)

	require.Equal(t, int64(1), s.Int64)
}

func TestFillUInt(t *testing.T) {

	var s struct {
		UInt uint `foo:"bar"`
	}

	sources := []Source{
		NewSourceWithCurrent(
			"foo",
			func(current any, field string) (Valuer, error) {
				require.Equal(t, uint(0), current)
				require.Equal(t, "bar", field)
				return Value("1"), nil
			},
		),
	}

	require.NoError(t, From(sources).To(&s))
	require.Equal(t, uint(1), s.UInt)
}

func TestFillUIntWithInvalidValue(t *testing.T) {

	var s struct {
		UInt uint `foo:"bar"`
	}
	s.UInt = uint(1)

	sources := []Source{
		NewSourceWithCurrent(
			"foo",
			func(current any, field string) (Valuer, error) {
				require.Equal(t, uint(1), current)
				require.Equal(t, "bar", field)
				return Value("invalid"), nil
			},
		),
	}

	err := From(sources).To(&s)
	require.Error(t, err)

	var parsedErr Error

	require.True(t, errors.As(err, &parsedErr))
	require.Equal(t, "bar", parsedErr.Field)
	require.Equal(t, "invalid", parsedErr.Value)
	require.Error(t, parsedErr.InnerError)

	require.Equal(t, uint(1), s.UInt)
}

func TestFillUInt8(t *testing.T) {

	var s struct {
		UInt8 uint8 `foo:"bar"`
	}

	sources := []Source{
		NewSourceWithCurrent(
			"foo",
			func(current any, field string) (Valuer, error) {
				require.Equal(t, uint8(0), current)
				require.Equal(t, "bar", field)
				return Value("1"), nil
			},
		),
	}

	require.NoError(t, From(sources).To(&s))
	require.Equal(t, uint8(1), s.UInt8)
}

func TestFillUInt8WithInvalidValue(t *testing.T) {

	var s struct {
		UInt8 uint8 `foo:"bar"`
	}
	s.UInt8 = uint8(1)

	sources := []Source{
		NewSourceWithCurrent(
			"foo",
			func(current any, field string) (Valuer, error) {
				require.Equal(t, uint8(1), current)
				require.Equal(t, "bar", field)
				return Value("invalid"), nil
			},
		),
	}

	err := From(sources).To(&s)
	require.Error(t, err)

	var parsedErr Error

	require.True(t, errors.As(err, &parsedErr))
	require.Equal(t, "bar", parsedErr.Field)
	require.Equal(t, "invalid", parsedErr.Value)
	require.Error(t, parsedErr.InnerError)

	require.Equal(t, uint8(1), s.UInt8)
}

func TestFillUInt16(t *testing.T) {

	var s struct {
		UInt16 uint16 `foo:"bar"`
	}

	sources := []Source{
		NewSourceWithCurrent(
			"foo",
			func(current any, field string) (Valuer, error) {
				require.Equal(t, uint16(0), current)
				require.Equal(t, "bar", field)
				return Value("1"), nil
			},
		),
	}

	require.NoError(t, From(sources).To(&s))
	require.Equal(t, uint16(1), s.UInt16)
}

func TestFillUInt16WithInvalidValue(t *testing.T) {

	var s struct {
		UInt16 uint16 `foo:"bar"`
	}
	s.UInt16 = uint16(1)

	sources := []Source{
		NewSourceWithCurrent(
			"foo",
			func(current any, field string) (Valuer, error) {
				require.Equal(t, uint16(1), current)
				require.Equal(t, "bar", field)
				return Value("invalid"), nil
			},
		),
	}

	err := From(sources).To(&s)
	require.Error(t, err)

	var parsedErr Error

	require.True(t, errors.As(err, &parsedErr))
	require.Equal(t, "bar", parsedErr.Field)
	require.Equal(t, "invalid", parsedErr.Value)
	require.Error(t, parsedErr.InnerError)

	require.Equal(t, uint16(1), s.UInt16)
}

func TestFillUInt32(t *testing.T) {

	var s struct {
		UInt32 uint32 `foo:"bar"`
	}

	sources := []Source{
		NewSourceWithCurrent(
			"foo",
			func(current any, field string) (Valuer, error) {
				require.Equal(t, uint32(0), current)
				require.Equal(t, "bar", field)
				return Value("1"), nil
			},
		),
	}

	require.NoError(t, From(sources).To(&s))
	require.Equal(t, uint32(1), s.UInt32)
}

func TestFillUInt32WithInvalidValue(t *testing.T) {

	var s struct {
		UInt32 uint32 `foo:"bar"`
	}
	s.UInt32 = uint32(1)

	sources := []Source{
		NewSourceWithCurrent(
			"foo",
			func(current any, field string) (Valuer, error) {
				require.Equal(t, uint32(1), current)
				require.Equal(t, "bar", field)
				return Value("invalid"), nil
			},
		),
	}

	err := From(sources).To(&s)
	require.Error(t, err)

	var parsedErr Error

	require.True(t, errors.As(err, &parsedErr))
	require.Equal(t, "bar", parsedErr.Field)
	require.Equal(t, "invalid", parsedErr.Value)
	require.Error(t, parsedErr.InnerError)

	require.Equal(t, uint32(1), s.UInt32)
}

func TestFillUInt64(t *testing.T) {

	var s struct {
		UInt64 uint64 `foo:"bar"`
	}

	sources := []Source{
		NewSourceWithCurrent(
			"foo",
			func(current any, field string) (Valuer, error) {
				require.Equal(t, uint64(0), current)
				require.Equal(t, "bar", field)
				return Value("1"), nil
			},
		),
	}

	require.NoError(t, From(sources).To(&s))
	require.Equal(t, uint64(1), s.UInt64)
}

func TestFillUInt64WithInvalidValue(t *testing.T) {

	var s struct {
		UInt64 uint64 `foo:"bar"`
	}
	s.UInt64 = uint64(1)

	sources := []Source{
		NewSourceWithCurrent(
			"foo",
			func(current any, field string) (Valuer, error) {
				require.Equal(t, uint64(1), current)
				require.Equal(t, "bar", field)
				return Value("invalid"), nil
			},
		),
	}

	err := From(sources).To(&s)
	require.Error(t, err)

	var parsedErr Error

	require.True(t, errors.As(err, &parsedErr))
	require.Equal(t, "bar", parsedErr.Field)
	require.Equal(t, "invalid", parsedErr.Value)
	require.Error(t, parsedErr.InnerError)

	require.Equal(t, uint64(1), s.UInt64)
}

func TestFillBool(t *testing.T) {

	var s struct {
		Bool bool `foo:"bar"`
	}

	sources := []Source{
		NewSourceWithCurrent(
			"foo",
			func(current any, field string) (Valuer, error) {
				require.Equal(t, false, current)
				require.Equal(t, "bar", field)
				return Value("true"), nil
			},
		),
	}

	require.NoError(t, From(sources).To(&s))
	require.Equal(t, true, s.Bool)
}

func TestFillBoolWithInvalidValue(t *testing.T) {

	var s struct {
		Bool bool `foo:"bar"`
	}
	s.Bool = true

	sources := []Source{
		NewSourceWithCurrent(
			"foo",
			func(current any, field string) (Valuer, error) {
				require.Equal(t, true, current)
				require.Equal(t, "bar", field)
				return Value("invalid"), nil
			},
		),
	}

	err := From(sources).To(&s)
	require.Error(t, err)

	var parsedErr Error

	require.True(t, errors.As(err, &parsedErr))
	require.Equal(t, "bar", parsedErr.Field)
	require.Equal(t, "invalid", parsedErr.Value)
	require.Error(t, parsedErr.InnerError)

	require.True(t, s.Bool)
}

func TestFillFloat32(t *testing.T) {

	var s struct {
		Float32 float32 `foo:"bar"`
	}
	sources := []Source{
		NewSourceWithCurrent(
			"foo",
			func(current any, field string) (Valuer, error) {
				require.Equal(t, float32(0), current)
				require.Equal(t, "bar", field)
				return Value("1.5"), nil
			},
		),
	}

	require.NoError(t, From(sources).To(&s))
	require.Equal(t, float32(1.5), s.Float32)
}

func TestFillFloat32WithInvalidValue(t *testing.T) {

	var s struct {
		Float32 float32 `foo:"bar"`
	}
	s.Float32 = float32(1.5)

	sources := []Source{
		NewSourceWithCurrent(
			"foo",
			func(current any, field string) (Valuer, error) {
				require.Equal(t, float32(1.5), current)
				require.Equal(t, "bar", field)
				return Value("invalid"), nil
			},
		),
	}

	err := From(sources).To(&s)
	require.Error(t, err)

	var parsedErr Error

	require.True(t, errors.As(err, &parsedErr))
	require.Equal(t, "bar", parsedErr.Field)
	require.Equal(t, "invalid", parsedErr.Value)
	require.Error(t, parsedErr.InnerError)

	require.Equal(t, float32(1.5), s.Float32)
}

func TestFillFloat64(t *testing.T) {

	var s struct {
		Float64 float64 `foo:"bar"`
	}
	sources := []Source{
		NewSourceWithCurrent(
			"foo",
			func(current any, field string) (Valuer, error) {
				require.Equal(t, float64(0), current)
				require.Equal(t, "bar", field)
				return Value("1.5"), nil
			},
		),
	}

	require.NoError(t, From(sources).To(&s))
	require.Equal(t, float64(1.5), s.Float64)
}

func TestFillFloat64WithInvalidValue(t *testing.T) {

	var s struct {
		Float64 float64 `foo:"bar"`
	}
	s.Float64 = float64(1.5)

	sources := []Source{
		NewSourceWithCurrent(
			"foo",
			func(current any, field string) (Valuer, error) {
				require.Equal(t, float64(1.5), current)
				require.Equal(t, "bar", field)
				return Value("invalid"), nil
			},
		),
	}

	err := From(sources).To(&s)
	require.Error(t, err)

	var parsedErr Error

	require.True(t, errors.As(err, &parsedErr))
	require.Equal(t, "bar", parsedErr.Field)
	require.Equal(t, "invalid", parsedErr.Value)
	require.Error(t, parsedErr.InnerError)

	require.Equal(t, float64(1.5), s.Float64)
}

func TestFillStruct(t *testing.T) {

	var s struct {
		Struct struct {
			Hello string `json:"hello"`
		} `foo:"bar"`
	}
	sources := []Source{
		NewSourceWithCurrent(
			"foo",
			func(current any, field string) (Valuer, error) {
				require.Equal(t, s.Struct, current)
				require.Equal(t, "bar", field)
				return Value(`{ "hello" : "world" }`), nil
			},
		),
	}

	require.NoError(t, From(sources).To(&s))
	require.Equal(t, "world", s.Struct.Hello)
}

func TestFillStructWithInvalidJson(t *testing.T) {

	var s struct {
		Struct struct {
			Hello string `json:"hello"`
		} `foo:"bar"`
	}
	s.Struct.Hello = "world"

	sources := []Source{
		NewSourceWithCurrent(
			"foo",
			func(current any, field string) (Valuer, error) {
				require.Equal(t, s.Struct, current)
				require.Equal(t, "world", s.Struct.Hello)
				require.Equal(t, "bar", field)
				return Value(`{ "hello" : invalidjson`), nil
			},
		),
	}

	err := From(sources).To(&s)
	require.Error(t, err)

	var parsedErr Error

	require.True(t, errors.As(err, &parsedErr))
	require.Equal(t, "bar", parsedErr.Field)
	require.Equal(t, `{ "hello" : invalidjson`, parsedErr.Value)
	require.Error(t, parsedErr.InnerError)

	require.Equal(t, "world", s.Struct.Hello)
}

func TestFillUnsupportedType(t *testing.T) {

	var s struct {
		Chan chan string `foo:"bar"`
	}

	sources := []Source{
		NewSourceWithCurrent(
			"foo",
			func(current any, field string) (Valuer, error) {
				require.Equal(t, chan string(nil), current)
				require.Equal(t, "bar", field)
				return Value("helloworld"), nil
			},
		),
	}

	err := From(sources).To(&s)
	require.Error(t, err)

	var parsedErr Error

	require.True(t, errors.As(err, &parsedErr))
	require.Equal(t, "bar", parsedErr.Field)
	require.Error(t, parsedErr.InnerError)
}

func TestFillIfSourceReturnsAnError(t *testing.T) {

	var s struct {
		String string `foo:"bar"`
	}
	s.String = "hello world"

	sources := []Source{
		NewSourceWithCurrent(
			"foo",
			func(current any, field string) (Valuer, error) {
				require.Equal(t, "hello world", s.String)
				require.Equal(t, "bar", field)
				return Value(""), errors.New("I am a test error")
			},
		),
	}

	err := From(sources).To(&s)
	require.Error(t, err)

	var parsedErr Error

	require.True(t, errors.As(err, &parsedErr))
	require.Equal(t, "bar", parsedErr.Field)
	require.Error(t, parsedErr.InnerError)
	require.Equal(t, "I am a test error", parsedErr.InnerError.Error())

	require.Equal(t, "hello world", s.String)
}
