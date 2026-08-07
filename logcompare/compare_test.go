package logcompare

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/TrailHuang/log-compare/config"
	"github.com/TrailHuang/log-compare/model"
)

// helper: 在临时目录里创建若干日志文件，返回目录路径
func writeLogFiles(t *testing.T, dir string, files map[string]string) {
	t.Helper()
	for name, content := range files {
		full := filepath.Join(dir, name)
		if err := os.MkdirAll(filepath.Dir(full), 0o755); err != nil {
			t.Fatalf("MkdirAll(%s) 失败: %v", filepath.Dir(full), err)
		}
		if err := os.WriteFile(full, []byte(content), 0o644); err != nil {
			t.Fatalf("WriteFile(%s) 失败: %v", full, err)
		}
	}
}

// 构造一个最小可用的 LogTypeConfig（内联 field_names，避免依赖外部文件）
func newLTConfig(name, delim, pattern string, matchKeys, required, filter []int) *config.LogTypeConfig {
	// 内联 field_names：{"0":"f0","1":"f1",...}
	fields := []string{`{"0":"f0","1":"f1","2":"f2","3":"f3","4":"f4","5":"f5"}`}
	_ = fields
	lt := &config.LogTypeConfig{
		Name:           name,
		Delimiter:      delim,
		FilePattern:    pattern,
		MatchKeys:      matchKeys,
		RequiredFields: required,
		FilterFields:   filter,
		FieldNames:     map[int]string{0: "f0", 1: "f1", 2: "f2", 3: "f3", 4: "f4", 5: "f5"},
	}
	return lt
}

// 直接调用 compareLogType 的白盒测试
func TestCompareLogType_ExactMatch_NoDiff(t *testing.T) {
	tmp := t.TempDir()
	stdDir := filepath.Join(tmp, "std")
	logDir := filepath.Join(tmp, "log")
	// reader 会把第一行当作 header 跳过，因此数据从第二行起
	writeLogFiles(t, stdDir, map[string]string{
		"A_1.txt": "h0|h1|h2|h3|h4|h5\na|b|c|d|e|f\n",
	})
	writeLogFiles(t, logDir, map[string]string{
		"A_1.txt": "h0|h1|h2|h3|h4|h5\na|b|c|d|e|f\n",
	})

	lt := newLTConfig("A", "|", "A_*.txt", []int{0, 1}, nil, nil)
	cfg := &config.Config{}
	r, err := compareLogType(cfg, stdDir, logDir, lt)
	if err != nil {
		t.Fatalf("compareLogType 失败: %v", err)
	}
	if r.RecordsWithDiff != 0 {
		t.Errorf("期望 0 差异, 实际 %d", r.RecordsWithDiff)
	}
	if r.TotalLogRecords != 1 || r.TotalStdRecords != 1 {
		t.Errorf("记录数不符: log=%d std=%d", r.TotalLogRecords, r.TotalStdRecords)
	}
}

func TestCompareLogType_ExactMatch_WithDiff(t *testing.T) {
	tmp := t.TempDir()
	stdDir := filepath.Join(tmp, "std")
	logDir := filepath.Join(tmp, "log")
	// match_keys=[0,1]，标准与日志的 0/1 相同，但字段 2 不同
	writeLogFiles(t, stdDir, map[string]string{
		"A_1.txt": "h0|h1|h2|h3|h4|h5\nk1|k2|std_val|d|e|f\n",
	})
	writeLogFiles(t, logDir, map[string]string{
		"A_1.txt": "h0|h1|h2|h3|h4|h5\nk1|k2|log_val|d|e|f\n",
	})

	lt := newLTConfig("A", "|", "A_*.txt", []int{0, 1}, nil, nil)
	cfg := &config.Config{}
	r, err := compareLogType(cfg, stdDir, logDir, lt)
	if err != nil {
		t.Fatalf("compareLogType 失败: %v", err)
	}
	if r.RecordsWithDiff != 1 {
		t.Fatalf("期望 1 差异, 实际 %d", r.RecordsWithDiff)
	}
	if len(r.ComparisonDetails) != 1 {
		t.Fatalf("期望 1 条明细, 实际 %d", len(r.ComparisonDetails))
	}
	cr := r.ComparisonDetails[0]
	if !cr.MatchFound {
		t.Errorf("期望精确匹配 MatchFound=true")
	}
	if cr.MatchedRecord == nil {
		t.Fatalf("期望 MatchedRecord 非 nil")
	}
	if len(cr.Differences) != 1 {
		t.Fatalf("期望 1 个差异字段, 实际 %d", len(cr.Differences))
	}
	d := cr.Differences[0]
	if d.Index != 2 || d.LogValue != "log_val" || d.StdValue != "std_val" {
		t.Errorf("差异字段不符: %+v", d)
	}
}

