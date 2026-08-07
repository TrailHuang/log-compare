package config

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"strconv"
	"strings"
)

// Config 顶层配置结构
type Config struct {
	LogTypes  []LogTypeConfig `json:"log_types"`
	LogDir    string          `json:"log_dir,omitempty"`
	configDir string
}

// LogTypeConfig 单个日志类型配置
type LogTypeConfig struct {
	Name           string          `json:"name"`
	LogPath        string          `json:"log_path,omitempty"`
	Delimiter      string          `json:"delimiter,omitempty"`
	FilePattern    string          `json:"file_pattern"`
	FieldNamesRaw  json.RawMessage `json:"field_names,omitempty"`
	FieldNames     map[int]string  `json:"-"`
	MatchKeys      []int           `json:"match_keys"`
	RequiredFields []int           `json:"-"`
	FilterFields   []int           `json:"-"`
	RequiredRaw    json.RawMessage `json:"required_fields,omitempty"`
	FilterRaw      json.RawMessage `json:"filter_fields,omitempty"`
}

// Load 从文件加载配置
func Load(path string) (*Config, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return nil, fmt.Errorf("读取配置文件失败: %w", err)
	}

	var cfg Config
	if err := json.Unmarshal(data, &cfg); err != nil {
		return nil, fmt.Errorf("解析配置文件失败: %w", err)
	}

	cfg.configDir = filepath.Dir(path)

	for i := range cfg.LogTypes {
		if err := cfg.LogTypes[i].loadFieldNames(cfg.configDir); err != nil {
			return nil, fmt.Errorf("加载 %s 字段配置失败: %w", cfg.LogTypes[i].Name, err)
		}
		if err := cfg.LogTypes[i].parseRangeField("required_fields", &cfg.LogTypes[i].RequiredRaw, &cfg.LogTypes[i].RequiredFields); err != nil {
			return nil, fmt.Errorf("解析 %s required_fields 失败: %w", cfg.LogTypes[i].Name, err)
		}
		if err := cfg.LogTypes[i].parseRangeField("filter_fields", &cfg.LogTypes[i].FilterRaw, &cfg.LogTypes[i].FilterFields); err != nil {
			return nil, fmt.Errorf("解析 %s filter_fields 失败: %w", cfg.LogTypes[i].Name, err)
		}
	}

	return &cfg, nil
}

// parseRangeField 解析支持范围格式的字段
func (lt *LogTypeConfig) parseRangeField(name string, raw *json.RawMessage, result *[]int) error {
	if len(*raw) == 0 {
		*result = []int{}
		return nil
	}

	var rawItems []json.RawMessage
	if err := json.Unmarshal(*raw, &rawItems); err != nil {
		return fmt.Errorf("解析 %s 失败: %w", name, err)
	}

	var items []int
	for _, rawItem := range rawItems {
		itemStr := strings.Trim(string(rawItem), "\"")
		if strings.Contains(itemStr, "-") {
			parts := strings.SplitN(itemStr, "-", 2)
			start, err := strconv.Atoi(strings.TrimSpace(parts[0]))
			if err != nil {
				return fmt.Errorf("无效的范围起始值: %s", parts[0])
			}
			end, err := strconv.Atoi(strings.TrimSpace(parts[1]))
			if err != nil {
				return fmt.Errorf("无效的范围结束值: %s", parts[1])
			}
			if start > end {
				return fmt.Errorf("范围起始值 %d 大于结束值 %d", start, end)
			}
			for i := start; i <= end; i++ {
				items = append(items, i)
			}
		} else {
			val, err := strconv.Atoi(itemStr)
			if err != nil {
				return fmt.Errorf("无效的数值: %s", itemStr)
			}
			items = append(items, val)
		}
	}

	*result = items
	return nil
}

// loadFieldNames 加载字段名称配置（支持内联或外部文件）
func (lt *LogTypeConfig) loadFieldNames(configDir string) error {
	if len(lt.FieldNamesRaw) == 0 {
		lt.FieldNames = make(map[int]string)
		return nil
	}

	var fileName string
	if err := json.Unmarshal(lt.FieldNamesRaw, &fileName); err == nil {
		fieldPath := filepath.Join(configDir, fileName)
		data, err := os.ReadFile(fieldPath)
		if err != nil {
			lt.FieldNames = make(map[int]string)
			return nil
		}

		var rawMap map[string]string
		if err := json.Unmarshal(data, &rawMap); err != nil {
			return fmt.Errorf("解析字段配置文件失败: %w", err)
		}

		lt.FieldNames = make(map[int]string, len(rawMap))
		for k, v := range rawMap {
			if idx, err := strconv.Atoi(k); err == nil {
				lt.FieldNames[idx] = v
			}
		}
		return nil
	}

	var rawMap map[string]string
	if err := json.Unmarshal(lt.FieldNamesRaw, &rawMap); err != nil {
		return fmt.Errorf("解析字段配置失败: %w", err)
	}

	lt.FieldNames = make(map[int]string, len(rawMap))
	for k, v := range rawMap {
		if idx, err := strconv.Atoi(k); err == nil {
			lt.FieldNames[idx] = v
		}
	}
	return nil
}

// Validate 验证配置合法性
func (c *Config) Validate() error {
	if len(c.LogTypes) == 0 {
		return fmt.Errorf("log_types 不能为空")
	}

	for i, lt := range c.LogTypes {
		if lt.Name == "" {
			return fmt.Errorf("log_types[%d]: name 不能为空", i)
		}
		if lt.FilePattern == "" {
			return fmt.Errorf("log_types[%d]: file_pattern 不能为空", i)
		}
		if len(lt.MatchKeys) == 0 {
			return fmt.Errorf("log_types[%d]: match_keys 不能为空", i)
		}
	}

	return nil
}

// GetLogTypeConfig 按名称获取日志类型配置
func (c *Config) GetLogTypeConfig(name string) (*LogTypeConfig, error) {
	for _, lt := range c.LogTypes {
		if lt.Name == name {
			return &lt, nil
		}
	}
	return nil, fmt.Errorf("未找到日志类型配置: %s", name)
}

// GetFieldName 获取字段名称，若未配置则返回默认格式
func (lt *LogTypeConfig) GetFieldName(index int) string {
	if name, ok := lt.FieldNames[index]; ok {
		return name
	}
	return fmt.Sprintf("field_%d", index+1)
}

// GetEffectiveLogDir 获取日志类型的实际日志目录
// 优先使用 log_dir + log_path 拼接，其次使用 log_path，最后使用外部传入的基础目录
func (c *Config) GetEffectiveLogDir(logType *LogTypeConfig, baseLogDir string) string {
	if c.LogDir != "" && logType.LogPath != "" {
		return filepath.Join(c.LogDir, logType.LogPath)
	}
	if logType.LogPath != "" {
		return logType.LogPath
	}
	return baseLogDir
}
