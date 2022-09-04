package httpx

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"
	"reflect"

	"github.com/codeduckcloud/trinity-go/core/utils"
	"github.com/go-chi/chi/v5"
)

var (
	contextType    = reflect.ValueOf(context.Background()).Type()
	httpWriterType = reflect.ValueOf(NewWriter()).Type()
	requestType    = reflect.ValueOf(&http.Request{}).Type()
)

type w struct {
}

func (w w) Header() http.Header {
	return nil
}
func (w w) Write([]byte) (int, error) {
	return 0, nil
}
func (w w) WriteHeader(statusCode int) {
}
func NewWriter() http.ResponseWriter {
	return w{}
}

type HTTPContextKey string

const (
	HttpxContext HTTPContextKey = "HTTPX_CONTEXT_KEY"
)

type Context struct {
	r    *http.Request
	code int
}

func NewContext(r *http.Request, code int) *Context {
	return &Context{
		r:    r,
		code: code,
	}
}

func GetHTTPStatusCode(ctx context.Context, defaultStatus int) int {
	if val, ok := ctx.Value(HttpxContext).(*Context); ok {
		if val != nil {
			if val.code != 0 {
				return val.code
			}
		}
	}
	return defaultStatus
}

func SetHttpStatusCode(ctx context.Context, status int) {
	val, ok := ctx.Value(HttpxContext).(*Context)
	if !ok {
		panic("httpx context not set ")
	}
	val.code = status
}

func GetRawRequest(ctx context.Context) *http.Request {
	val, ok := ctx.Value(HttpxContext).(*Context)
	if !ok {
		panic("httpx context not set ")
	}
	return val.r
}

func Parse(r *http.Request, v interface{}) error {
	if v == nil {
		return fmt.Errorf("parsing error , empty value to parse")
	}
	destVal := reflect.Indirect(reflect.ValueOf(v))
	inType := destVal.Type()
	for index := 0; index < destVal.NumField(); index++ {
		val := destVal.Field(index)
		if !val.CanSet() {
			return fmt.Errorf("di param : %v is not exported , cannot set", inType.Field(index).Name)
		}
		if headerParam, isExist := inType.Field(index).Tag.Lookup("header_param"); isExist {
			if _defaultHeaderParser.Exist(r.Header, headerParam) {
				headerValString := _defaultHeaderParser.Get(r.Header, headerParam)
				if err := utils.StringConverter(headerValString, &val); err != nil {
					return fmt.Errorf("header param %v converted error, cannot set ,err:%v ,  val : %v  ", inType.Field(index).Name, err, headerValString)
				}
				continue
			}
		}
		// check if path param
		if pathParam, isExist := inType.Field(index).Tag.Lookup("path_param"); isExist {
			paramValString := chi.URLParam(r, pathParam)
			if err := utils.StringConverter(paramValString, &val); err != nil {
				return fmt.Errorf("path param %v converted error, cannot set , err:%v  val : %v  ", inType.Field(index).Name, err, paramValString)
			}
			continue
		}
		// check if query param
		if queryParam, isExist := inType.Field(index).Tag.Lookup("query_param"); isExist {
			if queryParam == "" {
				switch val.Type().Kind() {
				case reflect.String:
					val.Set(reflect.ValueOf(r.URL.RawQuery))
				case reflect.Map:
					switch inType.Field(index).Type.String() {
					case "url.Values":
						val.Set(reflect.ValueOf(r.URL.Query()))
					case "map[string][]string":
						res := make(map[string][]string)
						for k := range r.URL.Query() {
							res[k] = r.URL.Query()[k]
						}
						val.Set(reflect.ValueOf(res))
					case "map[string]string":
						res := make(map[string]string)
						for k := range r.URL.Query() {
							res[k] = r.URL.Query().Get(k)
						}
						val.Set(reflect.ValueOf(res))
					case "map[string]interface {}":
						res := make(map[string]interface{})
						for k := range r.URL.Query() {
							res[k] = r.URL.Query().Get(k)
						}
						val.Set(reflect.ValueOf(res))
					default:
						return fmt.Errorf("unsupported map type to decode query param , actual:%v", inType.Field(index).Type.String())
					}
				default:
					return fmt.Errorf("param %v get all query param converted error, only support string , val : %v ", inType.Field(index).Name, r.URL.RawQuery)
				}
			} else {
				if _defaultQueryParser.Exist(r.URL.Query(), queryParam) {
					queryValString := _defaultQueryParser.Get(r.URL.Query(), queryParam)
					if err := utils.StringConverter(queryValString, &val); err != nil {
						return fmt.Errorf("param %v converted error, err :%v , val : %v ", inType.Field(index).Name, err, queryValString)
					}
				}
			}
			continue
		}
		// check if body param
		if bodyParam, isExist := inType.Field(index).Tag.Lookup("body_param"); isExist {
			respBytes, err := ioutil.ReadAll(r.Body)
			if err != nil {
				return fmt.Errorf("read request body error  , err : %v ", err)
			}
			r.Body = ioutil.NopCloser(bytes.NewBuffer(respBytes))
			if bodyParam == "" {
				switch val.Type().Kind() {
				case reflect.String:
					val.Set(reflect.ValueOf(string(respBytes)))
				case reflect.Struct, reflect.Slice:
					// if is []byte
					if fmt.Sprintf("%v", inType.Field(index).Type) == "[]uint8" {
						val.Set(reflect.Indirect(reflect.ValueOf(respBytes)))
					} else {
						targetVal := reflect.New(inType.Field(index).Type).Interface()
						if err := json.Unmarshal(respBytes, targetVal); err != nil {
							return fmt.Errorf("param %v converted error, err :%v , val : %v ", inType.Field(index).Name, err, string(respBytes))
						}
						val.Set(reflect.Indirect(reflect.ValueOf(targetVal)))
					}
				case reflect.Map:
					if fmt.Sprintf("%v", inType.Field(index).Type) != "map[string]interface {}" {
						return fmt.Errorf("param %v converted error, map only support map[string]interface{}, val : %v ", inType.Field(index).Name, string(respBytes))
					}
					bodyVal := make(map[string]interface{})
					if len(respBytes) > 0 {
						d := json.NewDecoder(bytes.NewReader(respBytes))
						d.UseNumber()
						if err := d.Decode(&bodyVal); err != nil {
							return fmt.Errorf("param %v converted error,err :%v , val : %v ", inType.Field(index).Name, err, string(respBytes))
						}
					}
					val.Set(reflect.ValueOf(bodyVal))
				case reflect.Interface:
					var bodyVal interface{}
					if len(respBytes) > 0 {
						if err := json.Unmarshal(respBytes, &bodyVal); err != nil {
							return fmt.Errorf("param %v converted error,err :%v , val : %v ", inType.Field(index).Name, err, string(respBytes))
						}
					}
					val.Set(reflect.ValueOf(bodyVal))
				case reflect.Ptr:
					newDest := reflect.New(val.Type().Elem()).Interface()
					if len(respBytes) > 0 {
						if err := json.Unmarshal(respBytes, newDest); err != nil {
							return fmt.Errorf("param %v converted error,err :%v , val : %v ", inType.Field(index).Name, err, string(respBytes))
						}
					}
					val.Set(reflect.ValueOf(newDest))
				default:
					return fmt.Errorf("unsupported type , only support string , struct ,Slice ,  map[string]interface{} , interface{} , []byte, actual: %v", val.Type().Kind())
				}
			} else {
				bodyVal := make(map[string]interface{})
				if len(respBytes) > 0 {
					if err := json.Unmarshal(respBytes, &bodyVal); err != nil {
						return fmt.Errorf("param %v converted error,err :%v , val : %v ", inType.Field(index).Name, err, string(respBytes))
					}
				}
				value, err := bodyParamConverter(bodyVal, bodyParam, inType.Field(index).Type)
				if err != nil {
					return fmt.Errorf("param %v converted error,err :%v , val : %v ", inType.Field(index).Name, err, bodyVal)
				}
				val.Set(reflect.ValueOf(value))
			}
			continue
		}
		switch val.Kind() {
		case reflect.Struct:
			newDest := reflect.New(val.Type()).Interface()
			if err := Parse(r, newDest); err != nil {
				return err
			}
			val.Set(reflect.ValueOf(newDest).Elem())
		case reflect.Ptr:
			newDest := reflect.New(val.Type().Elem()).Interface()
			if err := Parse(r, newDest); err != nil {
				return err
			}
			val.Set(reflect.ValueOf(newDest))
		}
	}
	if err := _defaultValidator.Struct(v); err != nil {
		return fmt.Errorf("httpx.Parse validate error, err: %v", err)
	}
	return nil
}

