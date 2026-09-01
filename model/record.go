package model

// LogRecord 日志记录
type LogRecord struct {
	Fields     []string
	FilePath   string
	LineNumber int
	TarName    string // 若来自 tar 包，记录包内文件名
}

// UnmatchedRecord 未匹配的记录（多出/缺少）
type UnmatchedRecord struct {
	Record   *LogRecord
	MatchKey string
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
	LogType             string
	TotalLogRecords     int
	TotalStdRecords     int
	MatchKeyCount       int
	StdMatchKeyCount    int
	CommonMatchKeys     int
	LogOnlyMatchKeys    int
	StdOnlyMatchKeys    int
	RecordsWithDiff     int               // 已精确匹配且字段存在差异的记录数
	UnmatchedLogRecords int               // 待对比端存在但标准端无对应记录的记录数（多出）
	UnmatchedStdRecords int               // 标准端存在但待对比端无对应记录的记录数（缺少）
	UnmatchedLogList    []UnmatchedRecord // 多出的记录明细
	UnmatchedStdList    []UnmatchedRecord // 缺少的记录明细
	RequiredStats       RequiredFieldStats
	ComparisonDetails   []ComparisonResult
}

// DiffRecordCount 返回该日志类型的总差异记录数：
// 字段差异记录 + 待对比端未匹配记录 + 标准端未匹配记录
func (r *LogTypeResult) DiffRecordCount() int {
	return r.RecordsWithDiff + r.UnmatchedLogRecords + r.UnmatchedStdRecords
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
