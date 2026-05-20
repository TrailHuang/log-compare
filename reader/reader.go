package reader

import (
	"log-compare/config"
	"log-compare/model"
)

// ReadResult 读取结果
type ReadResult struct {
	Records []model.LogRecord
	Header  []string
}

// ReadLogs 根据配置读取所有匹配的日志文件
func ReadLogs(logDir string, ltCfg *config.LogTypeConfig) (*ReadResult, error) {
	files, err := findFiles(logDir, ltCfg.FilePattern)
	if err != nil {
		return nil, err
	}

	var allRecords []model.LogRecord
	var header []string

	for _, f := range files {
		var result *ReadResult
		var err error

		if isTarFile(f) {
			result, err = readTarFile(f, ltCfg)
		} else {
			result, err = readTextFile(f, ltCfg)
		}

		if err != nil {
			continue
		}

		allRecords = append(allRecords, result.Records...)
		if header == nil && len(result.Header) > 0 {
			header = result.Header
		}
	}

	return &ReadResult{
		Records: allRecords,
		Header:  header,
	}, nil
}
