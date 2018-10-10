package govalidators

import (
	// "errors"
	"fmt"
	"reflect"
	"regexp"
	"strconv"
	"strings"
	"sync"
	"unicode/utf8"
)

//验证接口
type Validator interface {
	Validate(params map[string]interface{}, val reflect.Value, args ...string) (bool, error)
}

//验证函数
type ValidatorF func(params map[string]interface{}, val reflect.Value, args ...string) (bool, error)

func (f ValidatorF) Validate(params map[string]interface{}, val reflect.Value, args ...string) (bool, error) {
	return f(params, val, args...)
}

type Range struct {
	Min       string
	Max       string
	min       string
	max       string
	RangeEMsg map[string]string //keys: lessThan,equal,atLeast,between
}

//将structTag中的min和max解析到结构体中
func (self *Range) InitRangeNum(eParamsMap map[string]string, args ...string) error {
	self.min, self.max = self.Min, self.Max
	argsL := len(args)
	if (self.Min == "" && argsL == 0) || argsL > 2 {
		return formatError("[name] validator range error", eParamsMap)
	}
	if argsL == 1 {
		self.min = args[0]
	} else if argsL == 2 {
		self.min = args[0]
		self.max = args[1]
	}
	return nil
}

func (self *Range) CompareFloat(valNum float64, eParamsMap map[string]string, errorMap map[string]string) error {
	if self.min == "" ||
		(self.min != VALIDATOR_IGNORE_SIGN && !regexp.MustCompile(FLOAT_REG).MatchString(self.min)) ||
		(self.max != VALIDATOR_IGNORE_SIGN && !regexp.MustCompile(FLOAT_REG).MatchString(self.max) && self.max != "") {
		return formatError("[name] validator range error", eParamsMap)
	}
	if self.min == VALIDATOR_IGNORE_SIGN && (self.max == VALIDATOR_IGNORE_SIGN || self.max == "") {
		return nil
	}
	var min, max float64
	var ok bool
	var errKey, errStr string
	if self.min == VALIDATOR_IGNORE_SIGN {
		max, _ = strconv.ParseFloat(self.max, 64)
		if valNum > max {
			errKey = "lessThan"
			eParamsMap["max"] = self.max
		}
	} else if self.max == VALIDATOR_IGNORE_SIGN {
		min, _ = strconv.ParseFloat(self.min, 64)
		if valNum < min {
			errKey = "atLeast"
			eParamsMap["min"] = self.min
		}
	} else if self.max == "" {
		min, _ = strconv.ParseFloat(self.min, 64)
		if valNum != min {
			errKey = "equal"
			eParamsMap["min"] = self.min
		}
	} else {
		max, _ = strconv.ParseFloat(self.max, 64)
		min, _ = strconv.ParseFloat(self.min, 64)
		if valNum < min || valNum > max {
			errKey = "between"
			eParamsMap["min"] = self.min
			eParamsMap["max"] = self.max
		}
	}
	if errKey == "" {
		return nil
	}
	if errStr, ok = self.RangeEMsg[errKey]; !ok {
		errStr = errorMap[errKey]
	}
	return formatError(errStr, eParamsMap)
}

