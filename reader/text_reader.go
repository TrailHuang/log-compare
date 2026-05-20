package reader

import (
	"bufio"
	"log-compare/config"
	"log-compare/model"
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
			delimiter = detectDelimiter(line)
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

func detectDelimiter(line string) string {
	if strings.Contains(line, "|++|") {
		return "|++|"
	}
	if strings.Contains(line, ",") {
		return ","
	}
	if strings.Contains(line, "\t") {
		return "\t"
	}
	return "|"
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
