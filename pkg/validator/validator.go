// Package validator 参数校验工具，配合 binding 标签使用
package validator

import (
	"fmt"
	"reflect"
	"strings"

	"github.com/gin-gonic/gin/binding"
	"github.com/go-playground/validator/v10"
)

var defaultValidator *validator.Validate

func init() {
	if v, ok := binding.Validator.Engine().(*validator.Validate); ok {
		defaultValidator = v
		// 可在此注册自定义校验规则
		_ = defaultValidator.RegisterValidation("notblank", notBlank)
	}
}

func notBlank(fl validator.FieldLevel) bool {
	f := fl.Field()
	switch f.Kind() {
	case reflect.String:
		return strings.TrimSpace(f.String()) != ""
	default:
		return true
	}
}

// Validate 校验结构体，返回用户友好的错误信息
func Validate(s interface{}) error {
	if defaultValidator == nil {
		return fmt.Errorf("validator not initialized")
	}
	err := defaultValidator.Struct(s)
	if err == nil {
		return nil
	}
	if errs, ok := err.(validator.ValidationErrors); ok {
		msgs := make([]string, 0, len(errs))
		for _, e := range errs {
			msgs = append(msgs, fieldErrorToMessage(e))
		}
		return fmt.Errorf("%s", strings.Join(msgs, "; "))
	}
	return err
}

func fieldErrorToMessage(e validator.FieldError) string {
	field := e.Field()
	tag := e.Tag()
	switch tag {
	case "required":
		return field + " 不能为空"
	case "email":
		return "邮箱格式错误"
	case "gt":
		return field + " 必须大于 " + e.Param()
	case "gte":
		return field + " 必须大于等于 " + e.Param()
	case "lt":
		return field + " 必须小于 " + e.Param()
	case "lte":
		return field + " 必须小于等于 " + e.Param()
	case "min":
		return field + " 长度或数值不能小于 " + e.Param()
	case "max":
		return field + " 长度或数值不能大于 " + e.Param()
	case "notblank":
		return field + " 不能为空白"
	default:
		return field + " 校验失败"
	}
}
