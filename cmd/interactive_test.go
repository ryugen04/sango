package cmd

import (
	"testing"

	"github.com/ryugen04/sango/internal/config"
	"github.com/ryugen04/sango/internal/worktree"
)

func TestResolveWorktreeNameArg(t *testing.T) {
	originalTerminal := stdinIsTerminal
	originalPrompt := promptExistingWorktreeSelection
	t.Cleanup(func() {
		stdinIsTerminal = originalTerminal
		promptExistingWorktreeSelection = originalPrompt
	})

	ws := &worktree.WorktreeState{
		Active: "main",
		Worktrees: map[string]*worktree.WorktreeInfo{
			"main":    {},
			"feature": {},
		},
	}

	name, err := resolveWorktreeNameArg([]string{"feature"}, ws, "main", "title")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if name != "feature" {
		t.Fatalf("name = %q, want feature", name)
	}

	stdinIsTerminal = func() bool { return false }
	name, err = resolveWorktreeNameArg(nil, ws, "main", "title")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if name != "main" {
		t.Fatalf("name = %q, want main", name)
	}

	stdinIsTerminal = func() bool { return true }
	promptExistingWorktreeSelection = func(title string, names []string, active string) (string, error) {
		return "feature", nil
	}
	name, err = resolveWorktreeNameArg(nil, ws, "main", "title")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if name != "feature" {
		t.Fatalf("name = %q, want feature", name)
	}
}

func TestFilterRunnableServices(t *testing.T) {
	cfg := &config.Config{
		Services: map[string]*config.Service{
			"repo-only": {Type: "process", Repo: "git@example.com/repo.git"},
			"api":       {Type: "process", Command: "go run ."},
			"db":        {Type: "docker", Image: "postgres:16"},
		},
	}

	got := filterRunnableServices(cfg, []string{"repo-only", "api", "db", "api"})
	if len(got) != 2 {
		t.Fatalf("len = %d, want 2 (%v)", len(got), got)
	}
	if got[0] != "api" || got[1] != "db" {
		t.Fatalf("got = %v, want [api db]", got)
	}
}

func TestResolveUpTargetsInteractiveRepos(t *testing.T) {
	originalPrompt := promptUpSelection
	originalPorts := defaultPorts
	t.Cleanup(func() {
		promptUpSelection = originalPrompt
		defaultPorts = originalPorts
	})

	cfg := &config.Config{
		Services: map[string]*config.Service{
			"repo-a": {Type: "process", Repo: "git@example.com/repo-a.git"},
			"api":    {Type: "process", RepoName: "repo-a", Command: "go run ."},
			"worker": {Type: "process", RepoName: "repo-a", Command: "go run worker"},
			"db":     {Type: "docker", Shared: true, Image: "postgres:16"},
		},
	}

	promptUpSelection = func(cfg *config.Config, currentDefaultPorts bool) (*upInteractiveSelection, error) {
		return &upInteractiveSelection{
			Mode:         upTargetModeRepos,
			Repos:        []string{"repo-a"},
			DefaultPorts: true,
		}, nil
	}

	targets, err := resolveUpTargets(cfg, nil, "", true)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(targets) != 3 {
		t.Fatalf("len = %d, want 3 (%v)", len(targets), targets)
	}
	if !defaultPorts {
		t.Fatalf("defaultPorts = false, want true")
	}
}