// 关键回归：未精确匹配时不应走模糊回退，不应产生伪差异
func TestCompareLogType_NoExactMatch_NoFallbackDiff(t *testing.T) {
	tmp := t.TempDir()
	stdDir := filepath.Join(tmp, "std")
	logDir := filepath.Join(tmp, "log")
	// match_keys=[0,1]，两端 0/1 不同，不应匹配
	writeLogFiles(t, stdDir, map[string]string{
		"A_1.txt": "h0|h1|h2|h3|h4|h5\nstd_k|std_k2|c|d|e|f\n",
	})
	writeLogFiles(t, logDir, map[string]string{
		"A_1.txt": "h0|h1|h2|h3|h4|h5\nlog_k|log_k2|c|d|e|f\n",
	})

	lt := newLTConfig("A", "|", "A_*.txt", []int{0, 1}, nil, nil)
	cfg := &config.Config{}
	r, err := compareLogType(cfg, stdDir, logDir, lt)
	if err != nil {
		t.Fatalf("compareLogType 失败: %v", err)
	}
	if r.RecordsWithDiff != 0 {
		t.Errorf("未精确匹配不应产生差异, 实际 %d", r.RecordsWithDiff)
	}
	if len(r.ComparisonDetails) != 1 {
		t.Fatalf("期望 1 条明细, 实际 %d", len(r.ComparisonDetails))
	}
	cr := r.ComparisonDetails[0]
	if cr.MatchFound {
		t.Errorf("期望 MatchFound=false")
	}
	if cr.MatchedRecord != nil {
		t.Errorf("未精确匹配时 MatchedRecord 应为 nil")
	}
	if len(cr.Differences) != 0 {
		t.Errorf("未精确匹配不应有差异字段, 实际 %d", len(cr.Differences))
	}
}

