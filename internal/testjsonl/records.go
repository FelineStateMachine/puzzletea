package testjsonl

import (
	"bufio"
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/FelineStateMachine/puzzletea/pdfexport"
)

func LoadRecords(path string) ([]pdfexport.JSONLRecord, error) {
	if !strings.EqualFold(filepath.Ext(path), ".jsonl") {
		return nil, fmt.Errorf("%s: expected .jsonl input", path)
	}

	f, err := os.Open(path)
	if err != nil {
		return nil, fmt.Errorf("open input jsonl: %w", err)
	}
	defer f.Close()

	scanner := bufio.NewScanner(f)
	scanner.Buffer(make([]byte, 0, 64*1024), 16*1024*1024)

	records := make([]pdfexport.JSONLRecord, 0, 16)
	lineNo := 0
	for scanner.Scan() {
		lineNo++
		line := strings.TrimSpace(scanner.Text())
		if line == "" {
			continue
		}

		var record pdfexport.JSONLRecord
		if err := json.Unmarshal([]byte(line), &record); err != nil {
			return nil, fmt.Errorf("%s:%d: decode jsonl record: %w", path, lineNo, err)
		}
		if record.Schema != pdfexport.ExportSchemaV1 {
			return nil, fmt.Errorf("%s:%d: unsupported schema %q", path, lineNo, record.Schema)
		}

		records = append(records, record)
	}
	if err := scanner.Err(); err != nil {
		return nil, fmt.Errorf("read input jsonl: %w", err)
	}
	if len(records) == 0 {
		return nil, fmt.Errorf("%s: input jsonl is empty", path)
	}

	return records, nil
}
