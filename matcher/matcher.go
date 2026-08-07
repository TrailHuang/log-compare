package matcher

import (
	"github.com/TrailHuang/log-compare/model"
	"strings"
)

// BuildMatchKey 根据 match_keys 构建记录的唯一标识
func BuildMatchKey(record *model.LogRecord, matchKeys []int) string {
	parts := make([]string, 0, len(matchKeys))
	for _, idx := range matchKeys {
		if idx < len(record.Fields) {
			parts = append(parts, record.Fields[idx])
		} else {
			parts = append(parts, "")
		}
	}
	return strings.Join(parts, "|")
}

// GroupByMatchKey 按 match_key 对记录分组，返回 map 供快速查找
func GroupByMatchKey(records []*model.LogRecord, matchKeys []int) map[string][]*model.LogRecord {
	groups := make(map[string][]*model.LogRecord)
	for _, r := range records {
		key := BuildMatchKey(r, matchKeys)
		groups[key] = append(groups[key], r)
	}
	return groups
}

// FindMatchInMap 在 map 中查找匹配记录
// 优先返回 match_key 相同的记录，若无则返回差异最小的记录
// 返回匹配的记录、是否精确匹配
func FindMatchInMap(record *model.LogRecord, groups map[string][]*model.LogRecord, matchKeys []int) (*model.LogRecord, bool) {
	if len(groups) == 0 {
		return nil, false
	}

	targetKey := BuildMatchKey(record, matchKeys)

	if candidates, ok := groups[targetKey]; ok && len(candidates) > 0 {
		return candidates[0], true
	}

	return FindBestMatchInMap(record, groups), false
}

// FindBestMatchInMap 在 map 中查找差异最小的记录（遍历所有值）
func FindBestMatchInMap(record *model.LogRecord, groups map[string][]*model.LogRecord) *model.LogRecord {
	var best *model.LogRecord
	minDiff := len(record.Fields) + 1

	for _, candidates := range groups {
		for _, c := range candidates {
			diff := countDifferences(record, c)
			if diff < minDiff {
				minDiff = diff
				best = c
			}
		}
	}

	return best
}

// RemoveFromMap 从 map 中移除指定记录
func RemoveFromMap(groups map[string][]*model.LogRecord, matchKeys []int, record *model.LogRecord) {
	key := BuildMatchKey(record, matchKeys)
	candidates, ok := groups[key]
	if !ok {
		return
	}

	for i, c := range candidates {
		if c == record {
			groups[key] = append(candidates[:i], candidates[i+1:]...)
			if len(groups[key]) == 0 {
				delete(groups, key)
			}
			return
		}
	}
}

// FindMatch 在候选集中查找匹配记录
// 优先返回 match_key 相同的记录，若无则返回差异最小的记录
func FindMatch(record *model.LogRecord, candidates []*model.LogRecord, matchKeys []int) (*model.LogRecord, bool) {
	if len(candidates) == 0 {
		return nil, false
	}

	targetKey := BuildMatchKey(record, matchKeys)

	var exactMatches []*model.LogRecord
	for _, c := range candidates {
		if BuildMatchKey(c, matchKeys) == targetKey {
			exactMatches = append(exactMatches, c)
		}
	}

	if len(exactMatches) > 0 {
		return exactMatches[0], true
	}

	return findBestMatch(record, candidates), false
}

// FindMatchAndRemove 在候选集中查找匹配记录，并从候选池中移除（一对一匹配）
// 优先返回 match_key 相同的记录，若无则返回差异最小的记录
// 返回匹配的记录和剩余的候选记录
func FindMatchAndRemove(record *model.LogRecord, candidates []*model.LogRecord, matchKeys []int) (*model.LogRecord, []*model.LogRecord, bool) {
	if len(candidates) == 0 {
		return nil, candidates, false
	}

	targetKey := BuildMatchKey(record, matchKeys)

	var exactMatchIndex = -1
	for i, c := range candidates {
		if BuildMatchKey(c, matchKeys) == targetKey {
			exactMatchIndex = i
			break
		}
	}

	if exactMatchIndex >= 0 {
		matched := candidates[exactMatchIndex]
		remaining := make([]*model.LogRecord, 0, len(candidates)-1)
		remaining = append(remaining, candidates[:exactMatchIndex]...)
		remaining = append(remaining, candidates[exactMatchIndex+1:]...)
		return matched, remaining, true
	}

	bestIndex, best := findBestMatchWithIndex(record, candidates)
	if bestIndex >= 0 {
		remaining := make([]*model.LogRecord, 0, len(candidates)-1)
		remaining = append(remaining, candidates[:bestIndex]...)
		remaining = append(remaining, candidates[bestIndex+1:]...)
		return best, remaining, false
	}

	return nil, candidates, false
}

func findBestMatchWithIndex(record *model.LogRecord, candidates []*model.LogRecord) (int, *model.LogRecord) {
	var bestIndex = -1
	var best *model.LogRecord
	minDiff := len(record.Fields) + 1

	for i, c := range candidates {
		diff := countDifferences(record, c)
		if diff < minDiff {
			minDiff = diff
			bestIndex = i
			best = c
		}
	}

	return bestIndex, best
}

func findBestMatch(record *model.LogRecord, candidates []*model.LogRecord) *model.LogRecord {
	var best *model.LogRecord
	minDiff := len(record.Fields) + 1

	for _, c := range candidates {
		diff := countDifferences(record, c)
		if diff < minDiff {
			minDiff = diff
			best = c
		}
	}

	return best
}

func countDifferences(a, b *model.LogRecord) int {
	maxLen := len(a.Fields)
	if len(b.Fields) > maxLen {
		maxLen = len(b.Fields)
	}

	count := 0
	for i := 0; i < maxLen; i++ {
		va := ""
		vb := ""
		if i < len(a.Fields) {
			va = a.Fields[i]
		}
		if i < len(b.Fields) {
			vb = b.Fields[i]
		}
		if va != vb {
			count++
		}
	}
	return count
}

// GetMatchKeyStats 获取 match_key 统计信息
func GetMatchKeyStats(logKeys, stdKeys map[string][]*model.LogRecord) (common, logOnly, stdOnly int) {
	logKeySet := make(map[string]bool)
	for k := range logKeys {
		logKeySet[k] = true
	}
	stdKeySet := make(map[string]bool)
	for k := range stdKeys {
		stdKeySet[k] = true
	}

	for k := range logKeySet {
		if stdKeySet[k] {
			common++
		} else {
			logOnly++
		}
	}
	for k := range stdKeySet {
		if !logKeySet[k] {
			stdOnly++
		}
	}
	return
}
