package cmd

import (
	"os"
	"path/filepath"
	"testing"
	"time"

	"github.com/ryugen04/sango/internal/config"
	sangoLog "github.com/ryugen04/sango/internal/log"
	"github.com/ryugen04/sango/internal/service"
	"github.com/ryugen04/sango/internal/worktree"
)

func TestResolveProjectContextFindsRootFromNestedDirectory(t *testing.T) {
	resetCommandGlobals(t)

	root := t.TempDir()
	writeTestSangoYAML(t, root, "nested-app")
	nested := filepath.Join(root, "web", "src")
	if err := os.MkdirAll(nested, 0o755); err != nil {
		t.Fatalf("mkdir nested: %v", err)
	}

	ctx, err := findProjectContext(nested)
	if err != nil {
		t.Fatalf("findProjectContext: %v", err)
	}
	if ctx.Root != root {
		t.Fatalf("Root = %q, want %q", ctx.Root, root)
	}
	if ctx.ConfigPath != filepath.Join(root, "sango.yaml") {
		t.Fatalf("ConfigPath = %q", ctx.ConfigPath)
	}
	if ctx.Name != "nested-app" {
		t.Fatalf("Name = %q, want nested-app", ctx.Name)
	}
}

func TestResolveProjectContextUsesSymlinkTargetAsRoot(t *testing.T) {
	resetCommandGlobals(t)

	root := t.TempDir()
	configPath := filepath.Join(root, "sango.yaml")
	data := []byte(`name: symlink-app
services:
  web:
    type: process
    command: sleep 60
    port: 3000
`)
	if err := os.WriteFile(configPath, data, 0o644); err != nil {
		t.Fatalf("write root sango.yaml: %v", err)
	}
	resolvedConfigPath, err := filepath.EvalSymlinks(configPath)
	if err != nil {
		t.Fatalf("resolve root sango.yaml: %v", err)
	}
	resolvedRoot := filepath.Dir(resolvedConfigPath)

	ws := &worktree.WorktreeState{
		Active: "feature/login",
		Worktrees: map[string]*worktree.WorktreeInfo{
			"feature/login": {
				Offset:   200,
				Services: []string{"web"},
			},
		},
		SharedServices: map[string]*worktree.SharedService{},
	}
	if err := ws.Save(filepath.Join(root, ".sango")); err != nil {
		t.Fatalf("save worktrees.json: %v", err)
	}

	worktreeDir := filepath.Join(root, "worktrees", "feature", "login")
	if err := os.MkdirAll(worktreeDir, 0o755); err != nil {
		t.Fatalf("mkdir worktree: %v", err)
	}
	if err := os.Symlink(configPath, filepath.Join(worktreeDir, "sango.yaml")); err != nil {
		t.Fatalf("symlink sango.yaml: %v", err)
	}
	chdirForTest(t, worktreeDir)

	ctx, err := findProjectContext(worktreeDir)
	if err != nil {
		t.Fatalf("findProjectContext: %v", err)
	}
	if ctx.Root != resolvedRoot {
		t.Fatalf("Root = %q, want %q", ctx.Root, resolvedRoot)
	}
	if ctx.ConfigPath != resolvedConfigPath {
		t.Fatalf("ConfigPath = %q, want %q", ctx.ConfigPath, resolvedConfigPath)
	}

	cfg, err := loadConfig()
	if err != nil {
		t.Fatalf("loadConfig: %v", err)
	}
	orch, err := service.NewOrchestratorWithWorktree(cfg, cfgFile, service.OrchestratorOptions{})
	if err != nil {
		t.Fatalf("NewOrchestratorWithWorktree: %v", err)
	}
	ports := orch.ResolveServicePorts()
	if ports["web"] != 3200 {
		t.Fatalf("web port = %d, want 3200", ports["web"])
	}
}

func TestResolveProjectContextConfigTakesPrecedenceOverRoot(t *testing.T) {
	resetCommandGlobals(t)

	rootA := t.TempDir()
	rootB := t.TempDir()
	writeTestSangoYAML(t, rootA, "app-a")
	writeTestSangoYAML(t, rootB, "app-b")

	cfgFile = filepath.Join(rootA, "sango.yaml")
	rootDir = rootB

	ctx, err := resolveProjectContext("")
	if err != nil {
		t.Fatalf("resolveProjectContext: %v", err)
	}
	if ctx.Root != rootA || ctx.Name != "app-a" {
		t.Fatalf("ctx = %#v, want config root app-a", ctx)
	}
}

