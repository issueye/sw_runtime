package test

import (
	"os"
	"strings"
	"testing"

	"sw_runtime/internal/runtime"
)

// TestViperModule 测试 viper 配置模块
func TestViperModule(t *testing.T) {
	runner := runtime.NewOrPanic()
	defer runner.Close()

	script := `
		const viper = require('viper');
		global.viperLoaded = typeof viper === 'object' && viper !== null;
	`

	err := runner.RunCode(script)
	if err != nil {
		t.Fatalf("Script execution failed: %v", err)
	}

	result := runner.GetValue("viperLoaded")
	if result == nil || !result.ToBoolean() {
		t.Error("viper module should be loaded")
	}

	t.Log("Viper module test passed")
}

// TestViperNewInstance 测试创建 viper 实例
func TestViperNewInstance(t *testing.T) {
	runner := runtime.NewOrPanic()
	defer runner.Close()

	script := `
		const viper = require('viper');
		const config = viper.new();
		// new() 返回的是配置对象，它有 get, set 等方法
		global.hasGetMethod = typeof config.get === 'function';
		global.hasSetMethod = typeof config.set === 'function';
	`

	err := runner.RunCode(script)
	if err != nil {
		t.Fatalf("Script execution failed: %v", err)
	}

	tests := []string{"hasGetMethod", "hasSetMethod"}
	for _, name := range tests {
		result := runner.GetValue(name)
		if result == nil || !result.ToBoolean() {
			t.Errorf("%s should be true", name)
		}
	}

	t.Log("Viper new instance test passed")
}

// TestViperSetAndGet 测试设置和获取值
func TestViperSetAndGet(t *testing.T) {
	runner := runtime.NewOrPanic()
	defer runner.Close()

	script := `
		const viper = require('viper');
		const config = viper.new();

		// 设置值
		config.set('name', 'test-app');
		config.set('port', 8080);
		config.set('debug', true);
		config.set('score', 95.5);

		// 获取值
		global.name = config.getString('name');
		global.port = config.getInt('port');
		global.debug = config.getBool('debug');
		global.score = config.getFloat64('score');
	`

	err := runner.RunCode(script)
	if err != nil {
		t.Fatalf("Script execution failed: %v", err)
	}

	result := runner.GetValue("name")
	if result == nil || result.String() != "test-app" {
		t.Errorf("name should be 'test-app', got '%s'", result.String())
	}

	result = runner.GetValue("port")
	if result == nil || result.ToInteger() != 8080 {
		t.Errorf("port should be 8080, got %d", result.ToInteger())
	}

	result = runner.GetValue("debug")
	if result == nil || !result.ToBoolean() {
		t.Error("debug should be true")
	}

	result = runner.GetValue("score")
	if result == nil || result.ToFloat() != 95.5 {
		t.Errorf("score should be 95.5, got %f", result.ToFloat())
	}

	t.Log("Viper set and get test passed")
}

// TestViperGetStringSlice 测试获取字符串数组
func TestViperGetStringSlice(t *testing.T) {
	runner := runtime.NewOrPanic()
	defer runner.Close()

	script := `
		const viper = require('viper');
		const config = viper.new();

		// 设置数组
		config.set('hosts', ['localhost', '127.0.0.1', '0.0.0.0']);
		config.set('ports', [80, 443, 8080]);

		// 获取数组
		global.hosts = config.getStringSlice('hosts');
		global.ports = config.getIntSlice('ports');
	`

	err := runner.RunCode(script)
	if err != nil {
		t.Fatalf("Script execution failed: %v", err)
	}

	result := runner.GetValue("hosts")
	if result == nil {
		t.Error("hosts should not be nil")
	}

	result = runner.GetValue("ports")
	if result == nil {
		t.Error("ports should not be nil")
	}

	t.Log("Viper get string slice test passed")
}

// TestViperGetStringMap 测试获取字符串映射
func TestViperGetStringMap(t *testing.T) {
	runner := runtime.NewOrPanic()
	defer runner.Close()

	script := `
		const viper = require('viper');
		const config = viper.new();

		// 设置映射
		config.set('database.host', 'localhost');
		config.set('database.port', 5432);
		config.set('database.name', 'mydb');

		// 获取映射
		global.db = config.getStringMap('database');
	`

	err := runner.RunCode(script)
	if err != nil {
		t.Fatalf("Script execution failed: %v", err)
	}

	result := runner.GetValue("db")
	if result == nil {
		t.Error("db should not be nil")
	}

	t.Log("Viper get string map test passed")
}