func (self *Range) CompareInteger(valNum int64, eParamsMap map[string]string, errorMap map[string]string) error {
	if self.min == "" ||
		(self.min != VALIDATOR_IGNORE_SIGN && !regexp.MustCompile(INTEGER_REG).MatchString(self.min)) ||
		(self.max != VALIDATOR_IGNORE_SIGN && !regexp.MustCompile(INTEGER_REG).MatchString(self.max) && self.max != "") {
		return formatError("[name] validator range error", eParamsMap)
	}
	if self.min == VALIDATOR_IGNORE_SIGN && (self.max == VALIDATOR_IGNORE_SIGN || self.max == "") {
		return nil
	}

	var min, max int64
	var ok bool
	var errKey, errStr string
	if self.min == VALIDATOR_IGNORE_SIGN {
		max, _ = strconv.ParseInt(self.max, 10, 64)
		if valNum > max {
			errKey = "lessThan"
			eParamsMap["max"] = self.max
		}
	} else if self.max == VALIDATOR_IGNORE_SIGN {
		min, _ = strconv.ParseInt(self.min, 10, 64)
		if valNum < min {
			errKey = "atLeast"
			eParamsMap["min"] = self.min
		}
	} else if self.max == "" {
		min, _ = strconv.ParseInt(self.min, 10, 64)
		if valNum != min {
			errKey = "equal"
			eParamsMap["min"] = self.min
		}
	} else {
		max, _ = strconv.ParseInt(self.max, 10, 64)
		min, _ = strconv.ParseInt(self.min, 10, 64)
		if min >= max {
			return formatError("[name] validator range error", eParamsMap)
		}
		if valNum < min || valNum > max {
			errKey = "between"
			eParamsMap["min"] = self.min
			eParamsMap["max"] = self.max
		}
	}
	if errKey == "" {
		return nil
	}
	if errStr, ok = self.RangeEMsg[errKey]; !ok {
		errStr = errorMap[errKey]
	}
	return formatError(errStr, eParamsMap)
}

type RequiredValidator struct {
	EMsg string
}

func (self *RequiredValidator) Validate(params map[string]interface{}, val reflect.Value, args ...string) (bool, error) {
	eMsg := "[name] is must required"
	eParamsMap := map[string]string{
		"name": params["name"].(string),
	}
	if self.EMsg != "" {
		eMsg = self.EMsg
	}

	if isZeroValue(val) {
		return false, formatError(eMsg, eParamsMap)
	}
	return true, nil
}

/**
 * 当只有 Min 或者 Max 的值，另一个值为 nil 时，验证器为等于有值的对应值
 * 当只有 Min 或者 Max 的值，另一个值为 _ 时，验证器为忽略带 _ 的值
 * 栗子
 * string=1,2 表示 Min=1,Max=2,就是说 1 <= len(str) <= 2
 * string=1 表示 Min=1,Max=nil,就是说 len(str) = 1
 * string=1,_ 表示 Min=1,Max=_,就是说 1 <= len(str)
 */
type StringValidator struct {
	EMsg string
	Range
}

func (self *StringValidator) Validate(params map[string]interface{}, val reflect.Value, args ...string) (bool, error) {
	eMsg := "[name] is not a string"
	eParamsMap := map[string]string{
		"name": params["name"].(string),
	}
	if self.EMsg != "" {
		eMsg = self.EMsg
	}

	if !checkString(val.Kind()) {
		return false, formatError(eMsg, eParamsMap)
	}
	//后边不接参数，表示只判断类型
	if len(args) == 0 {
		return true, nil
	}
	err := self.InitRangeNum(eParamsMap, args...)
	if err != nil {
		return false, err
	}
	strNum := utf8.RuneCountInString(val.String())
	err = self.CompareInteger(int64(strNum), eParamsMap, stringErrorMap)
	if err != nil {
		return false, err
	}
	return true, nil
}

/**
 * 当只有 Min 或者 Max 的值，另一个值为 nil 时，验证器为等于有值的对应值
 * 当只有 Min 或者 Max 的值，另一个值为 _ 时，验证器为忽略带 _ 的值
 * 栗子
 * integer=1,2 表示 Min=1,Max=2,就是说 1 <= num <= 2
 * integer=1 表示 Min=1,Max=nil,就是说 num = 1
 * integer=1,_ 表示 Min=1,Max=_,就是说 1 <= num
 */
type IntegerValidator struct {
	EMsg string
	Range
}

func (self *IntegerValidator) Validate(params map[string]interface{}, val reflect.Value, args ...string) (bool, error) {
	eMsg := "[name] is not a integer"
	eParamsMap := map[string]string{
		"name": params["name"].(string),
	}
	if self.EMsg != "" {
		eMsg = self.EMsg
	}
	if !checkNumber(val.Kind(), INTEGER_KIND) {
		return false, formatError(eMsg, eParamsMap)
	}
	//后边不接参数，表示只判断类型
	if len(args) == 0 {
		return true, nil
	}
	err := self.InitRangeNum(eParamsMap, args...)
	if err != nil {
		return false, err
	}
	err = self.CompareInteger(val.Int(), eParamsMap, numberErrorMap)
	if err != nil {
		return false, err
	}
	return true, nil
}

