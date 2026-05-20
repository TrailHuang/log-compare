package reader

import (
	"archive/tar"
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
			continue
		}
		if !strings.HasSuffix(hdr.Name, ".txt") {
			continue
		}

		content, err := io.ReadAll(tr)
		if err != nil {
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
				delimiter = detectDelimiter(line)
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
