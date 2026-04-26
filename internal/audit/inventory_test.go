package audit

import (
	"encoding/json"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestCollectInventorySkipsRuntimeByDefault(t *testing.T) {
	root := t.TempDir()
	writeFile(t, filepath.Join(root, ".codex", "plans", "active", "sample.md"), "# sample\n")
	writeFile(t, filepath.Join(root, ".sango", "config.json"), "{}\n")
	writeFile(t, filepath.Join(root, ".sango", "logs", "agent.log"), "runtime\n")
	writeFile(t, filepath.Join(root, ".hwt", "state.json"), "{}\n")

	report, err := CollectInventory(root, Options{})
	if err != nil {
		t.Fatal(err)
	}

	if hasPath(report.Items, ".sango/logs/agent.log") {
		t.Fatalf("runtime file should be excluded by default")
	}
	if !hasCategory(report.Items, CategoryConfig) {
		t.Fatalf("expected config category item")
	}
	if report.Summary[CategoryRuntime] != 0 {
		t.Fatalf("runtime summary should be 0, got %d", report.Summary[CategoryRuntime])
	}
	if !hasWarning(report.Warnings, "missing target: .claude") {
		t.Fatalf("expected missing target warning for .claude")
	}
}

func TestCollectInventoryIncludesRuntimeWithFlag(t *testing.T) {
	root := t.TempDir()
	writeFile(t, filepath.Join(root, ".sango", "config.json"), "{}\n")
	writeFile(t, filepath.Join(root, ".sango", "logs", "agent.log"), "runtime\n")

	report, err := CollectInventory(root, Options{IncludeRuntime: true})
	if err != nil {
		t.Fatal(err)
	}

	if !hasPath(report.Items, ".sango/logs/agent.log") {
		t.Fatalf("runtime file should be included")
	}
	if report.Summary[CategoryRuntime] == 0 {
		t.Fatalf("runtime summary should be > 0")
	}
}

func TestReportJSONSchema(t *testing.T) {
	root := t.TempDir()
	writeFile(t, filepath.Join(root, ".sango", "config.json"), "{}\n")

	report, err := CollectInventory(root, Options{})
	if err != nil {
		t.Fatal(err)
	}

	raw, err := json.Marshal(report)
	if err != nil {
		t.Fatal(err)
	}

	var decoded map[string]any
	if err := json.Unmarshal(raw, &decoded); err != nil {
		t.Fatal(err)
	}

	for _, key := range []string{"root", "generated_at", "targets", "excluded", "items", "summary", "warnings"} {
		if _, ok := decoded[key]; !ok {
			t.Fatalf("missing key: %s", key)
		}
	}
}

func writeFile(t *testing.T, path, body string) {
	t.Helper()
	if err := os.MkdirAll(filepath.Dir(path), 0o755); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(path, []byte(body), 0o644); err != nil {
		t.Fatal(err)
	}
}

func hasPath(items []Item, path string) bool {
	for _, item := range items {
		if item.Path == path {
			return true
		}
	}
	return false
}

func hasCategory(items []Item, category string) bool {
	for _, item := range items {
		if item.Category == category {
			return true
		}
	}
	return false
}

func hasWarning(warnings []string, expected string) bool {
	for _, warning := range warnings {
		if strings.Contains(warning, expected) {
			return true
		}
	}
	return false
}