// 回归双向流错配场景：日志端 src_port=48954 不应匹配到标准端 src_port=38680
func TestCompareLogType_BiFlowNoCrossMatch(t *testing.T) {
	tmp := t.TempDir()
	stdDir := filepath.Join(tmp, "std")
	logDir := filepath.Join(tmp, "log")
	// 标准端两条记录：src_port 分别 38680 / 48954，其它 match key 相同
	stdContent := "h0|h1|h2|h3|h4|h5\n" +
		"ip1|38680|ip2|58557|6|ruleA|std_up|std_down\n" +
		"ip1|48954|ip2|58557|6|ruleA|std_up2|std_down2\n"
	// 日志端只有一条 src_port=48954，精确匹配应命中标准端第二条
	logContent := "h0|h1|h2|h3|h4|h5\nip1|48954|ip2|58557|6|ruleA|log_up2|log_down2\n"
	writeLogFiles(t, stdDir, map[string]string{"A_1.txt": stdContent})
	writeLogFiles(t, logDir, map[string]string{"A_1.txt": logContent})

	// match_keys = [0,1,2,3,4,5]（含 src_port=索引1）
	lt := newLTConfig("A", "|", "A_*.txt", []int{0, 1, 2, 3, 4, 5}, nil, nil)
	cfg := &config.Config{}
	r, err := compareLogType(cfg, stdDir, logDir, lt)
	if err != nil {
		t.Fatalf("compareLogType 失败: %v", err)
	}
	if len(r.ComparisonDetails) != 1 {
		t.Fatalf("期望 1 条明细, 实际 %d", len(r.ComparisonDetails))
	}
	cr := r.ComparisonDetails[0]
	// 关键点：必须精确匹配到 src_port=48954 的标准记录，而非错配到 38680
	if !cr.MatchFound {
		t.Errorf("期望精确匹配命中 src_port=48954 的标准记录")
	}
	if cr.MatchedRecord == nil || cr.MatchedRecord.Fields[1] != "48954" {
		t.Errorf("应匹配到 src_port=48954 的标准记录, got %+v", cr.MatchedRecord)
	}
	// 字段 6/7 (up/down) 不同，应有差异；但 src_port(1) 不应在差异中
	if r.RecordsWithDiff != 1 {
		t.Errorf("期望 1 条差异（up/down 不同）, 实际 %d", r.RecordsWithDiff)
	}
	for _, d := range cr.Differences {
		if d.Index == 1 {
			t.Errorf("src_port 不应出现在差异中（match_key）: %+v", d)
		}
	}
}

func TestCompareLogType_FilterFieldsSkipped(t *testing.T) {
	tmp := t.TempDir()
	stdDir := filepath.Join(tmp, "std")
	logDir := filepath.Join(tmp, "log")
	// 字段 2 在 filter_fields 中，即使不同也不应报告
	writeLogFiles(t, stdDir, map[string]string{
		"A_1.txt": "h0|h1|h2|h3|h4|h5\nk1|k2|std_filtered|d|e|f\n",
	})
	writeLogFiles(t, logDir, map[string]string{
		"A_1.txt": "h0|h1|h2|h3|h4|h5\nk1|k2|log_filtered|d|e|f\n",
	})

	lt := newLTConfig("A", "|", "A_*.txt", []int{0, 1}, nil, []int{2})
	cfg := &config.Config{}
	r, err := compareLogType(cfg, stdDir, logDir, lt)
	if err != nil {
		t.Fatalf("compareLogType 失败: %v", err)
	}
	if r.RecordsWithDiff != 0 {
		t.Errorf("filter_fields 应跳过字段 2, 期望 0 差异, 实际 %d", r.RecordsWithDiff)
	}
}

func TestCompareLogType_RequiredFieldsMissing(t *testing.T) {
	tmp := t.TempDir()
	stdDir := filepath.Join(tmp, "std")
	logDir := filepath.Join(tmp, "log")
	// 日志端字段 3 为空，required_fields 包含 3
	writeLogFiles(t, stdDir, map[string]string{
		"A_1.txt": "h0|h1|h2|h3|h4|h5\nk1|k2|c|std_d|e|f\n",
	})
	writeLogFiles(t, logDir, map[string]string{
		"A_1.txt": "h0|h1|h2|h3|h4|h5\nk1|k2|c||e|f\n",
	})

	lt := newLTConfig("A", "|", "A_*.txt", []int{0, 1}, []int{3}, nil)
	cfg := &config.Config{}
	r, err := compareLogType(cfg, stdDir, logDir, lt)
	if err != nil {
		t.Fatalf("compareLogType 失败: %v", err)
	}
	if r.RequiredStats.RecordsWithMissing != 1 {
		t.Errorf("期望 1 条必填缺失, 实际 %d", r.RequiredStats.RecordsWithMissing)
	}
	if r.RequiredStats.MissingFieldsSummary["f3"] != 1 {
		t.Errorf("期望 f3 缺失统计为 1, 实际 %v", r.RequiredStats.MissingFieldsSummary)
	}
}

