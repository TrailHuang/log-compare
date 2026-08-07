package comparator

import (
	"github.com/TrailHuang/log-compare/config"
	"github.com/TrailHuang/log-compare/model"
	"testing"
)

func TestCompareFields_NoDifferences(t *testing.T) {
	logRecord := &model.LogRecord{Fields: []string{"a", "b", "c"}}
	stdRecord := &model.LogRecord{Fields: []string{"a", "b", "c"}}
	ltCfg := &config.LogTypeConfig{}

	diffs := CompareFields(logRecord, stdRecord, ltCfg)

	if len(diffs) != 0 {
		t.Errorf("CompareFields() = %d diffs, want 0", len(diffs))
	}
}

func TestCompareFields_WithDifferences(t *testing.T) {
	logRecord := &model.LogRecord{Fields: []string{"a", "b", "c"}}
	stdRecord := &model.LogRecord{Fields: []string{"a", "X", "c"}}
	ltCfg := &config.LogTypeConfig{
		FieldNames: map[int]string{1: "field_b"},
	}

	diffs := CompareFields(logRecord, stdRecord, ltCfg)

	if len(diffs) != 1 {
		t.Fatalf("CompareFields() = %d diffs, want 1", len(diffs))
	}
	if diffs[0].Index != 1 {
		t.Errorf("diff.Index = %d, want 1", diffs[0].Index)
	}
	if diffs[0].Name != "field_b" {
		t.Errorf("diff.Name = %q, want %q", diffs[0].Name, "field_b")
	}
}

func TestCompareFields_FilterFields(t *testing.T) {
	logRecord := &model.LogRecord{Fields: []string{"a", "b", "c", "d"}}
	stdRecord := &model.LogRecord{Fields: []string{"a", "X", "Y", "d"}}
	ltCfg := &config.LogTypeConfig{
		FilterFields: []int{1, 2},
	}

	diffs := CompareFields(logRecord, stdRecord, ltCfg)

	if len(diffs) != 0 {
		t.Errorf("CompareFields() with filter = %d diffs, want 0", len(diffs))
	}
}

func TestCompareFields_DifferentLengths(t *testing.T) {
	logRecord := &model.LogRecord{Fields: []string{"a", "b"}}
	stdRecord := &model.LogRecord{Fields: []string{"a", "b", "c"}}
	ltCfg := &config.LogTypeConfig{}

	diffs := CompareFields(logRecord, stdRecord, ltCfg)

	if len(diffs) != 1 {
		t.Fatalf("CompareFields() = %d diffs, want 1", len(diffs))
	}
	if diffs[0].Index != 2 {
		t.Errorf("diff.Index = %d, want 2", diffs[0].Index)
	}
}

func TestCompareFields_MultipleDifferences(t *testing.T) {
	logRecord := &model.LogRecord{Fields: []string{"a", "b", "c", "d"}}
	stdRecord := &model.LogRecord{Fields: []string{"a", "X", "Y", "d"}}
	ltCfg := &config.LogTypeConfig{}

	diffs := CompareFields(logRecord, stdRecord, ltCfg)

	if len(diffs) != 2 {
		t.Fatalf("CompareFields() = %d diffs, want 2", len(diffs))
	}
	if diffs[0].Index != 1 || diffs[1].Index != 2 {
		t.Errorf("diff indices = %v, want [1, 2]", []int{diffs[0].Index, diffs[1].Index})
	}
}
