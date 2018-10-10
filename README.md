govalidators
===========
golang初学者，在项目开发过程中造了一个简单的验证器轮子，欢迎大大们提宝贵建议和指导

### 安装
  go get github.com/smokezl/govalidators

### 导入
```go
import "github.com/smokezl/govalidators"
```

### 基本使用方式
为 struct 指定验证器
```go
package main

import "github.com/smokezl/govalidators"

type Class struct {
  Cid       int64  `validate:"required||integer=1,1000000"`
  Cname     string `validate:"required||string=1,5||unique"`
  BeginTime string `validate:"required||datetime=H:i"`
}

type Student struct {
  Uid          int64    `validate:"required||integer=1,1000000"`
  Name         string   `validate:"required||string=1,5"`
  Age          int64    `validate:"required||integer=10,30"`
  Sex          string   `validate:"required||in=male,female"`
  Email        string   `validate:"email"`
  PersonalPage string   `validate:"url"`
  Hobby        []string `validate:"array=_,2||unique||in=swimming,running,drawing"`
  CreateTime   string   `validate:"datetime"`
  Class        []Class  `validate:"array=1,3"`
}
```
验证
```go
validator := govalidators.New()
if err := validator.Validate(student); err != nil {
  fmt.Println(err)
}
```

### 自定义验证器

##### 1.支持自定义函数，必须是 ValidatorF 类型，ValidatorF 类型如下
```go
type ValidatorF func(params map[string]interface{}, val reflect.Value, args ...string) (bool, error)
```
自定义函数
```go
func validationMethod(params map[string]interface{}, val reflect.Value, args ...string) (bool, error){
  fmt.Println("validationMethod")
  ...
  return true, nil
}
```
##### 2.支持自定义struct，必须实现 Validator 接口，Validator 接口如下
```go
type Validator interface {
  Validate(params map[string]interface{}, val reflect.Value, args ...string) (bool, error)
}
```
自定义struct
```go
type UserValidator struct {
  EMsg string
}

func (self *UserValidator) Validate(params map[string]interface{}, val reflect.Value, args ...string) (bool, error) {
  fmt.Println("UserValidator")
  return true, nil
}
```
##### 3.定义好验证器后，初始化验证器
```go
validator := govalidators.New()
validator.SetValidators(map[string]interface{}{
  "user" : &UserValidator{},
  "vm" : validationMethod,
})
```
##### 4.在需要验证的字段中，增加自定义验证器
```go
Email        string   `validate:"email||user||vm"`
```
##### 5.验证
```go
if err := validator.Validate(student); err != nil {
  fmt.Println(err)
}
```
##### 6.也可以对现有的验证器进行参数设置
```go
validator := govalidators.New()
validator.SetValidators(map[string]interface{}{
  "string": &govalidators.StringValidator{
      Range: govalidators.Range{
        RangeEMsg: map[string]string{
          "between": "[name] 长度必须在 [min] 和 [max] 之间",
        },
      },
    },
  "datetime": &govalidators.DateTimeValidator{
    FmtStr: "Y-m-d",
  },
  "Email": &govalidators.EmailValidator{
    Reg: `^(\d)+$`,
  },
})
if err := validator.Validate(student); err != nil {
  fmt.Println(err)
}
```
### 现有验证器介绍
##### 1.涉及到判断范围(字符串长度、数组长度、数字大小)验证器的公共属性
```go
type Range struct {
  Min       string //最小值，外部可设置，支持0-9数字和 _ 符号，会将值赋值给 Range.min
  Max       string //最大值，外部可设置，支持0-9数字和 _ 符号，会将值赋值给 Range.max
  min       string //最小值，比对使用，支持0-9数字和 _ 符号，接收 Range.Min 和 struct 中传进来的值
  max       string //最大值，比对使用，支持0-9数字和 _ 符号，接收 Range.Max 和 struct 中传进来的值

  /**
   * 自定义范围判断错误 msg 格式，map 的 keys 有 lessThan,equal,atLeast,between ,根据类型的不同，msg 文案也不同，[min] 表示 Range.min, [max] 表示 Range.max
   * var stringErrorMap = map[string]string{
   *   "lessThan": "[name] should be less than [max] chars long",
   *   "equal":    "[name] should be equal [min] chars long",
   *   "atLeast":  "[name] should be at least [min] chars long",
   *   "between":  "[name] should be betwween [min] and [max] chars long",
   * }
   * var numberErrorMap = map[string]string{
   *   "lessThan": "[name] should be less than [max]",
   *   "equal":    "[name] should be equal [min]",
   *   "atLeast":  "[name] should be at least [min]",
   *   "between":  "[name] should be betwween [min] and [max]",
   * }
   * var arrayErrorMap = map[string]string{
   *   "lessThan": "array [name] length should be less than [max]",
   *   "equal":    "array [name] length should be equal [min]",
   *   "atLeast":  "array [name] length should be at least [min]",
   *   "between":  "array [name] length should be betwween [min] and [max]",
   * }
   */
  RangeEMsg map[string]string 
}
```
##### 2.required，判断属性值是否为对应类型的零值
```go
type RequiredValidator struct{
  EMsg string //自定义错误 msg 格式，默认为 [name] is must required，[name] 表示属性名，下同
}
```
##### 3.string(=_,n/=n,m,=n,=n,_)，判断属性值是否是字符串类型；如果后边接 = 参数，还会判断字符串长度是否合法
```go
type StringValidator struct{
  EMsg string //自定义错误 msg 格式，默认为 [name] is not a string
  Range       //涉及到判断范围(字符串长度、数组长度、数字大小)验证器的公共属性
}
```
##### 4.integer(=_,n/=n,m,=n,=n,_)，判断属性值是否是整数类型；如果后边接 = 参数，还会判断整数值是否合法
```go
type IntegerValidator struct{
  EMsg string //自定义错误 msg 格式，默认为 [name] is not a integer
  Range       //涉及到判断范围(字符串长度、数组长度、数字大小)验证器的公共属性
}
```
##### 5.array(=_,n/=n,m,=n,=n,_)，判断属性值是否是 map/slice/array 类型；如果后边接 = 参数，还会判断其长度是否合法
```go
type ArrayValidator struct{
  EMsg string //自定义错误 msg 格式，默认为 [name] is not a array/map/slice
  Range       //涉及到判断范围(字符串长度、数组长度、数字大小)验证器的公共属性
}
```
##### 6.email，判断属性值是否是合法 email
```go
type EmailValidator struct{
  EMsg string //自定义错误 msg 格式，默认为 [name] is not a email address
  Reg  string //自定义 email 正则
}
```
##### 7.url，判断属性值是否是合法 url
```go
type UrlValidator struct{
  EMsg string //自定义错误 msg 格式，默认为 [name] is not a url
  Reg  string //自定义 url 正则
}
```
##### 8.in=?,?,?,?...，判断属性值是否在 in 后边定义的值中，仅支持 string、float、int、bool 类型或值类型为 string、float、int、bool 类型的array、slice、map
```go
type InValidator struct{
  EMsg string       //自定义错误 msg 格式，默认为 [name] is not in params [args]
  TypeEMsg  string  //自定义类型错误 msg 格式，默认为 [name] type invalid
}
```
##### 9.datetime(=Y m d H i s)，判断属性值是否属于日期格式，可以自定义 Y m d H i s 的组合，如 Y-m-d、Y/m/d H:i:s、Y-m-d H:i:s
```go
type DateTimeValidator struct{
  EMsg string       //自定义错误 msg 格式，默认为 [name] is not a date time
  FmtStr  string  //自定义Y m d H i s 组合，默认为 Y-m-d H:i:s
}
```
##### 10.unique,判断属性值是否是唯一的，仅支持 string、float、int、bool 类型或值类型为 string、float、int、bool 类型的array、slice、map
```go
type UniqueValidator struct{
  EMsg string       //自定义错误 msg 格式，默认为 [name] is not unique
}
```

