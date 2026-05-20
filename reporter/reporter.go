package reporter

import (
	"fmt"
	"log-compare/model"
	"os"
	"strings"
)

// PrintTerminal 打印终端摘要报告
func PrintTerminal(results map[string]*model.LogTypeResult) {
	fmt.Println(strings.Repeat("=", 60))
	fmt.Println("日志对比工具 - 执行结果")
	fmt.Println(strings.Repeat("=", 60))

	totalLog := 0
	totalStd := 0
	totalDiff := 0

	for _, r := range results {
		totalLog += r.TotalLogRecords
		totalStd += r.TotalStdRecords
		totalDiff += r.RecordsWithDiff
	}

	fmt.Printf("\n总体统计:\n")
	fmt.Printf("总生成日志记录数: %d\n", totalLog)
	fmt.Printf("总标准日志记录数: %d\n", totalStd)
	fmt.Printf("总差异记录数: %d\n", totalDiff)

	if totalLog == totalStd {
		fmt.Println("✓ 总体条目数一致性: 一致")
	} else {
		fmt.Println("✗ 总体条目数一致性: 不一致")
	}

	fmt.Printf("\n各类日志差异统计:\n")
	fmt.Println(strings.Repeat("-", 40))

	for logType, r := range results {
		fmt.Printf("%s:\n", logType)
		fmt.Printf("  生成日志记录数: %d\n", r.TotalLogRecords)
		fmt.Printf("  标准日志记录数: %d\n", r.TotalStdRecords)
		fmt.Printf("  存在差异的记录数: %d\n", r.RecordsWithDiff)
		if r.RequiredStats.RecordsWithMissing > 0 {
			fmt.Printf("  缺失必填字段的记录数: %d\n", r.RequiredStats.RecordsWithMissing)
		}
	}

	fmt.Println(strings.Repeat("=", 60))
}

// WriteFile 写入详细报告文件
func WriteFile(results map[string]*model.LogTypeResult, outputPath string) error {
	f, err := os.Create(outputPath)
	if err != nil {
		return fmt.Errorf("创建报告文件失败: %w", err)
	}
	defer f.Close()

	w := &writer{f: f}

	w.writeln(strings.Repeat("=", 80))
	w.writeln("日志对比工具详细报告")
	w.writeln(strings.Repeat("=", 80))

	for logType, r := range results {
		w.writeln(fmt.Sprintf("\n日志类型: %s", logType))
		w.writeln(strings.Repeat("-", 40))
		w.writeln(fmt.Sprintf("生成日志记录数: %d", r.TotalLogRecords))
		w.writeln(fmt.Sprintf("标准日志记录数: %d", r.TotalStdRecords))
		w.writeln(fmt.Sprintf("存在差异的记录数: %d", r.RecordsWithDiff))
		w.writeln(fmt.Sprintf("Match Key 数量: 日志=%d, 标准=%d, 共同=%d",
			r.MatchKeyCount, r.StdMatchKeyCount, r.CommonMatchKeys))

		if r.RequiredStats.RecordsWithMissing > 0 {
			w.writeln(fmt.Sprintf("缺失必填字段的记录数: %d", r.RequiredStats.RecordsWithMissing))
			for field, count := range r.RequiredStats.MissingFieldsSummary {
				w.writeln(fmt.Sprintf("  %s: %d 条", field, count))
			}
		}

		diffCount := 0
		for _, cr := range r.ComparisonDetails {
			if len(cr.Differences) > 0 {
				diffCount++
				w.writeln(fmt.Sprintf("\n记录 #%d:", diffCount))
				w.writeln(fmt.Sprintf("  文件: %s (行号: %d)", cr.LogRecord.FilePath, cr.LogRecord.LineNumber))
				w.writeln(fmt.Sprintf("  Match Key: %s", cr.MatchKey))
				w.writeln(fmt.Sprintf("  差异字段:"))
				for _, d := range cr.Differences {
					w.writeln(fmt.Sprintf("    [%d] %s: 日志=%q, 标准=%q", d.Index, d.Name, d.LogValue, d.StdValue))
				}
			}
		}
	}

	return w.err
}

type writer struct {
	f   *os.File
	err error
}

func (w *writer) writeln(s string) {
	if w.err != nil {
		return
	}
	_, w.err = w.f.WriteString(s + "\n")
}
