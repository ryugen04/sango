package cmd

import (
	"bytes"
	"encoding/json"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestRunAuditInventoryBothWritesTextAndJSON(t *testing.T) {
	root := t.TempDir()
	writeAuditFixture(t, filepath.Join(root, ".codex", "plans", "active", "sample.md"), "# sample\n")
	writeAuditFixture(t, filepath.Join(root, ".sango", "config.json"), "{}\n")
	writeAuditFixture(t, filepath.Join(root, ".sango", "logs", "agent.log"), "runtime\n")

	var stdout bytes.Buffer
	var stderr bytes.Buffer
	if err := runAuditInventory(root, ".sango/audit/inventory.json", "both", false, &stdout, &stderr); err != nil {
		t.Fatal(err)
	}

	if !strings.Contains(stdout.String(), "inventory root: "+root) {
		t.Fatalf("stdout should include inventory summary: %s", stdout.String())
	}
	if !strings.Contains(stdout.String(), "json written: ") {
		t.Fatalf("stdout should include json output path: %s", stdout.String())
	}
	if !strings.Contains(stderr.String(), "warning: missing target: .claude") {
		t.Fatalf("stderr should include warning: %s", stderr.String())
	}

	reportPath := filepath.Join(root, ".sango", "audit", "inventory.json")
	raw, err := os.ReadFile(reportPath)
	if err != nil {
		t.Fatal(err)
	}

	var report map[string]any
	if err := json.Unmarshal(raw, &report); err != nil {
		t.Fatal(err)
	}
	items, ok := report["items"].([]any)
	if !ok || len(items) == 0 {
		t.Fatalf("json should contain items")
	}
	if strings.Contains(string(raw), ".sango/logs/agent.log") {
		t.Fatalf("runtime file should not be written without --include-runtime")
	}
}

func TestRunAuditInventoryIncludesRuntime(t *testing.T) {
	root := t.TempDir()
	writeAuditFixture(t, filepath.Join(root, ".sango", "logs", "agent.log"), "runtime\n")

	var stdout bytes.Buffer
	var stderr bytes.Buffer
	if err := runAuditInventory(root, ".sango/audit/inventory.json", "json", true, &stdout, &stderr); err != nil {
		t.Fatal(err)
	}

	if stdout.Len() != 0 {
		t.Fatalf("json format should not print text summary")
	}
	if !strings.Contains(stderr.String(), "warning: missing target: .claude") {
		t.Fatalf("stderr should include missing target warnings: %s", stderr.String())
	}

	raw, err := os.ReadFile(filepath.Join(root, ".sango", "audit", "inventory.json"))
	if err != nil {
		t.Fatal(err)
	}
	if !strings.Contains(string(raw), ".sango/logs/agent.log") {
		t.Fatalf("runtime file should be present in json output")
	}
}

func writeAuditFixture(t *testing.T, path, body string) {
	t.Helper()
	if err := os.MkdirAll(filepath.Dir(path), 0o755); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(path, []byte(body), 0o644); err != nil {
		t.Fatal(err)
	}
}
