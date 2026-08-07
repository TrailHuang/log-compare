package validator

import (
	"github.com/TrailHuang/log-compare/config"
	"github.com/TrailHuang/log-compare/model"
)

// ValidateRequired 检查必填字段是否为空
func ValidateRequired(record *model.LogRecord, ltCfg *config.LogTypeConfig) []string {
	var missing []string
	for _, idx := range ltCfg.RequiredFields {
		if idx >= len(record.Fields) || record.Fields[idx] == "" {
			missing = append(missing, ltCfg.GetFieldName(idx))
		}
	}
	return missing
}

// ValidateAll 批量校验所有记录
func ValidateAll(records []*model.LogRecord, ltCfg *config.LogTypeConfig) model.RequiredFieldStats {
	stats := model.RequiredFieldStats{
		TotalRecords:         len(records),
		MissingFieldsSummary: make(map[string]int),
	}

	for _, r := range records {
		missing := ValidateRequired(r, ltCfg)
		if len(missing) > 0 {
			stats.RecordsWithMissing++
			for _, field := range missing {
				stats.MissingFieldsSummary[field]++
			}
		}
	}

	return stats
}
