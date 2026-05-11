package service

import (
	"os"
	"path/filepath"
	"testing"
	"time"

	"github.com/ryugen04/sango/internal/config"
	"github.com/ryugen04/sango/internal/worktree"
)

func TestResolveTargets(t *testing.T) {
	cfg := &config.Config{
		Services: map[string]*config.Service{
			"api": {Type: "process"},
			"web": {Type: "process"},
			"db":  {Type: "docker"},
		},
		Profiles: map[string]config.Profile{
			"backend": {Services: []string{"api", "db"}},
		},
	}

	// 引数指定
	targets := ResolveTargets(cfg, []string{"api"}, "")
	if len(targets) != 1 || targets[0] != "api" {
		t.Errorf("expected [api], got %v", targets)
	}

	// プロファイル指定
	targets = ResolveTargets(cfg, nil, "backend")
	if len(targets) != 2 {
		t.Errorf("expected 2 targets for backend profile, got %d", len(targets))
	}

	// 何も指定なし
	targets = ResolveTargets(cfg, nil, "")
	if targets != nil {
		t.Errorf("expected nil, got %v", targets)
	}

	// 存在しないプロファイル
	targets = ResolveTargets(cfg, nil, "nonexistent")
	if targets != nil {
		t.Errorf("expected nil for nonexistent profile, got %v", targets)
	}
}

func TestBuildDAG(t *testing.T) {
	cfg := &config.Config{
		Services: map[string]*config.Service{
			"api": {Type: "process", DependsOn: []string{"db"}},
			"db":  {Type: "docker"},
			"web": {Type: "process", DependsOn: []string{"api"}},
		},
	}

	d := BuildDAG(cfg)
	order, err := d.Resolve("web")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	// web -> api -> db の順で解決されるはず
	if len(order) != 3 {
		t.Fatalf("expected 3 services, got %d: %v", len(order), order)
	}
	// dbが最初に来るべき
	if order[0] != "db" {
		t.Errorf("expected db first, got %s", order[0])
	}
}

func TestResolveWorkingDir(t *testing.T) {
	// WorkingDirが設定されている場合
	svc := &config.Service{WorkingDir: "src"}
	dir := ResolveWorkingDir(svc, "main", "api")
	expected := "main/api/src"
	if dir != expected {
		t.Errorf("expected %q, got %q", expected, dir)
	}

	// WorkingDirが空でディレクトリも存在しない場合
	svc2 := &config.Service{}
	dir2 := ResolveWorkingDir(svc2, "main", "nonexistent")
	if dir2 != "" {
		t.Errorf("expected empty string, got %q", dir2)
	}
}

func TestResolveWorkingDir_WithRepoName(t *testing.T) {
	// repo_name + WorkingDir: 参照先のディレクトリを使う
	svc := &config.Service{RepoName: "example-backend", WorkingDir: "subdir"}
	dir := ResolveWorkingDir(svc, "main", "billing-api")
	expected := "main/example-backend/subdir"
	if dir != expected {
		t.Errorf("expected %q, got %q", expected, dir)
	}

	// repo_nameなし + WorkingDir: 従来通りserviceNameを使う
	svc2 := &config.Service{WorkingDir: "src"}
	dir2 := ResolveWorkingDir(svc2, "main", "api")
	expected2 := "main/api/src"
	if dir2 != expected2 {
		t.Errorf("expected %q, got %q", expected2, dir2)
	}

	// repo_nameあり + WorkingDir空 + ディレクトリ不在: 空文字列を返す
	svc3 := &config.Service{RepoName: "example-backend"}
	dir3 := ResolveWorkingDir(svc3, "/nonexistent-path", "billing-api")
	if dir3 != "" {
		t.Errorf("expected empty string for nonexistent dir, got %q", dir3)
	}
}

func TestMergeEnv(t *testing.T) {
	env := map[string]string{"A": "1", "B": "2"}
	dynamic := map[string]string{"B": "3", "C": "4"}

	merged := MergeEnv(env, dynamic)
	if merged["A"] != "1" {
		t.Errorf("expected A=1, got A=%s", merged["A"])
	}
	if merged["B"] != "3" {
		t.Errorf("expected B=3 (dynamic override), got B=%s", merged["B"])
	}
	if merged["C"] != "4" {
		t.Errorf("expected C=4, got C=%s", merged["C"])
	}
}

func TestResolveActiveWorktree(t *testing.T) {
	// worktreeフラグが指定されている場合
	result := ResolveActiveWorktree(".sango", "feature/auth")
	if result != "feature/auth" {
		t.Errorf("expected feature/auth, got %s", result)
	}

	// フラグなしでworktrees.jsonが存在しない場合
	result = ResolveActiveWorktree("/nonexistent/.sango", "")
	if result != "main" {
		t.Errorf("expected main, got %s", result)
	}
}

func TestResolveActiveWorktreeWithBaseDirDetectsCustomCWD(t *testing.T) {
	tmpdir := t.TempDir()
	sangoDir := filepath.Join(tmpdir, ".sango")
	ws := &worktree.WorktreeState{
		Active: "main",
		Worktrees: map[string]*worktree.WorktreeInfo{
			"feature/auth": {
				Offset:    100,
				CreatedAt: time.Now(),
				Services:  []string{"api"},
			},
			"main": {
				Offset:    0,
				CreatedAt: time.Now(),
				Services:  []string{"api"},
			},
		},
		SharedServices: map[string]*worktree.SharedService{},
	}
	if err := ws.Save(sangoDir); err != nil {
		t.Fatalf("Save: %v", err)
	}

	cwd := filepath.Join(tmpdir, ".worktrees", "feature", "auth", "api")
	if err := os.MkdirAll(cwd, 0o755); err != nil {
		t.Fatalf("MkdirAll: %v", err)
	}
	t.Chdir(cwd)

	got := ResolveActiveWorktreeWithBaseDir(sangoDir, "", ".worktrees")
	if got != "feature/auth" {
		t.Errorf("ResolveActiveWorktreeWithBaseDir() = %q, want %q", got, "feature/auth")
	}
}

func TestLoadAndValidateConfig(t *testing.T) {
	// 存在しないファイル
	_, err := LoadAndValidateConfig("nonexistent.yaml")
	if err == nil {
		t.Error("expected error for nonexistent file")
	}
}