func TestBuildSnapshotOutputIncludesStableIDsAndRuntimeShape(t *testing.T) {
	resetClock(t)
	resetPortHolder(t)

	root := t.TempDir()
	sangoDir := filepath.Join(root, ".sango")
	cfg := dashboardTestConfig()
	ws := &worktree.WorktreeState{
		Active: "auth-refactor",
		Worktrees: map[string]*worktree.WorktreeInfo{
			"auth-refactor": {
				Offset:   100,
				Services: []string{"api", "repo"},
			},
		},
		SharedServices: map[string]*worktree.SharedService{},
	}

	out := buildSnapshotOutput(root, sangoDir, cfg, ws)

	if out.SchemaVersion != 1 {
		t.Fatalf("SchemaVersion = %d, want 1", out.SchemaVersion)
	}
	if out.GeneratedAt != "2026-05-06T12:34:56Z" {
		t.Fatalf("GeneratedAt = %q", out.GeneratedAt)
	}
	if out.Project.ActiveWorktreeSet != "auth-refactor" {
		t.Fatalf("ActiveWorktreeSet = %q", out.Project.ActiveWorktreeSet)
	}
	if len(out.Repos) != 1 || out.Repos[0].ID != "repo" {
		t.Fatalf("Repos = %#v, want repo", out.Repos)
	}
	if len(out.WorktreeSets) != 1 || out.WorktreeSets[0].RepoWorktrees[0].ID != "auth-refactor:repo" {
		t.Fatalf("WorktreeSets = %#v", out.WorktreeSets)
	}

	instances := map[string]snapshotServiceInstance{}
	for _, instance := range out.ServiceInstances {
		instances[instance.ID] = instance
	}
	api := instances["auth-refactor:api"]
	if api.ServiceID != "api" || api.RepoID != "repo" || api.WorktreeSetID != "auth-refactor" {
		t.Fatalf("api instance = %#v", api)
	}
	if api.Status != "stopped" {
		t.Fatalf("api status = %q, want stopped", api.Status)
	}
	if len(api.Ports) != 1 || api.Ports[0].Actual != 3100 || api.Ports[0].URL != "http://localhost:3100" {
		t.Fatalf("api ports = %#v", api.Ports)
	}
	if api.Health.Status != "unknown" || api.Health.URL != "http://localhost:3100/health" {
		t.Fatalf("api health = %#v", api.Health)
	}

	db := instances["shared:db"]
	if !db.Shared || db.WorktreeSetID != "shared" || db.Ports[0].Actual != 5432 {
		t.Fatalf("shared db instance = %#v", db)
	}
}

func TestBuildWorktreeStatusOutputIncludesMatrixShape(t *testing.T) {
	resetClock(t)
	resetPortHolder(t)

	root := t.TempDir()
	sangoDir := filepath.Join(root, ".sango")
	cfg := dashboardTestConfig()
	ws := &worktree.WorktreeState{
		Active: "auth-refactor",
		Worktrees: map[string]*worktree.WorktreeInfo{
			"auth-refactor": {
				Offset:   100,
				Services: []string{"api", "repo"},
			},
		},
		SharedServices: map[string]*worktree.SharedService{},
	}

	out := buildWorktreeStatusOutput(sangoDir, ws, cfg)

	if out.ProjectRoot != root {
		t.Fatalf("ProjectRoot = %q, want %q", out.ProjectRoot, root)
	}
	if out.ActiveWorktreeSet != "auth-refactor" {
		t.Fatalf("ActiveWorktreeSet = %q", out.ActiveWorktreeSet)
	}
	if len(out.Repos) != 1 || out.Repos[0].ID != "repo" {
		t.Fatalf("Repos = %#v", out.Repos)
	}
	repoWT := out.WorktreeSets[0].RepoWorktrees["repo"]
	if repoWT.ID != "auth-refactor:repo" || repoWT.Exists {
		t.Fatalf("repo worktree = %#v", repoWT)
	}
}

