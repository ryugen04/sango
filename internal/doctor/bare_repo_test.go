package doctor

import (
	"os"
	"os/exec"
	"path/filepath"
	"testing"

	"github.com/ryugen04/sango/internal/worktree"
)

func TestCheckBareRepoFetchRefspecPass(t *testing.T) {
	sangoDir := t.TempDir()
	bareDir := filepath.Join(sangoDir, "bare", "api.git")
	initBareRepo(t, bareDir)
	if err := worktree.EnsureBareRepoFetchRefspec(bareDir); err != nil {
		t.Fatalf("EnsureBareRepoFetchRefspec 失敗: %v", err)
	}

	results := CheckBareRepoFetchRefspec(sangoDir)
	if len(results) != 1 {
		t.Fatalf("len(results) = %d, want 1", len(results))
	}
	if results[0].Status != StatusPass {
		t.Fatalf("Status = %s, want %s: %s", results[0].Status, StatusPass, results[0].Message)
	}
}

func TestCheckBareRepoFetchRefspecFail(t *testing.T) {
	sangoDir := t.TempDir()
	bareDir := filepath.Join(sangoDir, "bare", "api.git")
	initBareRepo(t, bareDir)

	results := CheckBareRepoFetchRefspec(sangoDir)
	if len(results) != 1 {
		t.Fatalf("len(results) = %d, want 1", len(results))
	}
	if results[0].Status != StatusFail {
		t.Fatalf("Status = %s, want %s", results[0].Status, StatusFail)
	}
	if results[0].Fix == "" {
		t.Fatal("Fix が空です")
	}
}

func initBareRepo(t *testing.T, dir string) {
	t.Helper()
	if err := os.MkdirAll(dir, 0o755); err != nil {
		t.Fatalf("bare repo ディレクトリ作成失敗: %v", err)
	}
	cmd := exec.Command("git", "init", "--bare")
	cmd.Dir = dir
	cmd.Env = append(os.Environ(), "GIT_CONFIG_GLOBAL=/dev/null", "GIT_CONFIG_SYSTEM=/dev/null")
	out, err := cmd.CombinedOutput()
	if err != nil {
		t.Fatalf("git init --bare 失敗: %v\n%s", err, out)
	}
}
