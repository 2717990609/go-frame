// Package config 配置引擎单元测试
package config

import (
	"os"
	"testing"
)

func TestNewEngine(t *testing.T) {
	e := NewEngine()
	if e == nil || e.data == nil {
		t.Fatal("NewEngine should return non-nil engine with initialized data")
	}
}

func TestEngine_Load_and_Get(t *testing.T) {
	e := NewEngine()
	data := map[string]interface{}{
		"server": map[string]interface{}{
			"port": 8080,
			"host": "0.0.0.0",
		},
		"database": map[string]interface{}{
			"name": "testdb",
		},
	}
	if err := e.Load(data); err != nil {
		t.Fatalf("Load failed: %v", err)
	}
	if e.Get("server.port") != 8080 {
		t.Errorf("expected server.port=8080, got %v", e.Get("server.port"))
	}
	if e.Get("server.host") != "0.0.0.0" {
		t.Errorf("expected server.host=0.0.0.0, got %v", e.Get("server.host"))
	}
	if e.Get("database.name") != "testdb" {
		t.Errorf("expected database.name=testdb, got %v", e.Get("database.name"))
	}
	if e.Get("nonexistent") != nil {
		t.Errorf("expected nonexistent=nil, got %v", e.Get("nonexistent"))
	}
}

func TestEngine_GetString(t *testing.T) {
	e := NewEngine()
	e.Load(map[string]interface{}{
		"key": "value",
		"num": 42,
	})
	if v := e.GetString("key", "default"); v != "value" {
		t.Errorf("GetString(key)=%q, want value", v)
	}
	if v := e.GetString("num", "default"); v != "42" {
		t.Errorf("GetString(num)=%q, want 42", v)
	}
	if v := e.GetString("missing", "default"); v != "default" {
		t.Errorf("GetString(missing)=%q, want default", v)
	}
}

func TestEngine_GetInt(t *testing.T) {
	e := NewEngine()
	e.Load(map[string]interface{}{
		"port": 8080,
		"str":  "8081",
	})
	if v := e.GetInt("port", 0); v != 8080 {
		t.Errorf("GetInt(port)=%d, want 8080", v)
	}
	if v := e.GetInt("str", 0); v != 8081 {
		t.Errorf("GetInt(str)=%d, want 8081", v)
	}
	if v := e.GetInt("missing", 999); v != 999 {
		t.Errorf("GetInt(missing)=%d, want 999", v)
	}
}

func TestEngine_GetBool(t *testing.T) {
	e := NewEngine()
	e.Load(map[string]interface{}{
		"enabled": true,
		"str":     "false",
	})
	if v := e.GetBool("enabled", false); !v {
		t.Error("GetBool(enabled) want true")
	}
	if v := e.GetBool("str", true); v {
		t.Error("GetBool(str) want false")
	}
	if v := e.GetBool("missing", true); !v {
		t.Error("GetBool(missing) want default true")
	}
}

func TestEngine_GetStringSlice(t *testing.T) {
	e := NewEngine()
	e.Load(map[string]interface{}{
		"arr": []interface{}{"a", "b", "c"},
	})
	slice := e.GetStringSlice("arr")
	if len(slice) != 3 || slice[0] != "a" || slice[1] != "b" || slice[2] != "c" {
		t.Errorf("GetStringSlice(arr)=%v", slice)
	}
	if e.GetStringSlice("missing") != nil {
		t.Error("GetStringSlice(missing) should return nil")
	}
}

func TestEngine_expandEnvVars(t *testing.T) {
	os.Setenv("TEST_PORT", "9000")
	defer os.Unsetenv("TEST_PORT")

	e := NewEngine()
	data := map[string]interface{}{
		"port": "${TEST_PORT:8080}",
		"host": "${MISSING_VAR:localhost}",
	}
	if err := e.Load(data); err != nil {
		t.Fatalf("Load failed: %v", err)
	}
	if v := e.GetString("port", ""); v != "9000" {
		t.Errorf("env expanded port=%q, want 9000", v)
	}
	if v := e.GetString("host", ""); v != "localhost" {
		t.Errorf("default host=%q, want localhost", v)
	}
}

func TestEngine_Set(t *testing.T) {
	e := NewEngine()
	e.Load(map[string]interface{}{})
	e.Set("a.b.c", "value")
	if e.Get("a.b.c") != "value" {
		t.Errorf("Set/Get a.b.c failed: got %v", e.Get("a.b.c"))
	}
}
