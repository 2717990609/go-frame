// Package plugin 插件加载器实现
package plugin

import (
	"fmt"
	"os"
	"path/filepath"
	"plugin"
)

// loader 插件加载器实现
type loader struct {
	registry Registry
}

// NewLoader 创建插件加载器
func NewLoader(registry Registry) Loader {
	return &loader{
		registry: registry,
	}
}

// LoadFromDir 从目录加载插件
func (l *loader) LoadFromDir(dir string) error {
	return filepath.Walk(dir, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}
		
		// 只处理.so文件
		if !info.IsDir() && filepath.Ext(path) == ".so" {
			if err := l.LoadFromFile(path); err != nil {
				return fmt.Errorf("加载插件文件 %s 失败: %w", path, err)
			}
		}
		
		return nil
	})
}

// LoadFromFile 从文件加载插件
func (l *loader) LoadFromFile(file string) error {
	// 加载.so文件
	p, err := plugin.Open(file)
	if err != nil {
		return fmt.Errorf("打开插件文件失败: %w", err)
	}
	
	// 查找NewPlugin函数
	newPluginSymbol, err := p.Lookup("NewPlugin")
	if err != nil {
		return fmt.Errorf("查找NewPlugin函数失败: %w", err)
	}
	
	// 类型断言
	newPlugin, ok := newPluginSymbol.(func() Plugin)
	if !ok {
		return fmt.Errorf("NewPlugin函数签名不正确")
	}
	
	// 创建插件实例
	plug := newPlugin()
	
	// 注册插件
	return l.registry.Register(plug)
}

// LoadBuiltin 加载内置插件
func (l *loader) LoadBuiltin() error {
	// 这里可以加载所有内置插件
	// 例如：从代码中直接注册内置插件
	return nil
}