func TestCompareLogType_MultipleRecords_OneUnmatched(t *testing.T) {
	tmp := t.TempDir()
	stdDir := filepath.Join(tmp, "std")
	logDir := filepath.Join(tmp, "log")
	// 标准端 2 条，日志端 3 条（1 条无对应）
	stdContent := "h0|h1|h2|h3|h4|h5\nk1|k2|c1|d|e|f\nk3|k4|c2|d|e|f\n"
	logContent := "h0|h1|h2|h3|h4|h5\nk1|k2|c1|d|e|f\nk3|k4|c2_diff|d|e|f\nk5|k6|c3|d|e|f\n"
	writeLogFiles(t, stdDir, map[string]string{"A_1.txt": stdContent})
	writeLogFiles(t, logDir, map[string]string{"A_1.txt": logContent})

	lt := newLTConfig("A", "|", "A_*.txt", []int{0, 1}, nil, nil)
	cfg := &config.Config{}
	r, err := compareLogType(cfg, stdDir, logDir, lt)
	if err != nil {
		t.Fatalf("compareLogType 失败: %v", err)
	}
	if r.TotalLogRecords != 3 || r.TotalStdRecords != 2 {
		t.Errorf("记录数不符: log=%d std=%d", r.TotalLogRecords, r.TotalStdRecords)
	}
	// 只有第 2 条精确匹配且有差异；第 3 条未匹配不应计差异
	if r.RecordsWithDiff != 1 {
		t.Errorf("期望 1 条差异, 实际 %d", r.RecordsWithDiff)
	}
	// 检查第三条未匹配
	unmatched := 0
	for _, cr := range r.ComparisonDetails {
		if !cr.MatchFound {
			unmatched++
		}
	}
	if unmatched != 1 {
		t.Errorf("期望 1 条未匹配, 实际 %d", unmatched)
	}
}

func TestCompareLogType_EmptyDirs(t *testing.T) {
	tmp := t.TempDir()
	stdDir := filepath.Join(tmp, "std")
	logDir := filepath.Join(tmp, "log")
	if err := os.MkdirAll(stdDir, 0o755); err != nil {
		t.Fatal(err)
	}
	if err := os.MkdirAll(logDir, 0o755); err != nil {
		t.Fatal(err)
	}

	lt := newLTConfig("A", "|", "A_*.txt", []int{0, 1}, nil, nil)
	cfg := &config.Config{}
	r, err := compareLogType(cfg, stdDir, logDir, lt)
	if err != nil {
		t.Fatalf("compareLogType 失败: %v", err)
	}
	if r.TotalLogRecords != 0 || r.TotalStdRecords != 0 {
		t.Errorf("空目录应返回 0 记录, log=%d std=%d", r.TotalLogRecords, r.TotalStdRecords)
	}
	if r.RecordsWithDiff != 0 {
		t.Errorf("空目录应无差异, 实际 %d", r.RecordsWithDiff)
	}
}

