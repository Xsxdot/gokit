package utils

import (
	"testing"
)

// 测试用的结构体
type TestModel struct {
	Name     string `json:"name" validate:"required,min=2,max=50" comment:"名称"`
	Email    string `json:"email" validate:"required,email" comment:"邮箱"`
	Age      int    `json:"age" validate:"min=1,max=150" comment:"年龄"`
	Page     int    `json:"page" validate:"omitempty,min=1" comment:"页码"`
	Limit    int    `json:"limit" validate:"omitempty,min=1,max=100" comment:"每页数量"`
	Keyword  string `json:"keyword" validate:"omitempty,max=100" comment:"关键词"`
	Optional string `json:"optional" validate:"omitempty,max=50" comment:"可选字段"`
}

func TestValidate_Success(t *testing.T) {
	model := TestModel{
		Name:  "测试名称",
		Email: "test@example.com",
		Age:   25,
		Page:  1,
		Limit: 10,
	}

	valid, msg := IsValid(model)
	if !valid {
		t.Errorf("期望验证成功，但失败: %s", msg)
	}
}

func TestValidate_Required(t *testing.T) {
	model := TestModel{
		Email: "test@example.com",
		Age:   25,
	}

	valid, msg := IsValid(model)
	if valid {
		t.Error("期望验证失败，但成功了")
	}
	if !contains(msg, "名称不能为空") {
		t.Errorf("期望包含 '名称不能为空'，实际: %s", msg)
	}
}

func TestValidate_StringMin(t *testing.T) {
	model := TestModel{
		Name:  "a", // 只有一个字符，小于 min=2
		Email: "test@example.com",
		Age:   25,
	}

	valid, msg := IsValid(model)
	if valid {
		t.Error("期望验证失败，但成功了")
	}
	if !contains(msg, "名称不能小于2个字符") {
		t.Errorf("期望包含 '名称不能小于2个字符'，实际: %s", msg)
	}
}

func TestValidate_StringMax(t *testing.T) {
	// 注意：validator 的 max 验证的是字符数(runes)，不是字节数
	// 中文字符占3字节但算作1个字符，所以需要超过100个字符才能触发验证
	longKeyword := ""
	for i := 0; i < 110; i++ {
		longKeyword += "字"
	}

	model := TestModel{
		Name:    "测试名称",
		Email:   "test@example.com",
		Age:     25,
		Keyword: longKeyword, // 110个字符，超过 max=100
	}

	valid, msg := IsValid(model)
	if valid {
		t.Error("期望验证失败，但成功了")
	}
	if !contains(msg, "关键词不能超过100个字符") {
		t.Errorf("期望包含 '关键词不能超过100个字符'，实际: %s", msg)
	}
}

func TestValidate_IntMax(t *testing.T) {
	model := TestModel{
		Name:  "测试名称",
		Email: "test@example.com",
		Age:   25,
		Limit: 200, // 超过 max=100
	}

	valid, msg := IsValid(model)
	if valid {
		t.Error("期望验证失败，但成功了")
	}
	if !contains(msg, "每页数量不能超过100") {
		t.Errorf("期望包含 '每页数量不能超过100'，实际: %s", msg)
	}
}

func TestValidate_IntMin(t *testing.T) {
	model := TestModel{
		Name:  "测试名称",
		Email: "test@example.com",
		Age:   0, // 小于 min=1
	}

	valid, msg := IsValid(model)
	if valid {
		t.Error("期望验证失败，但成功了")
	}
	if !contains(msg, "年龄不能小于1") {
		t.Errorf("期望包含 '年龄不能小于1'，实际: %s", msg)
	}
}

func TestValidate_Email(t *testing.T) {
	model := TestModel{
		Name:  "测试名称",
		Email: "invalid-email",
		Age:   25,
	}

	valid, msg := IsValid(model)
	if valid {
		t.Error("期望验证失败，但成功了")
	}
	if !contains(msg, "邮箱必须是有效的电子邮件地址") {
		t.Errorf("期望包含 '邮箱必须是有效的电子邮件地址'，实际: %s", msg)
	}
}

// optional 为 gokit 注册的占位校验，与 omitempty 不同：不跳过后续规则（如 optional,max=3）
func TestValidate_OptionalTag(t *testing.T) {
	type M struct {
		X string `json:"x" validate:"optional" comment:"可空"`
		Y string `json:"y" validate:"optional,max=3" comment:"可空有上限"`
		P *string `json:"p" validate:"optional" comment:"可空指针"`
	}
	if valid, msg := IsValid(M{}); !valid {
		t.Errorf("empty optional fields: %s", msg)
	}
	if valid, msg := IsValid(M{P: nil}); !valid {
		t.Errorf("nil pointer optional: %s", msg)
	}
	if valid, msg := IsValid(M{X: "a", Y: "ab"}); !valid {
		t.Errorf("within max: %s", msg)
	}
	if valid, _ := IsValid(M{X: "a", Y: "abcd"}); valid {
		t.Error("expect Y over max to fail")
	}
}

func TestValidate_MultipleErrors(t *testing.T) {
	model := TestModel{
		// Name 缺失 (required)
		Email: "invalid", // 格式错误
		Age:   0,         // 小于 min=1
	}

	valid, msg := IsValid(model)
	if valid {
		t.Error("期望验证失败，但成功了")
	}

	// 应包含多个错误，用分号分隔
	if !contains(msg, "名称") || !contains(msg, "邮箱") || !contains(msg, "年龄") {
		t.Errorf("期望包含多个错误信息，实际: %s", msg)
	}
}

// 辅助函数：检查字符串是否包含子串
func contains(s, substr string) bool {
	return len(s) >= len(substr) && (s == substr || len(s) > 0 && containsHelper(s, substr))
}

func containsHelper(s, substr string) bool {
	for i := 0; i <= len(s)-len(substr); i++ {
		if s[i:i+len(substr)] == substr {
			return true
		}
	}
	return false
}