govalidators
===========
golang初学者，在项目开发过程中造了一个简单的验证器轮子，欢迎大大们提宝贵建议和指导

#### 安装
  go get github.com/smokezl/govalidators

#### 导入
```go
import "github.com/smokezl/govalidators"
```

#### 基本使用方式
为 struct 指定验证器
```go
package main

import "github.com/smokezl/govalidators"

type Class struct {
  Cid       int64  `validate:"required||integer=1,10000"`
  Cname     string `validate:"required||string=1,5"`
  BeginTime string `validate:"required||dateTime=H:i"`
}

type Student struct {
  Uid          int64    `validate:"required||integer=1,10000"`
  Name         string   `validate:"required||string=1,5"`
  Age          int64    `validate:"required||integer=10,30"`
  Sex          string   `validate:"required||in=male,female"`
  Email        string   `validate:"email"`
  PersonalPage string   `validate:"url"`
  Hobby        []string `validate:"array=_,2||unique||in=swimming,running,singing,drawing"`
  CreateTime   string   `validate:"dateTime"`
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

支持自定义函数，必须满足 ValidatorF 类型，ValidatorF 如下
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
支持自定义struct，必须实现 Validator 接口，Validator 如下
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
定义好验证器后，初始化验证器
```go
validator := govalidators.New()
validator.SetValidators(map[string]interface{}{
  "user" : &UserValidator{},
  "vm" : validationMethod,
})
if err := validator.Validate(student); err != nil {
  fmt.Println(err)
}
```
也可以对现有的验证器进行参数设置
```go
validator := govalidators.New()
validator.SetValidators(map[string]interface{}{
  "string": &StringValidator{
    Range: Range{
      RangeEMsg: map[string]string{
        "between": "[name] between [min] [max]",
      },
    },
  },
  "datetime": &DateTimeValidator{
    FmtStr: "Y-m-d",
  },
  "Email":&EmailValidator{
    Reg: "^[\w-]+$",
  },
})
if err := validator.Validate(student); err != nil {
  fmt.Println(err)
}
```
### 现有验证器介绍













