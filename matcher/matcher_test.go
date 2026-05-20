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

// --- FindMatchInMap ---

func TestFindMatchInMap_ExactMatch(t *testing.T) {
	std1 := &model.LogRecord{Fields: []string{"ip1", "port1", "v1"}}
	std2 := &model.LogRecord{Fields: []string{"ip2", "port2", "v2"}}
	groups := map[string][]*model.LogRecord{
		"ip1|port1": {std1},
		"ip2|port2": {std2},
	}

	logRecord := &model.LogRecord{Fields: []string{"ip1", "port1", "v3"}}
	matchKeys := []int{0, 1}

	matched, found := FindMatchInMap(logRecord, groups, matchKeys)
	if !found {
		t.Error("expected exact match")
	}
	if matched != std1 {
		t.Error("should return std1")
	}
}

func TestFindMatchInMap_FallbackToBest(t *testing.T) {
	std1 := &model.LogRecord{Fields: []string{"ip1", "port1", "v1"}}
	std2 := &model.LogRecord{Fields: []string{"ip2", "port2", "v2"}}
	groups := map[string][]*model.LogRecord{
		"ip1|port1": {std1},
		"ip2|port2": {std2},
	}

	// no exact match key, should fallback to best match by field diff
	logRecord := &model.LogRecord{Fields: []string{"ip1", "port1", "v1"}}
	matchKeys := []int{0, 1}

	matched, found := FindMatchInMap(logRecord, groups, matchKeys)
	if !found {
		t.Error("expected exact match")
	}
	if matched != std1 {
		t.Error("should return std1")
	}
}

func TestFindMatchInMap_FallbackBestMatch(t *testing.T) {
	std1 := &model.LogRecord{Fields: []string{"ipX", "portX", "completely", "different"}}
	groups := map[string][]*model.LogRecord{
		"ipX|portX": {std1},
	}

	logRecord := &model.LogRecord{Fields: []string{"ip1", "port1", "completely", "different"}}
	matchKeys := []int{0, 1}

	matched, found := FindMatchInMap(logRecord, groups, matchKeys)
	if found {
		t.Error("should not be exact match")
	}
	if matched != std1 {
		t.Error("should return best match")
	}
}

func TestFindMatchInMap_EmptyGroups(t *testing.T) {
	groups := map[string][]*model.LogRecord{}
	logRecord := &model.LogRecord{Fields: []string{"a", "b"}}
	matchKeys := []int{0, 1}

	matched, found := FindMatchInMap(logRecord, groups, matchKeys)
	if found {
		t.Error("expected false")
	}
	if matched != nil {
		t.Error("expected nil")
	}
}

// --- FindBestMatchInMap ---

func TestFindBestMatchInMap_FindsClosest(t *testing.T) {
	std1 := &model.LogRecord{Fields: []string{"a", "b", "c"}}
	std2 := &model.LogRecord{Fields: []string{"x", "y", "z"}}
	groups := map[string][]*model.LogRecord{
		"k1": {std1},
		"k2": {std2},
	}

	logRecord := &model.LogRecord{Fields: []string{"a", "b", "d"}}
	best := FindBestMatchInMap(logRecord, groups)
	if best != std1 {
		t.Error("should return std1 (closest match)")
	}
}

func TestFindBestMatchInMap_EmptyGroups(t *testing.T) {
	groups := map[string][]*model.LogRecord{}
	logRecord := &model.LogRecord{Fields: []string{"a"}}
	best := FindBestMatchInMap(logRecord, groups)
	if best != nil {
		t.Error("expected nil")
	}
}

func TestFindBestMatchInMap_MultipleCandidatesPerKey(t *testing.T) {
	std1 := &model.LogRecord{Fields: []string{"a", "b"}}
	std2 := &model.LogRecord{Fields: []string{"x", "y"}}
	std3 := &model.LogRecord{Fields: []string{"a", "z"}}
	groups := map[string][]*model.LogRecord{
		"k1": {std1, std2},
		"k2": {std3},
	}

	logRecord := &model.LogRecord{Fields: []string{"a", "b"}}
	best := FindBestMatchInMap(logRecord, groups)
	if best != std1 {
		t.Error("should return std1 (exact fields)")
	}
}

