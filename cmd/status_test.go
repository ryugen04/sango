package cmd

import (
	"testing"

	"github.com/ryugen04/sango/internal/service"
)

func TestBuildStatusOutputSortsAndExtractsSharedServices(t *testing.T) {
	listening := true
	result := &service.StatusResult{
		Worktree: "main",
		Services: []service.ServiceInfo{
			{Name: "web", Status: "running"},
			{Name: "db", Status: "running", IsShared: true},
			{Name: "api", Status: "stopped", PortListening: &listening},
		},
		Worktrees: []service.WorktreeInfo{
			{
				Name: "feature-b",
				Services: []service.ServiceInfo{
					{Name: "web"},
					{Name: "api"},
				},
			},
			{
				Name: "feature-a",
				Services: []service.ServiceInfo{
					{Name: "web"},
					{Name: "api"},
				},
			},
		},
	}

	output := buildStatusOutput(result)

	if output.Worktree != "main" {
		t.Fatalf("Worktree = %q, want main", output.Worktree)
	}
	if len(output.SharedServices) != 1 || output.SharedServices[0].Name != "db" {
		t.Fatalf("SharedServices = %#v, want db only", output.SharedServices)
	}
	if output.Services[0].Name != "api" || output.Services[1].Name != "db" || output.Services[2].Name != "web" {
		t.Fatalf("Services should be sorted by name: %#v", output.Services)
	}
	if output.Worktrees[0].Name != "feature-a" || output.Worktrees[1].Name != "feature-b" {
		t.Fatalf("Worktrees should be sorted by name: %#v", output.Worktrees)
	}
}