// TestViperSetDefault 测试设置默认值
func TestViperSetDefault(t *testing.T) {
	runner := runtime.NewOrPanic()
	defer runner.Close()

	script := `
		const viper = require('viper');
		const config = viper.new();

		// 设置默认值（不覆盖已存在的值）
		config.setDefault('timeout', 30);
		config.setDefault('retries', 3);

		// 设置一个存在的值
		config.set('timeout', 60);

		// 验证
		global.timeout = config.getInt('timeout');
		global.retries = config.getInt('retries');
	`

	err := runner.RunCode(script)
	if err != nil {
		t.Fatalf("Script execution failed: %v", err)
	}

	result := runner.GetValue("timeout")
	if result == nil || result.ToInteger() != 60 {
		t.Errorf("timeout should be 60, got %d", result.ToInteger())
	}

	result = runner.GetValue("retries")
	if result == nil || result.ToInteger() != 3 {
		t.Errorf("retries should be 3, got %d", result.ToInteger())
	}

	t.Log("Viper set default test passed")
}

// TestViperIsSet 测试检查键是否存在
func TestViperIsSet(t *testing.T) {
	runner := runtime.NewOrPanic()
	defer runner.Close()

	script := `
		const viper = require('viper');
		const config = viper.new();

		config.set('existing', 'value');

		global.existingSet = config.isSet('existing');
		global.missingSet = config.isSet('missing');
	`

	err := runner.RunCode(script)
	if err != nil {
		t.Fatalf("Script execution failed: %v", err)
	}

	result := runner.GetValue("existingSet")
	if result == nil || !result.ToBoolean() {
		t.Error("existing should be set")
	}

	result = runner.GetValue("missingSet")
	if result == nil || result.ToBoolean() {
		t.Error("missing should not be set")
	}

	t.Log("Viper isSet test passed")
}

// TestViperAllSettings 测试获取所有设置
func TestViperAllSettings(t *testing.T) {
	runner := runtime.NewOrPanic()
	defer runner.Close()

	script := `
		const viper = require('viper');
		const config = viper.new();

		config.set('name', 'test');
		config.set('version', '1.0.0');
		config.set('enabled', true);

		global.settings = config.allSettings();
	`

	err := runner.RunCode(script)
	if err != nil {
		t.Fatalf("Script execution failed: %v", err)
	}

	result := runner.GetValue("settings")
	if result == nil {
		t.Error("settings should not be nil")
	}

	t.Log("Viper allSettings test passed")
}

// TestViperKeys 测试获取所有键
func TestViperKeys(t *testing.T) {
	runner := runtime.NewOrPanic()
	defer runner.Close()

	script := `
		const viper = require('viper');
		const config = viper.new();

		config.set('a', 1);
		config.set('b', 2);
		config.set('c', 3);

		global.keys = config.keys();
	`

	err := runner.RunCode(script)
	if err != nil {
		t.Fatalf("Script execution failed: %v", err)
	}

	result := runner.GetValue("keys")
	if result == nil {
		t.Error("keys should not be nil")
	}

	t.Log("Viper keys test passed")
}

// TestViperUnmarshal 测试解码配置到对象
func TestViperUnmarshal(t *testing.T) {
	runner := runtime.NewOrPanic()
	defer runner.Close()

	script := `
		const viper = require('viper');
		const config = viper.new();

		config.set('name', 'test-app');
		config.set('port', 8080);
		config.set('debug', true);

		// 使用 allSettings 获取所有配置
		const all = config.allSettings();
		global.configName = all.name;
		global.configPort = all.port;
		global.configDebug = all.debug;
	`

	err := runner.RunCode(script)
	if err != nil {
		t.Fatalf("Script execution failed: %v", err)
	}

	result := runner.GetValue("configName")
	if result == nil || result.String() != "test-app" {
		t.Errorf("configName should be 'test-app', got '%s'", result.String())
	}

	result = runner.GetValue("configPort")
	if result == nil || result.ToInteger() != 8080 {
		t.Errorf("configPort should be 8080, got %d", result.ToInteger())
	}

	result = runner.GetValue("configDebug")
	if result == nil || !result.ToBoolean() {
		t.Error("configDebug should be true")
	}

	t.Log("Viper allSettings test passed")
}

