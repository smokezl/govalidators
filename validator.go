package govalidators

import (
	"fmt"
	"reflect"
	"strings"
	"sync"
)

const (
	STRUCT_EMPTY          = "struct %v is empty"
	VALIDATOR_VALUE_SIGN  = "="
	VALIDATOR_RANGE_SPLIT = ","
	VALIDATOR_IGNORE_SIGN = "_"

	//邮箱验证正则
	MAIL_REG = `\A[\w+\-.]+@[a-z\d\-]+(\.[a-z]+)*\.[a-z]+\z`
	//url验证正则
	URL_REG = `^(http|ftp|https):\/\/[\w\-_]+(\.[\w\-_]+)+([\w\-\.,@?^=%&:/~\+#]*[\w\-\@?^=%&/~\+#])?$`
	//是否为整数正则
	INTEGER_REG = `^(-)?[0-9]+$`
	//是否为float正则
	FLOAT_REG = `^(-)?[0-9]+(.[0-9]+)$`
	//年正则
	YEAR_REG = `(19|2[0-4])\d{2}`
	//月正则
	MONTH_REF = `(10|11|12|0[1-9])`
	//日正则
	DAY_REF = `(30|31|0[1-9]|[1-2][0-9])`
	//小时正则
	HOUR_REF = `(20|21|22|23|[0-1]\d)`
	//分钟正则
	MINUTE_REF = `([0-5]\d)`
	//秒正则
	SECOND_REF = `([0-5]\d)`
)

/****************************************************
 * range 验证错误提示 map
 ****************************************************/
var stringErrorMap = map[string]string{
	"lessThan": "[name] should be less than [max] chars long",
	"equal":    "[name] should be equal [min] chars long",
	"atLeast":  "[name] should be at least [min] chars long",
	"between":  "[name] should be betwween [min] and [max] chars long",
}

var numberErrorMap = map[string]string{
	"lessThan": "[name] should be less than [max]",
	"equal":    "[name] should be equal [min]",
	"atLeast":  "[name] should be at least [min]",
	"between":  "[name] should be betwween [min] and [max]",
}

var arrayErrorMap = map[string]string{
	"lessThan": "array [name] length should be less than [max]",
	"equal":    "array [name] length should be equal [min]",
	"atLeast":  "array [name] length should be at least [min]",
	"between":  "array [name] length should be betwween [min] and [max]",
}

/****************************************************
 * range 验证错误提示 map
 ****************************************************/

var defaultValidator = map[string]interface{}{
	"required": &RequiredValidator{},
	"string":   &StringValidator{},
	"integer":  &IntegerValidator{},
	"array":    &ArrayValidator{},
	"email":    &EmailValidator{},
	"url":      &UrlValidator{},
	"in":       &InValidator{},
	"datetime": &DateTimeValidator{},
	"unique":   &UniqueValidator{},
}

var errorMsg map[string][]string

type goValidator struct {
	tagName           string
	skipOnStructEmpty bool
	validatorSplit    string
	validator         map[string]interface{}
}

func New() *goValidator {
	return &goValidator{
		tagName:           "validate",
		skipOnStructEmpty: true,
		validatorSplit:    "||",
		validator:         defaultValidator,
	}
}

func (self *goValidator) SetTag(tag string) *goValidator {
	self.tagName = tag
	return self
}

func (self *goValidator) SetSkipOnStructEmpty(skip bool) *goValidator {
	self.skipOnStructEmpty = skip
	return self
}

func (self *goValidator) SetValidatorSplit(str string) *goValidator {
	self.validatorSplit = str
	return self
}

func (self *goValidator) SetValidator(validatorK string, validator interface{}) *goValidator {
	self.validator[validatorK] = validator
	return self
}

func (self *goValidator) SetValidators(validatorMap map[string]interface{}) *goValidator {
	for validatorK, validatorV := range validatorMap {
		self.validator[validatorK] = validatorV
	}
	return self
}

func (self *goValidator) LazyValidate(s interface{}) (err error) {
	syncMap := &sync.Map{}
	parentKey := "validate"
	errArr := self.validate(s, true, syncMap, parentKey)
	syncMap = nil
	if errArr != nil {
		err = errArr[0]
	}
	return
}

func (self *goValidator) Validate(s interface{}) (err []error) {
	syncMap := &sync.Map{}
	parentKey := "validate"
	err = self.validate(s, false, syncMap, parentKey)
	syncMap = nil
	return
}

