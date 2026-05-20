package matcher

import (
	"log-compare/model"
	"testing"
)

func TestBuildMatchKey(t *testing.T) {
	record := &model.LogRecord{
		Fields: []string{"a", "b", "c", "d", "e"},
	}

	matchKeys := []int{0, 2, 4}
	key := BuildMatchKey(record, matchKeys)

	expected := "a|c|e"
	if key != expected {
		t.Errorf("BuildMatchKey() = %q, want %q", key, expected)
	}
}

func TestBuildMatchKey_IndexOutOfRange(t *testing.T) {
	record := &model.LogRecord{
		Fields: []string{"a", "b"},
	}

	matchKeys := []int{0, 5}
	key := BuildMatchKey(record, matchKeys)

	expected := "a|"
	if key != expected {
		t.Errorf("BuildMatchKey() = %q, want %q", key, expected)
	}
}

func TestFindMatch_ExactMatch(t *testing.T) {
	logRecord := &model.LogRecord{Fields: []string{"ip1", "port1", "ip2", "port2"}}
	stdRecords := []*model.LogRecord{
		{Fields: []string{"ip1", "port1", "ip2", "port2"}},
		{Fields: []string{"ip3", "port3", "ip4", "port4"}},
	}

	matchKeys := []int{0, 1, 2, 3}
	matched, found := FindMatch(logRecord, stdRecords, matchKeys)

	if !found {
		t.Error("FindMatch() should find exact match")
	}
	if matched != stdRecords[0] {
		t.Error("FindMatch() should return the exact matching record")
	}
}

func TestFindMatch_NoMatch(t *testing.T) {
	logRecord := &model.LogRecord{Fields: []string{"ip1", "port1"}}
	stdRecords := []*model.LogRecord{
		{Fields: []string{"ip3", "port3"}},
	}

	matchKeys := []int{0, 1}
	matched, found := FindMatch(logRecord, stdRecords, matchKeys)

	if found {
		t.Error("FindMatch() should not find match when keys differ")
	}
	if matched == nil {
		t.Error("FindMatch() should return best match even without exact key match")
	}
}

func TestFindMatch_EmptyCandidates(t *testing.T) {
	logRecord := &model.LogRecord{Fields: []string{"ip1", "port1"}}
	matchKeys := []int{0, 1}

	matched, found := FindMatch(logRecord, nil, matchKeys)

	if found {
		t.Error("FindMatch() with empty candidates should return false")
	}
	if matched != nil {
		t.Error("FindMatch() with empty candidates should return nil")
	}
}

func TestGroupByMatchKey(t *testing.T) {
	records := []*model.LogRecord{
		{Fields: []string{"a", "1"}},
		{Fields: []string{"a", "2"}},
		{Fields: []string{"b", "3"}},
	}

	matchKeys := []int{0}
	groups := GroupByMatchKey(records, matchKeys)

	if len(groups) != 2 {
		t.Errorf("GroupByMatchKey() = %d groups, want 2", len(groups))
	}
	if len(groups["a"]) != 2 {
		t.Errorf("Group 'a' has %d records, want 2", len(groups["a"]))
	}
	if len(groups["b"]) != 1 {
		t.Errorf("Group 'b' has %d records, want 1", len(groups["b"]))
	}
}

func TestGetMatchKeyStats(t *testing.T) {
	logKeys := map[string][]*model.LogRecord{
		"a": {{Fields: []string{"a"}}},
		"b": {{Fields: []string{"b"}}},
	}
	stdKeys := map[string][]*model.LogRecord{
		"a": {{Fields: []string{"a"}}},
		"c": {{Fields: []string{"c"}}},
	}

	common, logOnly, stdOnly := GetMatchKeyStats(logKeys, stdKeys)

	if common != 1 {
		t.Errorf("common = %d, want 1", common)
	}
	if logOnly != 1 {
		t.Errorf("logOnly = %d, want 1", logOnly)
	}
	if stdOnly != 1 {
		t.Errorf("stdOnly = %d, want 1", stdOnly)
	}
}