func TestResolveWorktreeServicesUsesDefaultServicesNonInteractive(t *testing.T) {
	originalTerminal := stdinIsTerminal
	t.Cleanup(func() {
		stdinIsTerminal = originalTerminal
	})
	stdinIsTerminal = func() bool { return false }

	cfg := &config.Config{
		Services: map[string]*config.Service{
			"repo-a": {Type: "process", Repo: "git@example.com/repo-a.git"},
			"api":    {Type: "process", RepoName: "repo-a", Command: "go run ."},
			"worker": {Type: "process", RepoName: "repo-a", Command: "go run worker"},
			"db":     {Type: "docker", Shared: true, Image: "postgres:16"},
		},
		Worktree: config.WorktreeConfig{
			Create: config.WorktreeCreateConfig{DefaultServices: []string{"api"}},
		},
	}

	got, err := resolveWorktreeServices(cfg, "")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	want := []string{"api", "db", "repo-a", "worker"}
	if len(got) != len(want) {
		t.Fatalf("got = %v, want %v", got, want)
	}
	for i := range want {
		if got[i] != want[i] {
			t.Fatalf("got = %v, want %v", got, want)
		}
	}
}

func TestDefaultSelectedRepos(t *testing.T) {
	cfg := &config.Config{
		Services: map[string]*config.Service{
			"repo-a": {Type: "process", Repo: "git@example.com/repo-a.git"},
			"api":    {Type: "process", RepoName: "repo-a", Command: "go run ."},
			"repo-b": {Type: "process", Repo: "git@example.com/repo-b.git"},
		},
		Worktree: config.WorktreeConfig{
			Create: config.WorktreeCreateConfig{DefaultServices: []string{"api", "repo-b"}},
		},
	}

	got := defaultSelectedRepos(cfg)
	if !got["repo-a"] || !got["repo-b"] || len(got) != 2 {
		t.Fatalf("defaultSelectedRepos = %#v, want repo-a and repo-b", got)
	}
}

func TestResolveWorktreeCreateBranchRemoteSelection(t *testing.T) {
	originalTerminal := stdinIsTerminal
	originalPrompt := promptCreateBranchSelection
	originalFrom := wtCreateFrom
	t.Cleanup(func() {
		stdinIsTerminal = originalTerminal
		promptCreateBranchSelection = originalPrompt
		wtCreateFrom = originalFrom
	})
	stdinIsTerminal = func() bool { return true }
	wtCreateFrom = ""
	promptCreateBranchSelection = func(remoteBranches []string) (string, error) {
		if len(remoteBranches) != 1 || remoteBranches[0] != "origin/feature/auth" {
			t.Fatalf("remoteBranches = %v, want [origin/feature/auth]", remoteBranches)
		}
		return "origin/feature/auth", nil
	}

	branch, err := resolveWorktreeCreateBranch(nil, []string{"origin/feature/auth"})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if branch != "feature/auth" {
		t.Fatalf("branch = %q, want feature/auth", branch)
	}
	if wtCreateFrom != "origin/feature/auth" {
		t.Fatalf("wtCreateFrom = %q, want origin/feature/auth", wtCreateFrom)
	}
}

func TestFuzzyFilterRemoteBranches(t *testing.T) {
	branches := []string{
		"origin/alice/feature/login",
		"origin/bob/fix/SANGO-123-worktree",
		"origin/feature/auth",
		"origin/release/2026-05",
	}

	tests := []struct {
		name  string
		query string
		want  []string
	}{
		{
			name:  "prefix after origin",
			query: "alice",
			want:  []string{"origin/alice/feature/login"},
		},
		{
			name:  "slash segment prefix",
			query: "feature",
			want:  []string{"origin/feature/auth", "origin/alice/feature/login"},
		},
		{
			name:  "case insensitive ticket",
			query: "sango-123",
			want:  []string{"origin/bob/fix/SANGO-123-worktree"},
		},
		{
			name:  "fuzzy ordered chars",
			query: "ftr",
			want:  []string{"origin/feature/auth", "origin/alice/feature/login", "origin/bob/fix/SANGO-123-worktree"},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := fuzzyFilterRemoteBranches(branches, tt.query)
			if len(got) != len(tt.want) {
				t.Fatalf("got = %v, want %v", got, tt.want)
			}
			for i := range tt.want {
				if got[i] != tt.want[i] {
					t.Fatalf("got = %v, want %v", got, tt.want)
				}
			}
		})
	}
}
