// Package config 配置引擎，支持环境变量注入和嵌套配置
package config

import (
	"fmt"
	"os"
	"regexp"
	"strconv"
	"strings"
)

// Engine 配置引擎
type Engine struct {
	data map[string]interface{}
}

// NewEngine 创建配置引擎
func NewEngine() *Engine {
	return &Engine{
		data: make(map[string]interface{}),
	}
}

// Load 从原始数据加载配置
func (e *Engine) Load(data map[string]interface{}) error {
	// 环境变量替换
	expandedData := e.expandEnvVars(data)
	if m, ok := expandedData.(map[string]interface{}); ok {
		e.data = m
	} else {
		e.data = make(map[string]interface{})
	}
	return nil
}

// Get 获取配置值
func (e *Engine) Get(key string) interface{} {
	return e.getNestedValue(key, e.data)
}

// GetString 获取字符串配置
func (e *Engine) GetString(key string, defaultValue string) string {
	val := e.Get(key)
	if val == nil {
		return defaultValue
	}
	return fmt.Sprintf("%v", val)
}

// GetInt 获取整型配置
func (e *Engine) GetInt(key string, defaultValue int) int {
	val := e.Get(key)
	if val == nil {
		return defaultValue
	}
	if i, err := strconv.Atoi(fmt.Sprintf("%v", val)); err == nil {
		return i
	}
	return defaultValue
}

// GetBool 获取布尔配置
func (e *Engine) GetBool(key string, defaultValue bool) bool {
	val := e.Get(key)
	if val == nil {
		return defaultValue
	}
	if b, err := strconv.ParseBool(fmt.Sprintf("%v", val)); err == nil {
		return b
	}
	return defaultValue
}

// GetStringSlice 获取字符串数组配置
func (e *Engine) GetStringSlice(key string) []string {
	val := e.Get(key)
	if val == nil {
		return nil
	}
	
	switch v := val.(type) {
	case []string:
		return v
	case []interface{}:
		result := make([]string, len(v))
		for i, item := range v {
			result[i] = fmt.Sprintf("%v", item)
		}
		return result
	default:
		return []string{fmt.Sprintf("%v", v)}
	}
}

// GetStruct 将配置解析到结构体
func (e *Engine) GetStruct(key string, dest interface{}) error {
	val := e.Get(key)
	if val == nil {
		return fmt.Errorf("配置项 %s 不存在", key)
	}
	
	// 这里可以集成更复杂的库如 mapstructure，暂时简化处理
	return fmt.Errorf("待实现结构体映射")
}

// getNestedValue 获取嵌套值，支持 "a.b.c" 语法
func (e *Engine) getNestedValue(key string, data map[string]interface{}) interface{} {
	keys := strings.Split(key, ".")
	current := data
	
	for i, k := range keys {
		if i == len(keys)-1 {
			return current[k]
		}
		
		if next, ok := current[k].(map[string]interface{}); ok {
			current = next
		} else {
			return nil
		}
	}
	
	return nil
}

// expandEnvVars 递归展开环境变量 ${VAR:default}
func (e *Engine) expandEnvVars(data interface{}) interface{} {
	switch v := data.(type) {
	case map[string]interface{}:
		result := make(map[string]interface{})
		for key, value := range v {
			result[key] = e.expandEnvVars(value)
		}
		return result
		
	case map[string]string:
		result := make(map[string]string)
		for key, value := range v {
			result[key] = e.expandEnvString(value)
		}
		return result
		
	case []interface{}:
		result := make([]interface{}, len(v))
		for i, item := range v {
			result[i] = e.expandEnvVars(item)
		}
		return result
		
	case string:
		return e.expandEnvString(v)
		
	default:
		return v
	}
}

// expandEnvString 展开字符串中的环境变量
var envVarPattern = regexp.MustCompile(`\$\{([^}:]+):?([^}]*)\}`)

func (e *Engine) expandEnvString(s string) string {
	return envVarPattern.ReplaceAllStringFunc(s, func(match string) string {
		submatches := envVarPattern.FindStringSubmatch(match)
		if len(submatches) != 3 {
			return match
		}
		
		varName := submatches[1]
		defaultValue := submatches[2]
		
		if envValue := os.Getenv(varName); envValue != "" {
			return envValue
		}
		
		return defaultValue
	})
}

// Set 设置配置值
func (e *Engine) Set(key string, value interface{}) {
	keys := strings.Split(key, ".")
	current := e.data
	
	for i, k := range keys {
		if i == len(keys)-1 {
			current[k] = value
			return
		}
		
		if next, ok := current[k].(map[string]interface{}); ok {
			current = next
		} else {
			current[k] = make(map[string]interface{})
			current = current[k].(map[string]interface{})
		}
	}
}