// TestViperMultipleInstances 测试多个 viper 实例
func TestViperMultipleInstances(t *testing.T) {
	runner := runtime.NewOrPanic()
	defer runner.Close()

	script := `
		const viper = require('viper');

		const config1 = viper.new('config1');
		const config2 = viper.new('config2');

		config1.set('name', 'first');
		config2.set('name', 'second');

		global.name1 = config1.getString('name');
		global.name2 = config2.getString('name');
	`

	err := runner.RunCode(script)
	if err != nil {
		t.Fatalf("Script execution failed: %v", err)
	}

	result := runner.GetValue("name1")
	if result == nil || result.String() != "first" {
		t.Errorf("name1 should be 'first', got '%s'", result.String())
	}

	result = runner.GetValue("name2")
	if result == nil || result.String() != "second" {
		t.Errorf("name2 should be 'second', got '%s'", result.String())
	}

	t.Log("Viper multiple instances test passed")
}

// TestViperWithConfigFile 测试从配置文件读取
func TestViperWithConfigFile(t *testing.T) {
	runner := runtime.NewOrPanic()
	defer runner.Close()

	// 创建临时配置文件
	content := []byte(`
app:
  name: test-app
  port: 3000
  debug: true

database:
  host: localhost
  port: 5432
  name: testdb
`)
	tmpFile, err := os.CreateTemp("", "test-viper-*.yaml")
	if err != nil {
		t.Fatalf("Failed to create temp file: %v", err)
	}
	defer os.Remove(tmpFile.Name())

	if _, err := tmpFile.Write(content); err != nil {
		t.Fatalf("Failed to write temp file: %v", err)
	}
	tmpFile.Close()

	// 转换 Windows 路径
	configPath := strings.ReplaceAll(tmpFile.Name(), "\\", "\\\\")

	script := `
		const viper = require('viper');
		const config = viper.new();

		config.setConfigFile('` + configPath + `');
		config.readInConfig();

		global.appName = config.getString('app.name');
		global.appPort = config.getInt('app.port');
		global.appDebug = config.getBool('app.debug');
		global.dbHost = config.getString('database.host');
	`

	err = runner.RunCode(script)
	if err != nil {
		t.Fatalf("Script execution failed: %v", err)
	}

	result := runner.GetValue("appName")
	if result == nil || result.String() != "test-app" {
		t.Errorf("appName should be 'test-app', got '%s'", result.String())
	}

	result = runner.GetValue("appPort")
	if result == nil || result.ToInteger() != 3000 {
		t.Errorf("appPort should be 3000, got %d", result.ToInteger())
	}

	result = runner.GetValue("appDebug")
	if result == nil || !result.ToBoolean() {
		t.Error("appDebug should be true")
	}

	result = runner.GetValue("dbHost")
	if result == nil || result.String() != "localhost" {
		t.Errorf("dbHost should be 'localhost', got '%s'", result.String())
	}

	t.Log("Viper with config file test passed")
}

// TestViperEnvBinding 测试环境变量绑定
func TestViperEnvBinding(t *testing.T) {
	runner := runtime.NewOrPanic()
	defer runner.Close()

	// 设置环境变量
	os.Setenv("TEST_APP_PORT", "9999")
	defer os.Unsetenv("TEST_APP_PORT")

	script := `
		const viper = require('viper');
		const config = viper.new();

		config.setEnvPrefix('TEST');
		config.bindEnv('APP_PORT');

		// 此时应该能从环境变量读取
		global.envPort = config.getInt('APP_PORT');
	`

	err := runner.RunCode(script)
	if err != nil {
		t.Fatalf("Script execution failed: %v", err)
	}

	result := runner.GetValue("envPort")
	if result == nil || result.ToInteger() != 9999 {
		t.Errorf("envPort should be 9999, got %d", result.ToInteger())
	}

	t.Log("Viper env binding test passed")
}

