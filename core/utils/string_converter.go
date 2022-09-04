package utils

import (
	"encoding/json"
	"fmt"
	"reflect"
	"strconv"
)

// StringConverter
/*
 @word the word need to be converted
 @destVal the target value need to convert to

 stringConverter will convert to destVal according to
 the type of destVal
*/
func StringConverter(word string, destVal *reflect.Value) error {
	switch destVal.Type().Kind() {
	case reflect.Int64:
		paramVal, err := strconv.ParseInt(word, 10, 64)
		if err != nil {
			return err
		}
		destVal.Set(reflect.ValueOf(paramVal))
	case reflect.Int32:
		paramVal, err := strconv.ParseInt(word, 10, 32)
		if err != nil {
			return err
		}
		destVal.Set(reflect.ValueOf(int32(paramVal)))
	case reflect.Int16:
		paramVal, err := strconv.ParseInt(word, 10, 16)
		if err != nil {
			return err
		}
		destVal.Set(reflect.ValueOf(int16(paramVal)))
	case reflect.Int:
		paramVal, err := strconv.Atoi(word)
		if err != nil {
			return err
		}
		destVal.Set(reflect.ValueOf(paramVal))
	case reflect.Uint:
		paramVal, err := strconv.Atoi(word)
		if err != nil {
			return err
		}
		destVal.Set(reflect.ValueOf(uint(paramVal)))
	case reflect.Uint64:
		paramVal, err := strconv.ParseUint(word, 10, 64)
		if err != nil {
			return err
		}
		destVal.Set(reflect.ValueOf(paramVal))
	case reflect.Uint32:
		paramVal, err := strconv.ParseUint(word, 10, 32)
		if err != nil {
			return err
		}
		destVal.Set(reflect.ValueOf(int32(paramVal)))
	case reflect.Uint16:
		paramVal, err := strconv.ParseUint(word, 10, 16)
		if err != nil {
			return err
		}
		destVal.Set(reflect.ValueOf(int16(paramVal)))
	case reflect.Float32:
		paramVal, err := strconv.ParseFloat(word, 32)
		if err != nil {
			return err
		}
		destVal.Set(reflect.ValueOf(float32(paramVal)))
	case reflect.Float64:
		paramVal, err := strconv.ParseFloat(word, 64)
		if err != nil {
			return err
		}
		destVal.Set(reflect.ValueOf(paramVal))
	case reflect.Bool:
		paramVal, _ := strconv.ParseBool(word)
		destVal.Set(reflect.ValueOf(paramVal))
	case reflect.String:
		destVal.Set(reflect.ValueOf(word))
	case reflect.Interface:
		destVal.Set(reflect.ValueOf(word))
	case reflect.Struct:
		targetVal := reflect.New(destVal.Type()).Interface()
		if err := json.Unmarshal([]byte(word), targetVal); err != nil {
			return err
		}
		destVal.Set(reflect.ValueOf(targetVal).Elem())
	case reflect.Ptr:
		targetVal := reflect.New(destVal.Type().Elem())
		dest := targetVal.Elem()
		if err := StringConverter(word, &dest); err != nil {
			return err
		}
		destVal.Set(targetVal)
	default:
		return fmt.Errorf("unsupported type")
	}
	return nil
}
