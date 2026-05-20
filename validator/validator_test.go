package validator

import (
	"log-compare/config"
	"log-compare/model"
	"testing"
)

func TestValidateRequired_AllPresent(t *testing.T) {
	record := &model.LogRecord{Fields: []string{"a", "b", "c"}}
	ltCfg := &config.LogTypeConfig{
		RequiredFields: []int{0, 1},
	}

	missing := ValidateRequired(record, ltCfg)

	if len(missing) != 0 {
		t.Errorf("ValidateRequired() = %v, want empty", missing)
	}
}

func TestValidateRequired_MissingFields(t *testing.T) {
	record := &model.LogRecord{Fields: []string{"a", "", "c"}}
	ltCfg := &config.LogTypeConfig{
		RequiredFields: []int{0, 1, 2},
		FieldNames:     map[int]string{1: "field_b"},
	}

	missing := ValidateRequired(record, ltCfg)

	if len(missing) != 1 {
		t.Fatalf("ValidateRequired() = %d missing, want 1", len(missing))
	}
	if missing[0] != "field_b" {
		t.Errorf("missing[0] = %q, want %q", missing[0], "field_b")
	}
}

func TestValidateRequired_IndexOutOfRange(t *testing.T) {
	record := &model.LogRecord{Fields: []string{"a"}}
	ltCfg := &config.LogTypeConfig{
		RequiredFields: []int{0, 5},
	}

	missing := ValidateRequired(record, ltCfg)

	if len(missing) != 1 {
		t.Fatalf("ValidateRequired() = %d missing, want 1", len(missing))
	}
}

func TestValidateRequired_EmptyRequiredFields(t *testing.T) {
	record := &model.LogRecord{Fields: []string{}}
	ltCfg := &config.LogTypeConfig{
		RequiredFields: []int{},
	}

	missing := ValidateRequired(record, ltCfg)

	if len(missing) != 0 {
		t.Errorf("ValidateRequired() with empty required = %v, want empty", missing)
	}
}

func TestValidateAll(t *testing.T) {
	records := []*model.LogRecord{
		{Fields: []string{"a", "b"}},
		{Fields: []string{"", "b"}},
		{Fields: []string{"a", ""}},
	}
	ltCfg := &config.LogTypeConfig{
		RequiredFields: []int{0, 1},
		FieldNames:     map[int]string{0: "imsi", 1: "msisdn"},
	}

	stats := ValidateAll(records, ltCfg)

	if stats.TotalRecords != 3 {
		t.Errorf("TotalRecords = %d, want 3", stats.TotalRecords)
	}
	if stats.RecordsWithMissing != 2 {
		t.Errorf("RecordsWithMissing = %d, want 2", stats.RecordsWithMissing)
	}
	if stats.MissingFieldsSummary["imsi"] != 1 {
		t.Errorf("imsi missing count = %d, want 1", stats.MissingFieldsSummary["imsi"])
	}
	if stats.MissingFieldsSummary["msisdn"] != 1 {
		t.Errorf("msisdn missing count = %d, want 1", stats.MissingFieldsSummary["msisdn"])
	}
}
