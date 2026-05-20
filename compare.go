package main

import (
	"fmt"
	"log-compare/comparator"
	"log-compare/config"
	"log-compare/matcher"
	"log-compare/model"
	"log-compare/reader"
	"log-compare/validator"
)

// Run 执行日志对比
func Run(cfg *config.Config, stdDir, logDir string) (*model.OverallResult, error) {
	overall := &model.OverallResult{
		LogTypeResults: make(map[string]*model.LogTypeResult),
	}

	for _, ltCfg := range cfg.LogTypes {
		result, err := compareLogType(stdDir, logDir, &ltCfg)
		if err != nil {
			return nil, fmt.Errorf("对比日志类型 %s 失败: %w", ltCfg.Name, err)
		}
		overall.LogTypeResults[ltCfg.Name] = result
	}

	return overall, nil
}

func compareLogType(standardDir, logDir string, ltCfg *config.LogTypeConfig) (*model.LogTypeResult, error) {
	logResult, err := reader.ReadLogs(logDir, ltCfg)
	if err != nil {
		return nil, err
	}

	stdResult, err := reader.ReadLogs(standardDir, ltCfg)
	if err != nil {
		return nil, err
	}

	stdGroups := matcher.GroupByMatchKey(toPtrSlice(stdResult.Records), ltCfg.MatchKeys)
	logGroups := matcher.GroupByMatchKey(toPtrSlice(logResult.Records), ltCfg.MatchKeys)

	common, logOnly, stdOnly := matcher.GetMatchKeyStats(logGroups, stdGroups)

	var comparisonDetails []model.ComparisonResult
	recordsWithDiff := 0

	for _, logRecord := range logResult.Records {
		matchedRecord, matchFound := matcher.FindMatchInMap(&logRecord, stdGroups, ltCfg.MatchKeys)

		var diffs []model.FieldDiff
		if matchedRecord != nil {
			diffs = comparator.CompareFields(&logRecord, matchedRecord, ltCfg)
			if len(diffs) > 0 {
				recordsWithDiff++
			}
			matcher.RemoveFromMap(stdGroups, ltCfg.MatchKeys, matchedRecord)
		}

		missing := validator.ValidateRequired(&logRecord, ltCfg)

		comparisonDetails = append(comparisonDetails, model.ComparisonResult{
			LogRecord:       &logRecord,
			MatchedRecord:   matchedRecord,
			Differences:     diffs,
			MatchKey:        matcher.BuildMatchKey(&logRecord, ltCfg.MatchKeys),
			MatchFound:      matchFound,
			RequiredMissing: missing,
		})
	}

	requiredStats := validator.ValidateAll(toPtrSlice(logResult.Records), ltCfg)

	return &model.LogTypeResult{
		LogType:           ltCfg.Name,
		TotalLogRecords:   len(logResult.Records),
		TotalStdRecords:   len(stdResult.Records),
		MatchKeyCount:     len(logGroups),
		StdMatchKeyCount:  len(stdGroups),
		CommonMatchKeys:   common,
		LogOnlyMatchKeys:  logOnly,
		StdOnlyMatchKeys:  stdOnly,
		RecordsWithDiff:   recordsWithDiff,
		RequiredStats:     requiredStats,
		ComparisonDetails: comparisonDetails,
	}, nil
}

func toPtrSlice(records []model.LogRecord) []*model.LogRecord {
	ptrs := make([]*model.LogRecord, len(records))
	for i := range records {
		ptrs[i] = &records[i]
	}
	return ptrs
}
