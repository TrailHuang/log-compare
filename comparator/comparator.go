package comparator

import (
	"log-compare/config"
	"log-compare/model"
)

// CompareFields 对比两条记录的字段差异
// filterFields 中的字段索引会被跳过
func CompareFields(logRecord, stdRecord *model.LogRecord, ltCfg *config.LogTypeConfig) []model.FieldDiff {
	filterSet := make(map[int]bool)
	for _, idx := range ltCfg.FilterFields {
		filterSet[idx] = true
	}

	maxLen := len(logRecord.Fields)
	if len(stdRecord.Fields) > maxLen {
		maxLen = len(stdRecord.Fields)
	}

	var diffs []model.FieldDiff
	for i := 0; i < maxLen; i++ {
		if filterSet[i] {
			continue
		}

		logVal := ""
		stdVal := ""
		if i < len(logRecord.Fields) {
			logVal = logRecord.Fields[i]
		}
		if i < len(stdRecord.Fields) {
			stdVal = stdRecord.Fields[i]
		}

		if logVal != stdVal {
			diffs = append(diffs, model.FieldDiff{
				Index:    i,
				Name:     ltCfg.GetFieldName(i),
				LogValue: logVal,
				StdValue: stdVal,
			})
		}
	}

	return diffs
}
