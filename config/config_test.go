package config

import (
	"os"
	"path/filepath"
	"testing"
)

func TestLoad_ValidConfig(t *testing.T) {
	content := `{
		"log_types": [
			{
				"name": "CMD",
				"file_pattern": "CMD_*.txt",
				"match_keys": [0, 1]
			}
		]
	}`

	tmpDir := t.TempDir()
	configPath := filepath.Join(tmpDir, "config.json")
	os.WriteFile(configPath, []byte(content), 0644)

	cfg, err := Load(configPath)
	if err != nil {
		t.Fatalf("Load() error: %v", err)
	}

	if len(cfg.LogTypes) != 1 {
		t.Errorf("LogTypes count = %d, want 1", len(cfg.LogTypes))
	}
}

func TestValidate_EmptyLogTypes(t *testing.T) {
	cfg := &Config{}

	err := cfg.Validate()
	if err == nil {
		t.Error("Validate() should error when log_types is empty")
	}
}

func TestValidate_MissingMatchKeys(t *testing.T) {
	cfg := &Config{
		LogTypes: []LogTypeConfig{
			{Name: "CMD", FilePattern: "CMD_*.txt"},
		},
	}

	err := cfg.Validate()
	if err == nil {
		t.Error("Validate() should error when match_keys is empty")
	}
}

func TestGetFieldName(t *testing.T) {
	ltCfg := &LogTypeConfig{
		FieldNames: map[int]string{0: "imsi", 1: "msisdn"},
	}

	if name := ltCfg.GetFieldName(0); name != "imsi" {
		t.Errorf("GetFieldName(0) = %q, want %q", name, "imsi")
	}
	if name := ltCfg.GetFieldName(99); name != "field_100" {
		t.Errorf("GetFieldName(99) = %q, want %q", name, "field_100")
	}
}

func TestGetLogTypeConfig(t *testing.T) {
	cfg := &Config{
		LogTypes: []LogTypeConfig{
			{Name: "CMD", FilePattern: "CMD_*.txt", MatchKeys: []int{0}},
			{Name: "DES", FilePattern: "DES_*.txt", MatchKeys: []int{1}},
		},
	}

	lt, err := cfg.GetLogTypeConfig("CMD")
	if err != nil {
		t.Fatalf("GetLogTypeConfig(CMD) error: %v", err)
	}
	if lt.Name != "CMD" {
		t.Errorf("Name = %q, want CMD", lt.Name)
	}

	_, err = cfg.GetLogTypeConfig("UNKNOWN")
	if err == nil {
		t.Error("GetLogTypeConfig(UNKNOWN) should error")
	}
}
