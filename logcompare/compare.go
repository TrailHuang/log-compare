// Package logcompare 提供日志对比的核心库能力。
//
// 它既可作为命令行工具使用（见根目录的 main 包），也可被其他 Go 项目
// 作为库引用。典型的库用法：
//
//	cfg, err := config.Load("conf/log_info.json")
//	if err != nil { return err }
//	overall, err := logcompare.Run(cfg, "path/to/standard", "path/to/log")
//	if err != nil { return err }
//	for name, r := range overall.LogTypeResults {
//	    // 处理每个日志类型的对比结果 r
//	    _ = name
//	}
package logcompare

import (
	"fmt"

	"github.com/TrailHuang/log-compare/comparator"
	"github.com/TrailHuang/log-compare/config"
	"github.com/TrailHuang/log-compare/matcher"
	"github.com/TrailHuang/log-compare/model"
	"github.com/TrailHuang/log-compare/reader"
	"github.com/TrailHuang/log-compare/validator"
)

// Run 对指定的标准日志目录与待对比日志目录执行完整的日志比对流程，
// 并按配置中的日志类型聚合返回总体对比结果。
//
// 参数 cfg 为已加载并通过 Validate 校验的配置；stdDir 为标准日志所在目录；
// logDir 为待对比日志所在目录。
//
// 返回值 overall 包含每个日志类型的对比明细与统计信息；err 表示流程中
// 出现的任何错误（例如读取失败、配置错误等）。
func Run(cfg *config.Config, stdDir, logDir string) (*model.OverallResult, error) {
	overall := &model.OverallResult{
		LogTypeResults: make(map[string]*model.LogTypeResult),
	}

	for _, ltCfg := range cfg.LogTypes {
		result, err := compareLogType(cfg, stdDir, logDir, &ltCfg)
		if err != nil {
			return nil, fmt.Errorf("对比日志类型 %s 失败: %w", ltCfg.Name, err)
		}
		overall.LogTypeResults[ltCfg.Name] = result
	}

	return overall, nil
}

func compareLogType(cfg *config.Config, standardDir, logDir string, ltCfg *config.LogTypeConfig) (*model.LogTypeResult, error) {
	effectiveLogDir := cfg.GetEffectiveLogDir(ltCfg, logDir)
	logResult, err := reader.ReadLogs(effectiveLogDir, ltCfg)
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
	var unmatchedLogList []model.UnmatchedRecord
	recordsWithDiff := 0
	unmatchedLogRecords := 0

	for _, logRecord := range logResult.Records {
		matchedRecord, matchFound := matcher.FindMatchInMap(&logRecord, stdGroups, ltCfg.MatchKeys)

		var diffs []model.FieldDiff
		// 仅在精确匹配（match_key 命中）时才进行字段比较并消耗标准端记录。
		// 模糊匹配回退会把不应配对的记录强行配对，产生伪差异并污染标准池。
		if matchFound && matchedRecord != nil {
			diffs = comparator.CompareFields(&logRecord, matchedRecord, ltCfg)
			if len(diffs) > 0 {
				recordsWithDiff++
			}
			matcher.RemoveFromMap(stdGroups, ltCfg.MatchKeys, matchedRecord)
		} else {
			// 未精确匹配：待对比端多出的记录计为差异，不参与字段比较
			unmatchedLogRecords++
			unmatchedLogList = append(unmatchedLogList, model.UnmatchedRecord{
				Record:   &logRecord,
				MatchKey: matcher.BuildMatchKey(&logRecord, ltCfg.MatchKeys),
			})
			matchedRecord = nil
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

	// 标准端在精确匹配消耗后剩余的记录即为"缺少"的记录
	unmatchedStdRecords := countGroupedRecords(stdGroups)
	unmatchedStdList := collectGroupedRecords(stdGroups, ltCfg.MatchKeys)

	return &model.LogTypeResult{
		LogType:             ltCfg.Name,
		TotalLogRecords:     len(logResult.Records),
		TotalStdRecords:     len(stdResult.Records),
		MatchKeyCount:       len(logGroups),
		StdMatchKeyCount:    len(stdGroups),
		CommonMatchKeys:     common,
		LogOnlyMatchKeys:    logOnly,
		StdOnlyMatchKeys:    stdOnly,
		RecordsWithDiff:     recordsWithDiff,
		UnmatchedLogRecords: unmatchedLogRecords,
		UnmatchedStdRecords: unmatchedStdRecords,
		UnmatchedLogList:    unmatchedLogList,
		UnmatchedStdList:    unmatchedStdList,
		RequiredStats:       requiredStats,
		ComparisonDetails:   comparisonDetails,
	}, nil
}

// countGroupedRecords 统计分组中剩余待匹配记录的总条数
func countGroupedRecords(groups map[string][]*model.LogRecord) int {
	total := 0
	for _, candidates := range groups {
		total += len(candidates)
	}
	return total
}

// collectGroupedRecords 收集分组中剩余待匹配的记录明细（含各自的 match key）
func collectGroupedRecords(groups map[string][]*model.LogRecord, matchKeys []int) []model.UnmatchedRecord {
	var list []model.UnmatchedRecord
	for _, candidates := range groups {
		for _, c := range candidates {
			list = append(list, model.UnmatchedRecord{
				Record:   c,
				MatchKey: matcher.BuildMatchKey(c, matchKeys),
			})
		}
	}
	return list
}

func toPtrSlice(records []model.LogRecord) []*model.LogRecord {
	ptrs := make([]*model.LogRecord, len(records))
	for i := range records {
		ptrs[i] = &records[i]
	}
	return ptrs
}