/**
 * 当只有 Min 或者 Max 的值，另一个值为 nil 时，验证器为等于有值的对应值
 * 当只有 Min 或者 Max 的值，另一个值为 _ 时，验证器为忽略带 _ 的值
 * 栗子
 * string=1,2 表示 Min=1,Max=2,就是说 1 <= len(array) <= 2
 * string=1 表示 Min=1,Max=nil,就是说 len(array) = 1
 * string=1,_ 表示 Min=1,Max=_,就是说 1 <= len(array)
 */
type ArrayValidator struct {
	EMsg string
	Range
}

func (self *ArrayValidator) Validate(params map[string]interface{}, val reflect.Value, args ...string) (bool, error) {
	eMsg := "[name] is not a array/map/slice"
	eParamsMap := map[string]string{
		"name": params["name"].(string),
	}
	if self.EMsg != "" {
		eMsg = self.EMsg
	}

	if !checkArray(val.Kind()) {
		return false, formatError(eMsg, eParamsMap)
	}

	//后边不接参数，表示只判断类型
	if len(args) == 0 {
		return true, nil
	}
	err := self.InitRangeNum(eParamsMap, args...)
	err = self.CompareInteger(int64(val.Len()), eParamsMap, arrayErrorMap)
	if err != nil {
		return false, err
	}
	return true, nil
}

/**
 * 仅支持 string、float、int、bool 类型
 * 或值类型为 string、float、int、bool 类型的array、slice、map
 */
type InValidator struct {
	EMsg     string
	TypeEMsg string
}

func (self *InValidator) Validate(params map[string]interface{}, val reflect.Value, args ...string) (bool, error) {
	eMsg := "[name] is not in params [args]"
	typeEMsg := "[name] type invalid"
	eParamsMap := map[string]string{
		"name": params["name"].(string),
		"args": fmt.Sprintf("%v", args),
	}
	if self.EMsg != "" {
		eMsg = self.EMsg
	}
	if self.TypeEMsg != "" {
		typeEMsg = self.TypeEMsg
	}
	var valsI []reflect.Value
	var argsI []interface{}
	kind := val.Kind()
	switch kind {
	case reflect.Slice, reflect.Array:
		kind = val.Type().Elem().Kind()
		arrLen := val.Len()
		for i := 0; i < arrLen; i++ {
			valsI = append(valsI, val.Index(i))
		}
	case reflect.Map:
		kind = val.Type().Elem().Kind()
		keys := val.MapKeys()
		for _, key := range keys {
			valsI = append(valsI, val.MapIndex(key))
		}
	default:
		valsI = append(valsI, val)
	}
	if !checkBool(kind) && !checkNumber(kind) && !checkString(kind) {
		return false, formatError(typeEMsg, eParamsMap)
	}
	if len(valsI) == 0 {
		return false, formatError(eMsg, eParamsMap)
	}
	//根据 val 类型将 args 转为对应格式
	for _, arg := range args {
		tmpArg, err := parseStr(arg, kind)
		if err != nil {
			return false, formatError(eMsg, eParamsMap)
		}
		argsI = append(argsI, tmpArg)
	}
	for _, valI := range valsI {
		if !InArray(parseReflectV(valI, kind), argsI) {
			return false, formatError(eMsg, eParamsMap)
		}
	}
	return true, nil
}

type EmailValidator struct {
	EMsg string
	Reg  string
}

func (self *EmailValidator) Validate(params map[string]interface{}, val reflect.Value, args ...string) (bool, error) {
	eMsg := "[name] is not a email address"
	eParamsMap := map[string]string{
		"name": params["name"].(string),
	}
	if self.EMsg != "" {
		eMsg = self.EMsg
	}
	if !checkString(val.Kind()) {
		return false, formatError(eMsg, eParamsMap)
	}
	reg := MAIL_REG
	if self.Reg != "" {
		reg = self.Reg
	}
	if !regexp.MustCompile(reg).MatchString(val.String()) {
		return false, formatError(eMsg, eParamsMap)
	}
	return true, nil
}

