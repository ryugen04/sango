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