// 端到端测试 Run：多日志类型
func TestRun_MultiLogTypes(t *testing.T) {
	tmp := t.TempDir()
	stdDir := filepath.Join(tmp, "std")
	logDir := filepath.Join(tmp, "log")
	// 类型 A：1 条匹配无差异
	writeLogFiles(t, stdDir, map[string]string{"A_1.txt": "h0|h1|h2|h3|h4|h5\nk1|k2|c|d|e|f\n"})
	writeLogFiles(t, logDir, map[string]string{"A_1.txt": "h0|h1|h2|h3|h4|h5\nk1|k2|c|d|e|f\n"})
	// 类型 B：1 条匹配有差异
	writeLogFiles(t, stdDir, map[string]string{"B_1.txt": "h0|h1|h2|h3|h4|h5\nk1|k2|std|d|e|f\n"})
	writeLogFiles(t, logDir, map[string]string{"B_1.txt": "h0|h1|h2|h3|h4|h5\nk1|k2|log|d|e|f\n"})

	cfg := &config.Config{
		LogTypes: []config.LogTypeConfig{
			{
				Name:           "A",
				Delimiter:      "|",
				FilePattern:    "A_*.txt",
				MatchKeys:      []int{0, 1},
				RequiredFields: []int{},
				FilterFields:   []int{},
				FieldNames:     map[int]string{0: "f0", 1: "f1", 2: "f2", 3: "f3", 4: "f4", 5: "f5"},
			},
			{
				Name:           "B",
				Delimiter:      "|",
				FilePattern:    "B_*.txt",
				MatchKeys:      []int{0, 1},
				RequiredFields: []int{},
				FilterFields:   []int{},
				FieldNames:     map[int]string{0: "f0", 1: "f1", 2: "f2", 3: "f3", 4: "f4", 5: "f5"},
			},
		},
	}

	overall, err := Run(cfg, stdDir, logDir)
	if err != nil {
		t.Fatalf("Run 失败: %v", err)
	}
	if len(overall.LogTypeResults) != 2 {
		t.Fatalf("期望 2 个日志类型, 实际 %d", len(overall.LogTypeResults))
	}
	if overall.LogTypeResults["A"].RecordsWithDiff != 0 {
		t.Errorf("A 期望 0 差异, 实际 %d", overall.LogTypeResults["A"].RecordsWithDiff)
	}
	if overall.LogTypeResults["B"].RecordsWithDiff != 1 {
		t.Errorf("B 期望 1 差异, 实际 %d", overall.LogTypeResults["B"].RecordsWithDiff)
	}
}

// 端到端测试 Run：日志目录不存在时返回空结果（reader 对不存在目录返回空而非错误）
func TestRun_MissingLogDir_NoError(t *testing.T) {
	tmp := t.TempDir()
	stdDir := filepath.Join(tmp, "std")
	logDir := filepath.Join(tmp, "log") // 不存在
	if err := os.MkdirAll(stdDir, 0o755); err != nil {
		t.Fatal(err)
	}
	writeLogFiles(t, stdDir, map[string]string{"A_1.txt": "h0|h1|h2|h3|h4|h5\nk1|k2|c|d|e|f\n"})

	cfg := &config.Config{
		LogTypes: []config.LogTypeConfig{
			{
				Name:           "A",
				Delimiter:      "|",
				FilePattern:    "A_*.txt",
				MatchKeys:      []int{0, 1},
				RequiredFields: []int{},
				FilterFields:   []int{},
				FieldNames:     map[int]string{0: "f0", 1: "f1"},
			},
		},
	}
	overall, err := Run(cfg, stdDir, logDir)
	if err != nil {
		t.Fatalf("目录不存在不应返回错误, 实际: %v", err)
	}
	if overall.LogTypeResults["A"].TotalLogRecords != 0 {
		t.Errorf("日志端目录不存在应返回 0 记录, 实际 %d", overall.LogTypeResults["A"].TotalLogRecords)
	}
}

// toPtrSlice 辅助函数测试
func TestToPtrSlice(t *testing.T) {
	records := []model.LogRecord{
		{Fields: []string{"a"}},
		{Fields: []string{"b"}},
		{Fields: []string{"c"}},
	}
	ptrs := toPtrSlice(records)
	if len(ptrs) != 3 {
		t.Fatalf("期望 3 个指针, 实际 %d", len(ptrs))
	}
	for i, p := range ptrs {
		if p == nil {
			t.Fatalf("ptrs[%d] 为 nil", i)
		}
		if p.Fields[0] != records[i].Fields[0] {
			t.Errorf("ptrs[%d] 内容不符: got %s want %s", i, p.Fields[0], records[i].Fields[0])
		}
	}
}

func TestToPtrSlice_Empty(t *testing.T) {
	ptrs := toPtrSlice(nil)
	if len(ptrs) != 0 {
		t.Errorf("空切片应返回 0 长度, 实际 %d", len(ptrs))
	}
}
