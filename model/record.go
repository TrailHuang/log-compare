package model

// LogRecord 日志记录
type LogRecord struct {
	Fields     []string
	FilePath   string
	LineNumber int
	TarName    string // 若来自 tar 包，记录包内文件名
}

// FieldDiff 字段差异
type FieldDiff struct {
	Index    int
	Name     string
	LogValue string
	StdValue string
}

// ComparisonResult 单条记录的对比结果
type ComparisonResult struct {
	LogRecord       *LogRecord
	MatchedRecord   *LogRecord
	Differences     []FieldDiff
	MatchKey        string
	MatchFound      bool
	RequiredMissing []string
}

// LogTypeResult 单个日志类型的对比结果
type LogTypeResult struct {
	LogType           string
	TotalLogRecords   int
	TotalStdRecords   int
	MatchKeyCount     int
	StdMatchKeyCount  int
	CommonMatchKeys   int
	LogOnlyMatchKeys  int
	StdOnlyMatchKeys  int
	RecordsWithDiff   int
	RequiredStats     RequiredFieldStats
	ComparisonDetails []ComparisonResult
	LogOnlyRecords    []LogRecord
	StdOnlyRecords    []LogRecord
}

// RequiredFieldStats 必填字段统计
type RequiredFieldStats struct {
	TotalRecords         int
	RecordsWithMissing   int
	MissingFieldsSummary map[string]int
}

// OverallResult 总体对比结果
type OverallResult struct {
	LogTypeResults map[string]*LogTypeResult
}