// TestViperWithJSONConfig 测试从 JSON 配置文件读取
func TestViperWithJSONConfig(t *testing.T) {
	runner := runtime.NewOrPanic()
	defer runner.Close()

	// 创建临时 JSON 配置文件
	content := []byte(`{
  "app": {
    "name": "json-app",
    "port": 8080,
    "debug": false
  },
  "database": {
    "host": "db.example.com",
    "port": 5432
  }
}`)
	tmpFile, err := os.CreateTemp("", "test-viper-*.json")
	if err != nil {
		t.Fatalf("Failed to create temp file: %v", err)
	}
	defer os.Remove(tmpFile.Name())

	if _, err := tmpFile.Write(content); err != nil {
		t.Fatalf("Failed to write temp file: %v", err)
	}
	tmpFile.Close()

	// 转换 Windows 路径
	configPath := strings.ReplaceAll(tmpFile.Name(), "\\", "\\\\")

	script := `
		const viper = require('viper');
		const config = viper.new();

		config.setConfigFile('` + configPath + `');
		config.setConfigType('json');
		config.readInConfig();

		global.appName = config.getString('app.name');
		global.appPort = config.getInt('app.port');
		global.appDebug = config.getBool('app.debug');
		global.dbHost = config.getString('database.host');
	`

	err = runner.RunCode(script)
	if err != nil {
		t.Fatalf("Script execution failed: %v", err)
	}

	result := runner.GetValue("appName")
	if result == nil || result.String() != "json-app" {
		t.Errorf("appName should be 'json-app', got '%s'", result.String())
	}

	result = runner.GetValue("appPort")
	if result == nil || result.ToInteger() != 8080 {
		t.Errorf("appPort should be 8080, got %d", result.ToInteger())
	}

	result = runner.GetValue("appDebug")
	if result == nil || result.ToBoolean() {
		t.Error("appDebug should be false")
	}

	result = runner.GetValue("dbHost")
	if result == nil || result.String() != "db.example.com" {
		t.Errorf("dbHost should be 'db.example.com', got '%s'", result.String())
	}

	t.Log("Viper with JSON config test passed")
}

// TestViperWithTOMLConfig 测试从 TOML 配置文件读取
func TestViperWithTOMLConfig(t *testing.T) {
	runner := runtime.NewOrPanic()
	defer runner.Close()

	// 创建临时 TOML 配置文件
	content := []byte(`[app]
name = "toml-app"
port = 9000
enabled = true

[database]
host = "mysql.example.com"
port = 3306
`)
	tmpFile, err := os.CreateTemp("", "test-viper-*.toml")
	if err != nil {
		t.Fatalf("Failed to create temp file: %v", err)
	}
	defer os.Remove(tmpFile.Name())

	if _, err := tmpFile.Write(content); err != nil {
		t.Fatalf("Failed to write temp file: %v", err)
	}
	tmpFile.Close()

	// 转换 Windows 路径
	configPath := strings.ReplaceAll(tmpFile.Name(), "\\", "\\\\")

	script := `
		const viper = require('viper');
		const config = viper.new();

		config.setConfigFile('` + configPath + `');
		config.setConfigType('toml');
		config.readInConfig();

		global.appName = config.getString('app.name');
		global.appPort = config.getInt('app.port');
		global.appEnabled = config.getBool('app.enabled');
		global.dbHost = config.getString('database.host');
	`

	err = runner.RunCode(script)
	if err != nil {
		t.Fatalf("Script execution failed: %v", err)
	}

	result := runner.GetValue("appName")
	if result == nil || result.String() != "toml-app" {
		t.Errorf("appName should be 'toml-app', got '%s'", result.String())
	}

	result = runner.GetValue("appPort")
	if result == nil || result.ToInteger() != 9000 {
		t.Errorf("appPort should be 9000, got %d", result.ToInteger())
	}

	result = runner.GetValue("appEnabled")
	if result == nil || !result.ToBoolean() {
		t.Error("appEnabled should be true")
	}

	result = runner.GetValue("dbHost")
	if result == nil || result.String() != "mysql.example.com" {
		t.Errorf("dbHost should be 'mysql.example.com', got '%s'", result.String())
	}

	t.Log("Viper with TOML config test passed")
}

// BenchmarkViperSetAndGet 性能测试
func BenchmarkViperSetAndGet(b *testing.B) {
	runner := runtime.NewOrPanic()
	defer runner.Close()

	script := `
		const viper = require('viper');
		const config = viper.new();

		for (let i = 0; i < 100; i++) {
			config.set('key' + i, 'value' + i);
		}
	`

	err := runner.RunCode(script)
	if err != nil {
		b.Fatalf("Script execution failed: %v", err)
	}
}