// bodyParamConverter
/*
 @bodyVal the father val of
 @key the key name of the parentVal
 @destType the type of dest type
 bodyParamConverter will get the key from the body value
 and convert the value to the dest type value
*/
func bodyParamConverter(bodyVal map[string]interface{}, key string, destType reflect.Type) (interface{}, error) {
	value, ok := bodyVal[key]
	if !ok {
		return nil, fmt.Errorf("key %v not exist", key)
	}
	switch destType.Kind() {
	case reflect.Int64:
		convertedValue, ok := value.(int64)
		if !ok {
			return nil, fmt.Errorf("key %v convert to int64 error ", key)
		}
		return convertedValue, nil
	case reflect.Int32:
		convertedValue, ok := value.(int32)
		if !ok {
			return nil, fmt.Errorf("key %v convert to int32 error ", key)
		}
		return convertedValue, nil
	case reflect.Int:
		convertedValue, ok := value.(int)
		if !ok {
			return nil, fmt.Errorf("key %v convert to int error ", key)
		}
		return convertedValue, nil
	case reflect.String:
		convertedValue, ok := value.(string)
		if !ok {
			return nil, fmt.Errorf("key %v convert to string error ", key)
		}
		return convertedValue, nil
	case reflect.Struct:
		c, _ := json.Marshal(value)
		targetVal := reflect.New(destType).Interface()
		decoder := json.NewDecoder(bytes.NewReader(c))
		if err := decoder.Decode(targetVal); err != nil {
			return nil, err
		}
		return reflect.Indirect(reflect.ValueOf(targetVal)).Interface(), nil
	default:
		return nil, fmt.Errorf("type %v not support", destType)
	}

}