// --- RemoveFromMap ---

func TestRemoveFromMap_RemovesExisting(t *testing.T) {
	std1 := &model.LogRecord{Fields: []string{"a", "b"}}
	std2 := &model.LogRecord{Fields: []string{"a", "c"}}
	matchKeys := []int{0}

	groups := map[string][]*model.LogRecord{
		"a": {std1, std2},
	}

	RemoveFromMap(groups, matchKeys, std1)

	if len(groups["a"]) != 1 {
		t.Errorf("got %d records, want 1", len(groups["a"]))
	}
	if groups["a"][0] != std2 {
		t.Error("should have kept std2")
	}
}

func TestRemoveFromMap_DeletesKeyWhenEmpty(t *testing.T) {
	std1 := &model.LogRecord{Fields: []string{"a", "b"}}
	matchKeys := []int{0}

	groups := map[string][]*model.LogRecord{
		"a": {std1},
	}

	RemoveFromMap(groups, matchKeys, std1)

	if _, ok := groups["a"]; ok {
		t.Error("key should be deleted when empty")
	}
}

func TestRemoveFromMap_NonExistingRecord(t *testing.T) {
	std1 := &model.LogRecord{Fields: []string{"a", "b"}}
	std2 := &model.LogRecord{Fields: []string{"a", "c"}}
	matchKeys := []int{0}

	groups := map[string][]*model.LogRecord{
		"a": {std1},
	}

	// std2 is not in the group, should be a no-op
	RemoveFromMap(groups, matchKeys, std2)

	if len(groups["a"]) != 1 {
		t.Errorf("got %d records, want 1", len(groups["a"]))
	}
}

func TestRemoveFromMap_NonExistingKey(t *testing.T) {
	std1 := &model.LogRecord{Fields: []string{"x", "y"}}
	matchKeys := []int{0}

	groups := map[string][]*model.LogRecord{
		"a": {{Fields: []string{"a", "b"}}},
	}

	// key "x" doesn't exist, should be a no-op
	RemoveFromMap(groups, matchKeys, std1)

	if len(groups) != 1 {
		t.Errorf("got %d groups, want 1", len(groups))
	}
}

// --- FindMatchAndRemove ---

func TestFindMatchAndRemove_ExactMatch(t *testing.T) {
	std1 := &model.LogRecord{Fields: []string{"ip1", "port1", "v1"}}
	std2 := &model.LogRecord{Fields: []string{"ip2", "port2", "v2"}}
	candidates := []*model.LogRecord{std1, std2}

	logRecord := &model.LogRecord{Fields: []string{"ip1", "port1", "v3"}}
	matchKeys := []int{0, 1}

	matched, remaining, found := FindMatchAndRemove(logRecord, candidates, matchKeys)
	if !found {
		t.Error("expected exact match")
	}
	if matched != std1 {
		t.Error("should return std1")
	}
	if len(remaining) != 1 || remaining[0] != std2 {
		t.Errorf("remaining = %v", remaining)
	}
}

func TestFindMatchAndRemove_BestMatch(t *testing.T) {
	std1 := &model.LogRecord{Fields: []string{"ip1", "port1", "v1"}}
	std2 := &model.LogRecord{Fields: []string{"ipX", "portX", "vX"}}
	candidates := []*model.LogRecord{std1, std2}

	logRecord := &model.LogRecord{Fields: []string{"ip1", "port1", "v1"}}
	matchKeys := []int{0, 1}

	matched, remaining, found := FindMatchAndRemove(logRecord, candidates, matchKeys)
	if !found {
		t.Error("expected exact match")
	}
	if matched != std1 {
		t.Error("should return std1")
	}
	if len(remaining) != 1 {
		t.Errorf("remaining count = %d, want 1", len(remaining))
	}
}

func TestFindMatchAndRemove_EmptyCandidates(t *testing.T) {
	logRecord := &model.LogRecord{Fields: []string{"a", "b"}}
	matchKeys := []int{0, 1}

	matched, remaining, found := FindMatchAndRemove(logRecord, nil, matchKeys)
	if found {
		t.Error("expected false")
	}
	if matched != nil {
		t.Error("expected nil")
	}
	if len(remaining) != 0 {
		t.Errorf("remaining count = %d", len(remaining))
	}
}