type UrlValidator struct {
	EMsg string
	Reg  string
}

func (self *UrlValidator) Validate(params map[string]interface{}, val reflect.Value, args ...string) (bool, error) {
	eMsg := "[name] is not a url"
	eParamsMap := map[string]string{
		"name": params["name"].(string),
	}
	if self.EMsg != "" {
		eMsg = self.EMsg
	}
	if !checkString(val.Kind()) {
		return false, formatError(eMsg, eParamsMap)
	}
	reg := URL_REG
	if self.Reg != "" {
		reg = self.Reg
	}
	if !regexp.MustCompile(reg).MatchString(val.String()) {
		return false, formatError(eMsg, eParamsMap)
	}
	return true, nil
}

type DateTimeValidator struct {
	EMsg   string
	FmtStr string
}

func (self *DateTimeValidator) Validate(params map[string]interface{}, val reflect.Value, args ...string) (bool, error) {
	eMsg := "[name] is not a date time"
	fmtStr := "Y-m-d H:i:s"
	eParamsMap := map[string]string{
		"name": params["name"].(string),
	}
	if self.EMsg != "" {
		eMsg = self.EMsg
	}
	if self.FmtStr != "" {
		fmtStr = self.FmtStr
	}
	if len(args) != 0 {
		fmtStr = args[0]
	}
	if !checkString(val.Kind()) {
		return false, formatError(eMsg, eParamsMap)
	}
	//拼接
	replaceArr := []string{
		"Y", YEAR_REG, "m", MONTH_REF, "d", DAY_REF, "H", HOUR_REF, "i", MINUTE_REF, "s", SECOND_REF,
	}
	replacer := strings.NewReplacer(replaceArr...)
	reg := `^` + replacer.Replace(fmtStr) + `$`
	if !regexp.MustCompile(reg).MatchString(val.String()) {
		return false, formatError(eMsg, eParamsMap)
	}
	return true, nil
}

/**
 * 仅支持 string、float、int、bool 类型
 * 或值类型为 string、float、int、bool 类型的array、slice、map
 */
type UniqueValidator struct {
	EMsg string
}

func (self *UniqueValidator) Validate(params map[string]interface{}, val reflect.Value, args ...string) (bool, error) {
	eMsg := "[name] is not unique"
	typeEMsg := "[name] type invalid"
	eParamsMap := map[string]string{
		"name": params["name"].(string),
	}
	if self.EMsg != "" {
		eMsg = self.EMsg
	}
	allKey := params["allKey"].(string)
	syncMap := params["syncMap"].(*sync.Map)

	kind := val.Kind()
	switch kind {
	case reflect.Slice, reflect.Array:
		kind = val.Type().Elem().Kind()
		if !checkBool(kind) && !checkNumber(kind) && !checkString(kind) {
			return false, formatError(typeEMsg, eParamsMap)
		}
		arrLen := val.Len()
		for i := 0; i < arrLen; i++ {
			tmpV := val.Index(i)
			tmpK := fmt.Sprintf("%v_%v", allKey, tmpV)
			//fmt.Println("------->", tmpK)
			_, ok := syncMap.Load(tmpK)
			if ok {
				return false, formatError(eMsg, eParamsMap)
			}
			syncMap.Store(tmpK, true)
		}
	default:
		if !checkBool(kind) && !checkNumber(kind) && !checkString(kind) {
			return false, formatError(typeEMsg, eParamsMap)
		}
		tmpK := fmt.Sprintf("%v_%v", allKey, val)
		_, ok := syncMap.Load(tmpK)
		//fmt.Println("=====>", syncMap)
		if ok {
			return false, formatError(eMsg, eParamsMap)
		}
		syncMap.Store(tmpK, true)
	}
	return true, nil
}
