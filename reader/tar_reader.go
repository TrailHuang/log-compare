package reader

import (
	"archive/tar"
	"fmt"
	"io"
	"log-compare/config"
	"log-compare/model"
	"os"
	"strings"
)

func readTarFile(filePath string, ltCfg *config.LogTypeConfig) (*ReadResult, error) {
	f, err := os.Open(filePath)
	if err != nil {
		return nil, err
	}
	defer f.Close()

	tr := tar.NewReader(f)
	var allRecords []model.LogRecord
	var header []string
	delimiter := ltCfg.Delimiter

	for {
		hdr, err := tr.Next()
		if err == io.EOF {
			break
		}
		if err != nil {
			fmt.Fprintf(os.Stderr, "警告: tar 条目读取失败 %s: %v\n", filePath, err)
			continue
		}
		if !strings.HasSuffix(hdr.Name, ".txt") {
			continue
		}

		content, err := io.ReadAll(tr)
		if err != nil {
			fmt.Fprintf(os.Stderr, "警告: tar 内容读取失败 %s/%s: %v\n", filePath, hdr.Name, err)
			continue
		}

		lines := strings.Split(string(content), "\n")
		lineNum := 0

		for _, line := range lines {
			lineNum++
			line = strings.TrimSpace(line)
			if line == "" {
				continue
			}

			if delimiter == "" {
				delimiter, err = detectDelimiter(line)
				if err != nil {
					return nil, fmt.Errorf("%s/%s: %w", filePath, hdr.Name, err)
				}
			}

			fields := splitLine(line, delimiter)

			if lineNum == 1 {
				if len(header) == 0 {
					header = fields
				}
				continue
			}

			allRecords = append(allRecords, model.LogRecord{
				Fields:     fields,
				FilePath:   filePath,
				LineNumber: lineNum,
				TarName:    hdr.Name,
			})
		}
	}

	return &ReadResult{
		Records: allRecords,
		Header:  header,
	}, nil
}
