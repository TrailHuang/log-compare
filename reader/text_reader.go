package reader

import (
	"bufio"
	"fmt"
	"github.com/TrailHuang/log-compare/config"
	"github.com/TrailHuang/log-compare/model"
	"os"
	"path/filepath"
	"strings"
)

func readTextFile(filePath string, ltCfg *config.LogTypeConfig) (*ReadResult, error) {
	f, err := os.Open(filePath)
	if err != nil {
		return nil, err
	}
	defer f.Close()

	var records []model.LogRecord
	var header []string
	scanner := bufio.NewScanner(f)
	lineNum := 0
	delimiter := ltCfg.Delimiter

	for scanner.Scan() {
		lineNum++
		line := strings.TrimSpace(scanner.Text())
		if line == "" {
			continue
		}

		if delimiter == "" {
			delimiter, err = detectDelimiter(line)
			if err != nil {
				return nil, fmt.Errorf("%s: %w", filePath, err)
			}
		}

		fields := splitLine(line, delimiter)

		if lineNum == 1 {
			header = fields
			continue
		}

		records = append(records, model.LogRecord{
			Fields:     fields,
			FilePath:   filePath,
			LineNumber: lineNum,
		})
	}

	return &ReadResult{
		Records: records,
		Header:  header,
	}, scanner.Err()
}

func detectDelimiter(line string) (string, error) {
	if strings.Contains(line, "|++|") {
		return "|++|", nil
	}
	if strings.Contains(line, ",") {
		return ",", nil
	}
	if strings.Contains(line, "\t") {
		return "\t", nil
	}
	if strings.Contains(line, "|") {
		return "|", nil
	}
	return "", fmt.Errorf("无法自动检测分隔符，请在配置中指定 delimiter")
}

func splitLine(line, delimiter string) []string {
	parts := strings.Split(line, delimiter)
	fields := make([]string, len(parts))
	for i, p := range parts {
		fields[i] = strings.TrimSpace(p)
	}
	return fields
}

func findFiles(baseDir, pattern string) ([]string, error) {
	var matches []string
	err := filepath.Walk(baseDir, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return nil
		}
		matched, _ := filepath.Match(pattern, info.Name())
		if matched && !info.IsDir() {
			matches = append(matches, path)
		}
		return nil
	})
	return matches, err
}

func isTarFile(path string) bool {
	return strings.HasSuffix(path, ".tar")
}
