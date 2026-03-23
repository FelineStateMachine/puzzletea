package pdfexport

import (
	"bufio"
	"bytes"
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"slices"
	"strings"
)

const ExportSchemaV1 = "puzzletea.export.v1"

type JSONLRecord struct {
	Schema string        `json:"schema"`
	Pack   JSONLPackMeta `json:"pack"`
	Puzzle JSONLPuzzle   `json:"puzzle"`
}

type JSONLPackMeta struct {
	Generated     string `json:"generated"`
	Version       string `json:"version"`
	Category      string `json:"category"`
	ModeSelection string `json:"mode_selection"`
	Count         int    `json:"count"`
	Seed          string `json:"seed"`
}

type JSONLPuzzle struct {
	Index int             `json:"index"`
	Name  string          `json:"name"`
	Game  string          `json:"game"`
	Mode  string          `json:"mode"`
	Save  json.RawMessage `json:"save"`
}

func ParseJSONLFiles(paths []string) ([]PackDocument, error) {
	docs := make([]PackDocument, 0, len(paths))
	for _, path := range paths {
		doc, err := ParseJSONLFile(path)
		if err != nil {
			return nil, err
		}
		docs = append(docs, doc)
	}
	return docs, nil
}

func ParseJSONLFile(path string) (PackDocument, error) {
	if !strings.EqualFold(filepath.Ext(path), ".jsonl") {
		return PackDocument{}, fmt.Errorf("%s: expected .jsonl input", path)
	}

	f, err := os.Open(path)
	if err != nil {
		return PackDocument{}, fmt.Errorf("open input jsonl: %w", err)
	}
	defer f.Close()

	scanner := bufio.NewScanner(f)
	scanner.Buffer(make([]byte, 0, 64*1024), 16*1024*1024)
	return parseJSONLScanner(path, scanner)
}

func ParseJSONLData(sourceName string, data []byte) (PackDocument, error) {
	scanner := bufio.NewScanner(bytes.NewReader(data))
	scanner.Buffer(make([]byte, 0, 64*1024), 16*1024*1024)
	return parseJSONLScanner(sourceName, scanner)
}

func ParseJSONLRecords(sourceName string, records []JSONLRecord) (PackDocument, error) {
	var b strings.Builder
	for _, record := range records {
		line, err := json.Marshal(record)
		if err != nil {
			return PackDocument{}, fmt.Errorf("encode jsonl record: %w", err)
		}
		b.Write(line)
		b.WriteByte('\n')
	}
	return ParseJSONLData(sourceName, []byte(b.String()))
}

func parseJSONLScanner(path string, scanner *bufio.Scanner) (PackDocument, error) {
	doc := PackDocument{SourcePath: path}
	doc.Metadata.SourceFileName = filepath.Base(path)
	puzzles := []Puzzle{}
	lineNo := 0
	seenAny := false
	unsupportedCategories := []string{}

	for scanner.Scan() {
		lineNo++
		line := strings.TrimSpace(scanner.Text())
		if line == "" {
			continue
		}
		seenAny = true

		var record JSONLRecord
		if err := json.Unmarshal([]byte(line), &record); err != nil {
			return PackDocument{}, fmt.Errorf("%s:%d: decode jsonl record: %w", path, lineNo, err)
		}
		if record.Schema != ExportSchemaV1 {
			return PackDocument{}, fmt.Errorf("%s:%d: unsupported schema %q", path, lineNo, record.Schema)
		}

		category := strings.TrimSpace(record.Puzzle.Game)
		if category == "" {
			category = strings.TrimSpace(record.Pack.Category)
		}
		mode := strings.TrimSpace(record.Puzzle.Mode)
		if mode == "" {
			mode = strings.TrimSpace(record.Pack.ModeSelection)
		}

		p := Puzzle{
			SourcePath:     path,
			SourceFileName: filepath.Base(path),
			Category:       category,
			ModeSelection:  mode,
			Name:           record.Puzzle.Name,
			Index:          record.Puzzle.Index,
			SaveData:       append([]byte(nil), record.Puzzle.Save...),
		}

		adapter, ok := LookupPrintAdapter(category)
		if !ok {
			if category != "" && !slices.Contains(unsupportedCategories, category) {
				unsupportedCategories = append(unsupportedCategories, category)
			}
			continue
		}
		payload, err := adapter.BuildPDFPayload(p.SaveData)
		if err != nil {
			return PackDocument{}, fmt.Errorf("%s:%d: build print payload: %w", path, lineNo, err)
		}
		if IsNilPrintPayload(payload) {
			continue
		}
		p.PrintPayload = payload

		if doc.Metadata.GeneratedRaw == "" {
			doc.Metadata.GeneratedRaw = record.Pack.Generated
			doc.Metadata.Version = record.Pack.Version
			doc.Metadata.Count = record.Pack.Count
			doc.Metadata.Seed = record.Pack.Seed
			doc.Metadata.Category = strings.TrimSpace(p.Category)
			if doc.Metadata.Category == "" {
				doc.Metadata.Category = record.Pack.Category
			}
			doc.Metadata.ModeSelection = strings.TrimSpace(p.ModeSelection)
			if doc.Metadata.ModeSelection == "" {
				doc.Metadata.ModeSelection = record.Pack.ModeSelection
			}
		}

		puzzles = append(puzzles, p)
	}

	if err := scanner.Err(); err != nil {
		return PackDocument{}, fmt.Errorf("read input jsonl: %w", err)
	}
	if !seenAny {
		return PackDocument{}, fmt.Errorf("%s: input jsonl is empty", path)
	}
	if len(puzzles) == 0 {
		if len(unsupportedCategories) > 0 {
			return PackDocument{}, fmt.Errorf("%s: no printable puzzles found; unsupported categories: %s", path, strings.Join(unsupportedCategories, ", "))
		}
		return PackDocument{}, fmt.Errorf("%s: no printable puzzles found", path)
	}
	if doc.Metadata.Count == 0 {
		doc.Metadata.Count = len(puzzles)
	}
	doc.Puzzles = puzzles
	return doc, nil
}