func TestLogJSONCursorAndOutputShape(t *testing.T) {
	t1 := time.Date(2026, 5, 6, 12, 33, 21, 0, time.UTC)
	t2 := t1.Add(time.Second)
	entries := []sangoLog.LogEntry{
		{Timestamp: t1, Service: "backend", Worktree: worktree.ToKey("auth-refactor"), Stream: "stderr", Level: "error", Message: "first"},
		{Timestamp: t2, Service: "backend", Worktree: worktree.ToKey("auth-refactor"), Stream: "stderr", Level: "error", Message: "second"},
	}

	output := buildLogOutputEntry(entries[0])
	if output.SchemaVersion != 1 {
		t.Fatalf("SchemaVersion = %d", output.SchemaVersion)
	}
	if output.ServiceID != "backend" || output.ServiceInstanceID != "auth-refactor:backend" {
		t.Fatalf("output = %#v", output)
	}
	if output.Cursor == "" {
		t.Fatalf("Cursor should be populated")
	}

	filtered := applyLogCursor(entries, output.Cursor)
	if len(filtered) != 1 || filtered[0].Message != "second" {
		t.Fatalf("filtered = %#v, want only second", filtered)
	}

	parsed, err := parseLogSince("2026-05-06T12:00:00Z")
	if err != nil {
		t.Fatalf("parseLogSince: %v", err)
	}
	if !parsed.Equal(time.Date(2026, 5, 6, 12, 0, 0, 0, time.UTC)) {
		t.Fatalf("parsed = %s", parsed)
	}
}

func dashboardTestConfig() *config.Config {
	return &config.Config{
		Name: "my-product",
		Services: map[string]*config.Service{
			"repo": {
				Type:          "process",
				Repo:          "git@example.com:org/repo.git",
				DefaultBranch: "main",
			},
			"api": {
				Type:      "process",
				RepoName:  "repo",
				Command:   "go run .",
				Port:      3000,
				OpenURL:   "http://localhost:${port}",
				DependsOn: []string{"db"},
				Healthcheck: &config.Healthcheck{
					URL: "http://localhost:${port}/health",
				},
			},
			"db": {
				Type:   "docker",
				Image:  "postgres:16",
				Shared: true,
				Port:   5432,
			},
		},
		Worktree: config.WorktreeConfig{BaseDir: "worktrees", DefaultBranch: "main"},
	}
}

func resetClock(t *testing.T) {
	originalNow := nowUTC
	nowUTC = func() time.Time {
		return time.Date(2026, 5, 6, 12, 34, 56, 0, time.UTC)
	}
	t.Cleanup(func() {
		nowUTC = originalNow
	})
}

func resetPortHolder(t *testing.T) {
	original := getPortHolder
	getPortHolder = func(int) (int, error) {
		return 0, os.ErrNotExist
	}
	t.Cleanup(func() {
		getPortHolder = original
	})
}

func resetCommandGlobals(t *testing.T) {
	originalCfgFile := cfgFile
	originalRootDir := rootDir
	originalWorktreeFlag := worktreeFlag
	t.Cleanup(func() {
		cfgFile = originalCfgFile
		rootDir = originalRootDir
		worktreeFlag = originalWorktreeFlag
	})
	cfgFile = ""
	rootDir = ""
	worktreeFlag = ""
}

func writeTestSangoYAML(t *testing.T, root, name string) {
	t.Helper()
	data := []byte("name: " + name + "\nservices: {}\n")
	if err := os.WriteFile(filepath.Join(root, "sango.yaml"), data, 0o644); err != nil {
		t.Fatalf("write sango.yaml: %v", err)
	}
}

func chdirForTest(t *testing.T, dir string) {
	t.Helper()
	previous, err := os.Getwd()
	if err != nil {
		t.Fatalf("getwd: %v", err)
	}
	if err := os.Chdir(dir); err != nil {
		t.Fatalf("chdir %s: %v", dir, err)
	}
	t.Cleanup(func() {
		if err := os.Chdir(previous); err != nil {
			t.Fatalf("restore cwd %s: %v", previous, err)
		}
	})
}