### 方法介绍
##### 1.func(goValidator)SetTag，设置 struct tag 中，验证标识，默认为 validate
```go
func (self *goValidator) SetTag(tag string) *goValidator
```

##### 2.func (goValidator)SetSkipOnStructEmpty，设置如果对应的值为空(零值)，跳过验证，默认为 true
```go
func (self *goValidator) SetSkipOnStructEmpty(skip bool) *goValidator
```

##### 3.func (goValidator) SetValidatorSplit(str string)，设置 struct tag 中，验证器分隔符，默认为 ||
```go
func (self *goValidator) SetValidatorSplit(str string) *goValidator
```

##### 4.func (goValidator) SetValidator(validatorK string, validator interface{})，设置自定义验证器，验证器必须满足 ValidatorF 类型或实现 Validator 接口
```go
func (self *goValidator) SetValidator(validatorK string, validator interface{}) *goValidator
```

##### 5.func (goValidator) SetValidators(validatorMap map[string]interface{})，批量设置自定义验证器，验证器必须满足 ValidatorF 类型或实现 Validator 接口
```go
func (self *goValidator) SetValidators(validatorMap map[string]interface{}) *goValidator 
``` 

##### 6.func (goValidator) LazyValidate(s interface{})，对 struct 进行验证，如果出现错误，不继续执行，并将错误返回
```go
func (self *goValidator) LazyValidate(s interface{}) (err error) 
```

##### 7.func (goValidator) Validate(s interface{})，对 struct 进行验证，如果出现错误，会继续执行，并将错误全部返回
```go
func (self *goValidator) Validate(s interface{}) (err []error) 
```

MIT licence.
