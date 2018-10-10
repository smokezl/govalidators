package govalidators

import (
	"errors"
	"reflect"
	"strconv"
	"strings"
)

type Kind uint

const (
	ALL_KIND Kind = iota
	ARRAY_KIND
	SLICE_KIND
	MAP_KIND
	STRING_KIND
	INTEGER_KIND
	FLOAT_KIND
	BOOL_KIND
)

//判断是否为 array、map、slice 的 map
var arrayMap = map[reflect.Kind]Kind{
	reflect.Array: ARRAY_KIND,
	reflect.Slice: SLICE_KIND,
	reflect.Map:   MAP_KIND,
}

//判断是否为字符串
var stringMap = map[reflect.Kind]Kind{
	reflect.String: STRING_KIND,
}

//判断是否为布尔类型
var boolMap = map[reflect.Kind]Kind{
	reflect.Bool: BOOL_KIND,
}

//判断是否为数字
var numberMap = map[reflect.Kind]Kind{
	reflect.Int:     INTEGER_KIND,
	reflect.Int8:    INTEGER_KIND,
	reflect.Int16:   INTEGER_KIND,
	reflect.Int32:   INTEGER_KIND,
	reflect.Int64:   INTEGER_KIND,
	reflect.Uint:    INTEGER_KIND,
	reflect.Uint8:   INTEGER_KIND,
	reflect.Uint16:  INTEGER_KIND,
	reflect.Uint32:  INTEGER_KIND,
	reflect.Uint64:  INTEGER_KIND,
	reflect.Float32: FLOAT_KIND,
	reflect.Float64: FLOAT_KIND,
}

func checkString(v interface{}) (ok bool) {
	var typeKind reflect.Kind
	if typeKind, ok = v.(reflect.Kind); !ok {
		typeKind = reflect.TypeOf(v).Kind()
	}
	_, ok = stringMap[typeKind]
	return
}

func checkBool(v interface{}) (ok bool) {
	var typeKind reflect.Kind
	if typeKind, ok = v.(reflect.Kind); !ok {
		typeKind = reflect.TypeOf(v).Kind()
	}
	_, ok = boolMap[typeKind]
	return
}

//val is kind or val
func checkNumber(v interface{}, args ...interface{}) (ok bool) {
	var t Kind = ALL_KIND
	var typeKind reflect.Kind
	if typeKind, ok = v.(reflect.Kind); !ok {
		typeKind = reflect.TypeOf(v).Kind()
	}
	if len(args) > 0 {
		if t, ok = args[0].(Kind); !ok {
			return
		}
	}
	kind, ok := numberMap[typeKind]
	if !ok || t == ALL_KIND {
		return
	}
	ok = kind == t
	return
}

//val is kind or val
func checkArray(v interface{}, args ...interface{}) (ok bool) {
	var t Kind = ALL_KIND
	var typeKind reflect.Kind
	if typeKind, ok = v.(reflect.Kind); !ok {
		typeKind = reflect.TypeOf(v).Kind()
	}
	if len(args) > 0 {
		if t, ok = args[0].(Kind); !ok {
			return
		}
	}
	kind, ok := arrayMap[typeKind]
	if !ok || t == ALL_KIND {
		return
	}
	ok = kind == t
	return
}

//检查 array、map、slice 中的值是否含有 array、map、slice、struct
func checkArrayValueIsMulti(value reflect.Value) (ok bool, fieldNum int) {
	kind := value.Type().Kind()

	//检查类型是否是 array、map、slice
	ok = checkArray(kind)
	if !ok {
		return
	}
	//检查值的类型是不是 map、array、map 或 struct
	valueKind := value.Type().Elem().Kind()

	ok = checkArray(valueKind)
	if !ok && valueKind != reflect.Struct {
		return
	}
	fieldNum = value.Len()
	//检查数组长度是否大于 0
	if fieldNum > 0 {
		ok = true
	}
	return
}

func InArray(val interface{}, listArr interface{}) (re bool) {
	lv := reflect.ValueOf(listArr)
	l := lv.Len()
	for i := 0; i < l; i++ {
		if reflect.DeepEqual(val, lv.Index(i).Interface()) {
			re = true
			break
		}
	}
	return
}

func isZeroValue(val reflect.Value) bool {
	typeKind := val.Kind()
	switch typeKind {
	case reflect.String, reflect.Array:
		return val.Len() == 0
	case reflect.Map, reflect.Slice:
		return val.Len() == 0 || val.IsNil()
	case reflect.Bool:
		return !val.Bool()
	case reflect.Int, reflect.Int8, reflect.Int16, reflect.Int32, reflect.Int64:
		return val.Int() == 0
	case reflect.Uint, reflect.Uint8, reflect.Uint16, reflect.Uint32, reflect.Uint64, reflect.Uintptr:
		return val.Uint() == 0
	case reflect.Float32, reflect.Float64:
		return val.Float() == 0
	case reflect.Interface, reflect.Ptr:
		return val.IsNil()
	}
	return reflect.DeepEqual(val.Interface(), reflect.Zero(val.Type()).Interface())
}

func formatError(format string, eParamsMap map[string]string) error {
	var params []string
	for k, v := range eParamsMap {
		params = append(params, "["+k+"]", v)
	}
	replacer := strings.NewReplacer(params...)
	return errors.New(replacer.Replace(format))
}

func parseStr(val string, kind reflect.Kind) (re interface{}, err error) {
	switch kind {
	case reflect.Int, reflect.Int8, reflect.Int16, reflect.Int32, reflect.Int64:
		re, err = strconv.ParseInt(val, 10, 64)
	case reflect.Uint, reflect.Uint8, reflect.Uint16, reflect.Uint32, reflect.Uint64:
		re, err = strconv.ParseUint(val, 10, 64)
	case reflect.Bool:
		re, err = strconv.ParseBool(val)
	case reflect.Float32, reflect.Float64:
		re, err = strconv.ParseFloat(val, 64)
	case reflect.String:
		re = val
	}
	return
}

func parseReflectV(val reflect.Value, kind reflect.Kind) (re interface{}) {
	switch kind {
	case reflect.Int, reflect.Int8, reflect.Int16, reflect.Int32, reflect.Int64:
		re = val.Int()
	case reflect.Uint, reflect.Uint8, reflect.Uint16, reflect.Uint32, reflect.Uint64:
		re = val.Uint()
	case reflect.Bool:
		re = val.Bool()
	case reflect.Float32, reflect.Float64:
		re = val.Float()
	case reflect.String:
		re = val.String()
	}
	return
}