func (self *goValidator) validate(s interface{}, lazyFlag bool, syncMap *sync.Map, parentKey string) (returnErr []error) {
	var errArr []error
	typeObj := reflect.TypeOf(s)
	typeValue := reflect.ValueOf(s)
	if typeObj.Kind() == reflect.Ptr {
		typeObj = typeObj.Elem()
		typeValue = typeValue.Elem()
	}
	switch typeObj.Kind() {
	case reflect.Slice, reflect.Array:
		//判断是否需要递归
		if ok, fieldNum := checkArrayValueIsMulti(typeValue); ok {
			for i := 0; i < fieldNum; i++ {
				tmpParentKey := fmt.Sprintf("%v_%v", parentKey, i)
				errArr = self.validate(typeValue.Index(i).Interface(), lazyFlag, syncMap, tmpParentKey)
				if len(errArr) > 0 {
					returnErr = append(returnErr, errArr...)
					if lazyFlag {
						return
					}
					continue
				}
			}
		} else {
			//不需要递归
			fmt.Println("======不递归=====>", typeValue)
		}
		break
	case reflect.Struct:
		numField := typeValue.NumField()
		if numField <= 0 {
			if self.skipOnStructEmpty {
				return
			}
			returnErr = append(returnErr, fmt.Errorf(STRUCT_EMPTY, typeObj.Name()))
			return
		}

		for i := 0; i < numField; i++ {
			fieldInfo := typeValue.Field(i)
			fieldTypeInfo := typeValue.Type().Field(i)
			fieldType := fieldInfo.Type().Kind()
			tag := fieldTypeInfo.Tag.Get(self.tagName)
			if tag != "" {
				//没有配置 required，并且 field 为 0 值的，直接跳过
				isZeroValue := isZeroValue(fieldInfo)
				if isZeroValue && !strings.Contains(tag, "required") && !self.skipOnStructEmpty {
					continue
				}
				errArr = self.validateValueFromTag(tag, lazyFlag, fieldTypeInfo, fieldInfo, syncMap, parentKey)
				if len(errArr) > 0 {
					returnErr = append(returnErr, errArr...)
					if lazyFlag {
						return
					}
					continue
				}
			}
			//判断是否需要递归
			if ok, fieldNum := checkArrayValueIsMulti(fieldInfo); ok {
				for i := 0; i < fieldNum; i++ {
					tmpParentKey := fmt.Sprintf("%v_%v", parentKey, fieldTypeInfo.Name)
					errArr = self.validate(fieldInfo.Index(i).Interface(), lazyFlag, syncMap, tmpParentKey)
					if len(errArr) > 0 {
						returnErr = append(returnErr, errArr...)
						if lazyFlag {
							return
						}
						continue
					}
				}
			}

			if fieldType == reflect.Struct {
				tmpParentKey := fmt.Sprintf("%v_%v", parentKey, fieldTypeInfo.Name)
				errArr = self.validate(fieldInfo.Interface(), lazyFlag, syncMap, tmpParentKey)
				if len(errArr) > 0 {
					returnErr = append(returnErr, errArr...)
					if lazyFlag {
						return
					}
					continue
				}
			}
		}
	}
	return
}

//根据 tag 申请验证器进行验证
func (self *goValidator) validateValueFromTag(tag string, lazyFlag bool, fieldTypeInfo reflect.StructField, fieldInfo reflect.Value, syncMap *sync.Map, parentKey string) (returnErr []error) {
	validatorT := reflect.TypeOf((*Validator)(nil)).Elem()
	ValidatorFT := reflect.TypeOf((*ValidatorF)(nil)).Elem()
	args := strings.Split(tag, self.validatorSplit)
	for _, argTmp := range args {
		var vK string = argTmp
		var vArgs []string
		//查找是否含有赋值符号
		num := strings.Index(argTmp, VALIDATOR_VALUE_SIGN)
		//等于 -1,说明不是像 required 这种不含有 = 号的，而是 array=1,2 这种的
		if num != -1 {
			vK = argTmp[0:num]
			vArgs = strings.Split(argTmp[num+1:], VALIDATOR_RANGE_SPLIT)
		}

		if _, ok := self.validator[vK]; !ok {
			returnErr = append(returnErr, fmt.Errorf("validator %v not exist", vK))
			if lazyFlag {
				return
			}
			continue
		}

		var validator Validator
		tmpValidator := self.validator[vK]
		vT := reflect.TypeOf(tmpValidator)
		if vT.ConvertibleTo(ValidatorFT) {
			tmpV, ok := tmpValidator.(func(params map[string]interface{}, val reflect.Value, args ...string) (bool, error))
			if !ok {
				returnErr = append(returnErr, fmt.Errorf("validator %v error", vK))
				if lazyFlag {
					return
				}
				continue
			}
			validator = ValidatorF(tmpV)
		} else if vT.Implements(validatorT) {
			validator = tmpValidator.(Validator)
		} else {
			returnErr = append(returnErr, fmt.Errorf("validator %v error", vK))
			if lazyFlag {
				return
			}
			continue
		}
		var params = map[string]interface{}{
			"name":    fieldTypeInfo.Name,
			"syncMap": syncMap,
			"allKey":  parentKey + "_" + fieldTypeInfo.Name,
		}
		valid, err := validator.Validate(params, fieldInfo, vArgs...)
		if valid == false {
			returnErr = append(returnErr, err)
			if lazyFlag {
				return
			}
			continue
		}
	}
	return
}
