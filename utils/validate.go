package utils

import (
	"reflect"
	"strings"
	"sync"

	"github.com/go-playground/locales/zh"
	ut "github.com/go-playground/universal-translator"
	"github.com/go-playground/validator/v10"
	zh_translations "github.com/go-playground/validator/v10/translations/zh"
)

var (
	validate     *validator.Validate
	translator   ut.Translator
	validateOnce sync.Once
)

// 常见中文错误信息映射（使用 {0} 字段名, {1} 参数值 占位符）
var customErrorMessages = map[string]string{
	"required": "{0}不能为空",
	"email":    "{0}必须是有效的电子邮件地址",
	"min":      "{0}不能小于{1}",
	"max":      "{0}不能超过{1}",
	"oneof":    "{0}必须是[{1}]中的一个",
	"len":      "{0}长度必须是{1}",
	"eq":       "{0}必须等于{1}",
	"ne":       "{0}不能等于{1}",
	"gt":       "{0}必须大于{1}",
	"gte":      "{0}必须大于或等于{1}",
	"lt":       "{0}必须小于{1}",
	"lte":      "{0}必须小于或等于{1}",
	"numeric":  "{0}必须是有效的数值",
	"datetime": "{0}必须是有效的日期时间格式",
	"alpha":    "{0}只能包含字母",
	"alphanum": "{0}只能包含字母和数字",
	"url":      "{0}必须是有效的URL",
	"json":     "{0}必须是有效的JSON格式",
}

// GetValidator 获取全局验证器实例
func GetValidator() (*validator.Validate, ut.Translator) {
	validateOnce.Do(func() {
		validate, translator = newValidator()
	})
	return validate, translator
}

// Validate 验证结构体并返回中文错误信息
func Validate(data interface{}) (string, error) {
	v, trans := GetValidator()
	return validateStruct(v, trans, data)
}

// ValidationError 从错误中提取第一个验证错误的中文描述
func ValidationError(err error) string {
	_, trans := GetValidator()
	return getValidationError(err, trans)
}

// IsValid 检查结构体是否有效，返回是否有效及错误信息
func IsValid(data interface{}) (bool, string) {
	errMsg, err := Validate(data)
	if err != nil {
		return false, errMsg
	}
	return true, ""
}

// newValidator 创建一个支持中文错误信息的验证器
func newValidator() (*validator.Validate, ut.Translator) {
	// 创建验证器实例
	v := validator.New()

	// optional：占位标签；须 callValidationEvenIfNull，否则 nil 指针字段会直接报校验失败且不会执行本函数
	if err := v.RegisterValidation("optional", func(_ validator.FieldLevel) bool {
		return true
	}, true); err != nil {
		panic("gokit/utils: register optional validation: " + err.Error())
	}

	// 注册函数，获取struct字段中的中文标签
	v.RegisterTagNameFunc(func(fld reflect.StructField) string {
		name := strings.SplitN(fld.Tag.Get("comment"), ",", 2)[0]
		if name == "" {
			name = strings.SplitN(fld.Tag.Get("json"), ",", 2)[0]
		}
		if name == "-" {
			return fld.Name
		}
		return name
	})

	// 创建中文翻译器
	zhTrans := zh.New()
	uni := ut.New(zhTrans, zhTrans)
	trans, _ := uni.GetTranslator("zh")

	// 注册默认的中文翻译器
	zh_translations.RegisterDefaultTranslations(v, trans)

	// 注册自定义的错误信息
	for tag, msg := range customErrorMessages {
		registerCustomTranslation(v, trans, tag, msg)
	}

	return v, trans
}

// registerCustomTranslation 注册自定义翻译
func registerCustomTranslation(v *validator.Validate, trans ut.Translator, tag string, message string) {
	v.RegisterTranslation(tag, trans, func(ut ut.Translator) error {
		return ut.Add(tag, message, true)
	}, func(ut ut.Translator, fe validator.FieldError) string {
		// 获取基础翻译
		t, _ := ut.T(fe.Tag(), fe.Field(), fe.Param())

		// 对于字符串类型的 min/max/len，添加"个字符"后缀
		if fe.Kind() == reflect.String && (fe.Tag() == "min" || fe.Tag() == "max" || fe.Tag() == "len") {
			return t + "个字符"
		}

		return t
	})
}

// validateStruct 验证结构体并返回中文错误信息
func validateStruct(v *validator.Validate, trans ut.Translator, s interface{}) (string, error) {
	err := v.Struct(s)
	if err == nil {
		return "", nil
	}

	errs := err.(validator.ValidationErrors)
	var errMessages []string
	for _, e := range errs {
		errMessages = append(errMessages, e.Translate(trans))
	}

	return strings.Join(errMessages, "; "), err
}

// getValidationError 从错误中提取第一个验证错误的中文描述
func getValidationError(err error, trans ut.Translator) string {
	if err == nil {
		return ""
	}

	validationErrors, ok := err.(validator.ValidationErrors)
	if !ok || len(validationErrors) == 0 {
		return err.Error()
	}

	return validationErrors[0].Translate(trans)
}