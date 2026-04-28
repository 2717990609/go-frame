// Package response 响应格式兼容性测试（保持错误码体系不变）
package response

import (
	"encoding/json"
	"testing"
)

func TestSuccess(t *testing.T) {
	data := map[string]string{"status": "ok"}
	r := Success(data)
	if r.Code != CodeSuccess {
		t.Errorf("Code=%d, want %d", r.Code, CodeSuccess)
	}
	if r.Msg != "success" {
		t.Errorf("Msg=%q, want success", r.Msg)
	}
	if r.Data == nil {
		t.Error("Data should not be nil")
	}
	if r.Ts <= 0 {
		t.Error("Ts should be positive")
	}
}

func TestSuccess_NilData(t *testing.T) {
	r := Success(nil)
	// 规范：禁止 nil data，应转换为空 map
	if r.Data == nil {
		t.Error("Success(nil) should not have nil Data")
	}
}

func TestError(t *testing.T) {
	r := Error(CodeParamError, "参数错误")
	if r.Code != CodeParamError {
		t.Errorf("Code=%d, want %d", r.Code, CodeParamError)
	}
	if r.Msg != "参数错误" {
		t.Errorf("Msg=%q, want 参数错误", r.Msg)
	}
	if r.Data == nil {
		t.Error("Data should not be nil")
	}
}

func TestResponse_JSONStructure(t *testing.T) {
	r := Success(map[string]string{"key": "value"})
	b, err := json.Marshal(r)
	if err != nil {
		t.Fatal(err)
	}
	var m map[string]interface{}
	if err := json.Unmarshal(b, &m); err != nil {
		t.Fatal(err)
	}
	for _, key := range []string{"code", "msg", "data", "ts"} {
		if _, ok := m[key]; !ok {
			t.Errorf("JSON missing key %q", key)
		}
	}
}

func TestErrorCodes(t *testing.T) {
	// 兼容性：错误码常量必须保持不变
	codes := map[string]int{
		"CodeSuccess":         CodeSuccess,
		"CodeParamError":      CodeParamError,
		"CodeBusinessError":   CodeBusinessError,
		"CodePermissionError": CodePermissionError,
		"CodeServerError":     CodeServerError,
		"CodeThirdPartyError": CodeThirdPartyError,
	}
	expected := map[string]int{
		"CodeSuccess":         200,
		"CodeParamError":      4001,
		"CodeBusinessError":   4101,
		"CodePermissionError": 4201,
		"CodeServerError":     5001,
		"CodeThirdPartyError": 6001,
	}
	for name, code := range codes {
		if expected[name] != code {
			t.Errorf("%s=%d, want %d", name, code, expected[name])
		}
	}
}
