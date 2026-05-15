package output

import (
	"bytes"
	"encoding/json"
	"strings"
	"testing"
)

func TestPrintPRTable(t *testing.T) {
	rows := []PRRow{
		{Number: 1, Title: "Fix critical bug", Branch: "fix/bug", Status: "Open", Author: "Alice"},
		{Number: 2, Title: "Add new feature", Branch: "feat/new", Status: "Open", Author: "Bob"},
	}
	var buf bytes.Buffer
	PrintPRTable(&buf, rows)

	out := buf.String()
	if !strings.Contains(out, "Fix critical bug") {
		t.Errorf("output missing PR title: %q", out)
	}
	if !strings.Contains(out, "Alice") {
		t.Errorf("output missing author: %q", out)
	}
}

func TestPrintPRTable_Truncation(t *testing.T) {
	longTitle := strings.Repeat("A", 60)
	rows := []PRRow{{Number: 1, Title: longTitle, Branch: "feat", Status: "Open", Author: "Alice"}}
	var buf bytes.Buffer
	PrintPRTable(&buf, rows)

	out := buf.String()
	if strings.Contains(out, longTitle) {
		t.Error("expected long title to be truncated")
	}
}

func TestPrintJSON(t *testing.T) {
	data := map[string]any{"id": 1, "title": "Test PR"}
	var buf bytes.Buffer
	if err := PrintJSON(&buf, data); err != nil {
		t.Fatalf("PrintJSON() error = %v", err)
	}

	var parsed map[string]any
	if err := json.Unmarshal(buf.Bytes(), &parsed); err != nil {
		t.Fatalf("output is not valid JSON: %v", err)
	}
	if parsed["title"] != "Test PR" {
		t.Errorf("parsed title = %v, want %q", parsed["title"], "Test PR")
	}
}
