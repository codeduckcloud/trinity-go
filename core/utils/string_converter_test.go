package utils

import (
	"reflect"
	"testing"

	"github.com/stretchr/testify/assert"
)

func Test_StringConverter(t *testing.T) {
	type Test struct {
		S1 string
		S2 int64
		S3 int32
		S4 int16
		S5 int
		S6 bool
		S7 interface{}
		S8 struct {
			Test string
			Code int
		}
		S9  *string
		S10 float32
		S11 float64
		S12 *bool
	}
	dest := &Test{}
	{
		val := getStructFieldValue(dest, 0)
		err := StringConverter("123", &val)
		assert.Equal(t, nil, err, "wrong err ")
		assert.Equal(t, "123", dest.S1, "wrong dest ")
	}
	{
		val := getStructFieldValue(dest, 1)
		err := StringConverter("123", &val)
		assert.Equal(t, nil, err, "wrong err ")
		assert.Equal(t, int64(123), dest.S2, "wrong dest ")
	}
	{
		val := getStructFieldValue(dest, 2)
		err := StringConverter("123", &val)
		assert.Equal(t, nil, err, "wrong err ")
		assert.Equal(t, int32(123), dest.S3, "wrong dest ")
	}
	{
		val := getStructFieldValue(dest, 3)
		err := StringConverter("123", &val)
		assert.Equal(t, nil, err, "wrong err ")
		assert.Equal(t, int16(123), dest.S4, "wrong dest ")
	}
	{
		val := getStructFieldValue(dest, 4)
		err := StringConverter("123", &val)
		assert.Equal(t, nil, err, "wrong err ")
		assert.Equal(t, int(123), dest.S5, "wrong dest ")
	}
	{
		val := getStructFieldValue(dest, 5)
		err := StringConverter("true", &val)
		assert.Equal(t, nil, err, "wrong err ")
		assert.Equal(t, true, dest.S6, "wrong dest ")
	}
	{
		val := getStructFieldValue(dest, 1)
		err := StringConverter("aaa", &val)
		assert.Equal(t, "strconv.ParseInt: parsing \"aaa\": invalid syntax", err.Error(), "wrong err ")
	}
	{
		val := getStructFieldValue(dest, 2)
		err := StringConverter("aaa", &val)
		assert.Equal(t, "strconv.ParseInt: parsing \"aaa\": invalid syntax", err.Error(), "wrong err ")
	}
	{
		val := getStructFieldValue(dest, 3)
		err := StringConverter("aaa", &val)
		assert.Equal(t, "strconv.ParseInt: parsing \"aaa\": invalid syntax", err.Error(), "wrong err ")
	}
	{
		val := getStructFieldValue(dest, 4)
		err := StringConverter("aaa", &val)
		assert.Equal(t, "strconv.Atoi: parsing \"aaa\": invalid syntax", err.Error(), "wrong err ")
	}
	{
		val := getStructFieldValue(dest, 6)
		err := StringConverter("true", &val)
		assert.Equal(t, nil, err, "wrong err ")
		assert.Equal(t, (interface{})("true"), dest.S7, "wrong dest ")
	}
	{
		val := getStructFieldValue(dest, 7)
		err := StringConverter(`{"Test":"123","Code":1}`, &val)
		assert.Equal(t, nil, err, "wrong err ")
		assert.Equal(t, struct {
			Test string
			Code int
		}{Test: "123", Code: 1}, dest.S8, "wrong dest ")
	}
	{
		val := getStructFieldValue(dest, 7)
		err := StringConverter(`{"Test":"123","Code":1}123`, &val)
		assert.Equal(t, "invalid character '1' after top-level value", err.Error(), "wrong err ")
	}

	{
		val := getStructFieldValue(dest, 9)
		err := StringConverter("3.4", &val)
		assert.Equal(t, nil, err, "wrong err ")
		assert.Equal(t, float32(3.4), dest.S10, "wrong dest ")
	}
	{
		val := getStructFieldValue(dest, 10)
		err := StringConverter("3.4", &val)
		assert.Equal(t, nil, err, "wrong err ")
		assert.Equal(t, float64(3.4), dest.S11, "wrong dest ")
	}

	{
		val := getStructFieldValue(dest, 8)
		err := StringConverter("true", &val)
		assert.Equal(t, nil, err, "wrong err ")
		assert.Equal(t, "true", *dest.S9, "wrong err ")
	}
	{
		val := getStructFieldValue(dest, 11)
		err := StringConverter("true", &val)
		assert.Equal(t, nil, err, "wrong err ")
		assert.Equal(t, true, *dest.S12, "wrong err ")
	}
	{
		val := getStructFieldValue(dest, 11)
		err := StringConverter("false", &val)
		assert.Equal(t, nil, err, "wrong err ")
		assert.Equal(t, false, *dest.S12, "wrong err ")
	}
}

func getStructFieldValue(dest interface{}, index int) reflect.Value {
	return reflect.Indirect(reflect.ValueOf(dest)).Field(index)
}

func Test_StringConverter_Uints(t *testing.T) {
	type Test struct {
		U1 uint
		U2 uint64
		U3 uint32
		U4 uint16
	}
	dest := &Test{}
	{
		val := getStructFieldValue(dest, 0)
		assert.NoError(t, StringConverter("123", &val))
		assert.Equal(t, uint(123), dest.U1)
	}
	{
		val := getStructFieldValue(dest, 0)
		err := StringConverter("abc", &val)
		assert.Error(t, err)
	}
	{
		val := getStructFieldValue(dest, 1)
		assert.NoError(t, StringConverter("456", &val))
		assert.Equal(t, uint64(456), dest.U2)
	}
	{
		val := getStructFieldValue(dest, 1)
		err := StringConverter("abc", &val)
		assert.Error(t, err)
	}
	{
		val := getStructFieldValue(dest, 2)
		assert.NoError(t, StringConverter("789", &val))
	}
	{
		val := getStructFieldValue(dest, 2)
		err := StringConverter("abc", &val)
		assert.Error(t, err)
	}
	{
		val := getStructFieldValue(dest, 3)
		assert.NoError(t, StringConverter("42", &val))
	}
	{
		val := getStructFieldValue(dest, 3)
		err := StringConverter("abc", &val)
		assert.Error(t, err)
	}
}

func Test_StringConverter_FloatParseErr(t *testing.T) {
	type Test struct {
		F32 float32
		F64 float64
	}
	dest := &Test{}
	{
		val := getStructFieldValue(dest, 0)
		err := StringConverter("abc", &val)
		assert.Error(t, err)
	}
	{
		val := getStructFieldValue(dest, 1)
		err := StringConverter("abc", &val)
		assert.Error(t, err)
	}
}

func Test_StringConverter_PtrErr(t *testing.T) {
	type Test struct {
		S *int
	}
	dest := &Test{}
	val := getStructFieldValue(dest, 0)
	err := StringConverter("abc", &val)
	assert.Error(t, err)
}

func Test_StringConverter_Unsupported(t *testing.T) {
	type Test struct {
		S []string
	}
	dest := &Test{}
	val := getStructFieldValue(dest, 0)
	err := StringConverter("abc", &val)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "unsupported type")
}