func TestFindMatchAndRemove_BestMatchFallback(t *testing.T) {
	std1 := &model.LogRecord{Fields: []string{"x", "1", "close"}}
	std2 := &model.LogRecord{Fields: []string{"y", "2", "far_far_far"}}
	candidates := []*model.LogRecord{std1, std2}

	// match key is field[2], logRecord has "other" so no exact match
	// std1 is closest (1 diff), std2 is far (3 diffs)
	logRecord := &model.LogRecord{Fields: []string{"x", "1", "other"}}
	matchKeys := []int{2}

	matched, remaining, found := FindMatchAndRemove(logRecord, candidates, matchKeys)
	if found {
		t.Error("should not be exact match")
	}
	if matched != std1 {
		t.Error("should return best match std1")
	}
	if len(remaining) != 1 || remaining[0] != std2 {
		t.Errorf("remaining = %v", remaining)
	}
}

func TestFindMatchAndRemove_RemovesFromMiddle(t *testing.T) {
	std1 := &model.LogRecord{Fields: []string{"a", "1"}}
	std2 := &model.LogRecord{Fields: []string{"b", "2"}}
	std3 := &model.LogRecord{Fields: []string{"c", "3"}}
	candidates := []*model.LogRecord{std1, std2, std3}

	logRecord := &model.LogRecord{Fields: []string{"b", "2"}}
	matchKeys := []int{0, 1}

	matched, remaining, found := FindMatchAndRemove(logRecord, candidates, matchKeys)
	if !found {
		t.Error("expected exact match")
	}
	if matched != std2 {
		t.Error("should return std2")
	}
	if len(remaining) != 2 {
		t.Fatalf("remaining count = %d, want 2", len(remaining))
	}
	if remaining[0] != std1 || remaining[1] != std3 {
		t.Error("remaining should be [std1, std3]")
	}
}

// --- countDifferences ---

func TestCountDifferences_Identical(t *testing.T) {
	a := &model.LogRecord{Fields: []string{"a", "b", "c"}}
	b := &model.LogRecord{Fields: []string{"a", "b", "c"}}
	if countDifferences(a, b) != 0 {
		t.Error("expected 0 differences")
	}
}

func TestCountDifferences_AllDifferent(t *testing.T) {
	a := &model.LogRecord{Fields: []string{"a", "b"}}
	b := &model.LogRecord{Fields: []string{"x", "y"}}
	if countDifferences(a, b) != 2 {
		t.Error("expected 2 differences")
	}
}

func TestCountDifferences_DifferentLengths(t *testing.T) {
	a := &model.LogRecord{Fields: []string{"a"}}
	b := &model.LogRecord{Fields: []string{"a", "b", "c"}}
	if countDifferences(a, b) != 2 {
		t.Errorf("expected 2 differences, got %d", countDifferences(a, b))
	}
}

func TestCountDifferences_EmptyFields(t *testing.T) {
	a := &model.LogRecord{Fields: []string{}}
	b := &model.LogRecord{Fields: []string{"a"}}
	if countDifferences(a, b) != 1 {
		t.Errorf("expected 1 difference, got %d", countDifferences(a, b))
	}
}

// --- GroupByMatchKey ---

func TestGroupByMatchKey_DuplicateKeys(t *testing.T) {
	records := []*model.LogRecord{
		{Fields: []string{"a", "1"}},
		{Fields: []string{"a", "2"}},
		{Fields: []string{"a", "3"}},
	}

	matchKeys := []int{0}
	groups := GroupByMatchKey(records, matchKeys)

	if len(groups) != 1 {
		t.Errorf("got %d groups, want 1", len(groups))
	}
	if len(groups["a"]) != 3 {
		t.Errorf("group 'a' has %d records, want 3", len(groups["a"]))
	}
}

func TestGroupByMatchKey_EmptyRecords(t *testing.T) {
	groups := GroupByMatchKey(nil, []int{0})
	if len(groups) != 0 {
		t.Errorf("got %d groups, want 0", len(groups))
	}
}
