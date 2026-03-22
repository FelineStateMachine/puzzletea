package cmd

import (
	"bufio"
	"encoding/json"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"strings"

	tea "charm.land/bubbletea/v2"
	"github.com/FelineStateMachine/puzzletea/pdfexport"
	"github.com/FelineStateMachine/puzzletea/registry"
	"github.com/spf13/cobra"
)

const (
	testRenderWidth  = 120
	testRenderHeight = 40
)

var testOutput string

type testInputRecord struct {
	lineNo int
	record pdfexport.JSONLRecord
}

var testCmd = &cobra.Command{
	Use:   "test <input.jsonl>",
	Short: "Render saved puzzle exports into an ANSI review artifact",
	Long:  "Parse a PuzzleTea export JSONL file, import each saved puzzle, and render the full board view in a stable ANSI text format for visual review.",
	Args:  cobra.ExactArgs(1),
	RunE:  runTest,
}

func init() {
	testCmd.Flags().StringVarP(&testOutput, "output", "o", "", "write rendered output to a file (defaults to stdout)")
}

func runTest(cmd *cobra.Command, args []string) error {
	applyTheme(flagTheme)

	records, err := loadTestRecords(args[0])
	if err != nil {
		return err
	}

	output, err := renderTestRecords(args[0], records)
	if err != nil {
		return err
	}

	return writeTestOutput(cmd.OutOrStdout(), testOutput, output)
}

func loadTestRecords(path string) ([]testInputRecord, error) {
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

	records := make([]testInputRecord, 0, 16)
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

		records = append(records, testInputRecord{lineNo: lineNo, record: record})
	}
	if err := scanner.Err(); err != nil {
		return nil, fmt.Errorf("read input jsonl: %w", err)
	}
	if len(records) == 0 {
		return nil, fmt.Errorf("%s: input jsonl is empty", path)
	}

	return records, nil
}

func renderTestRecords(path string, records []testInputRecord) (string, error) {
	var b strings.Builder

	for _, item := range records {
		record := item.record
		gameType := strings.TrimSpace(record.Puzzle.Game)
		if gameType == "" {
			gameType = strings.TrimSpace(record.Pack.Category)
		}
		if gameType == "" {
			return "", fmt.Errorf("%s:%d: missing puzzle game/category", path, item.lineNo)
		}

		g, err := registry.Import(gameType, record.Puzzle.Save)
		if err != nil {
			return "", fmt.Errorf("%s:%d: import %q: %w", path, item.lineNo, gameType, err)
		}

		sized, _ := g.Update(tea.WindowSizeMsg{Width: testRenderWidth, Height: testRenderHeight})
		fmt.Fprintf(
			&b,
			"=== %s | %s | %s | #%d ===\n",
			gameType,
			testHeaderValue(record.Puzzle.Mode, record.Pack.ModeSelection),
			testHeaderValue(record.Puzzle.Name, "unnamed"),
			record.Puzzle.Index,
		)
		b.WriteString(sized.View())
		b.WriteString("\n\n")
	}

	return b.String(), nil
}

func testHeaderValue(value, fallback string) string {
	value = strings.TrimSpace(value)
	if value != "" {
		return value
	}
	return fallback
}

func writeTestOutput(stdout io.Writer, path, content string) error {
	if strings.TrimSpace(path) == "" {
		_, err := io.WriteString(stdout, content)
		return err
	}

	dir := filepath.Dir(path)
	if dir != "" && dir != "." {
		if err := os.MkdirAll(dir, 0o755); err != nil {
			return fmt.Errorf("create output directory: %w", err)
		}
	}
	if err := os.WriteFile(path, []byte(content), 0o644); err != nil {
		return fmt.Errorf("write output file: %w", err)
	}
	return nil
}
