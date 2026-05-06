package cmd

import (
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"sort"
	"strings"

	"github.com/ryugen04/sango/internal/config"
	"github.com/ryugen04/sango/internal/port"
	"github.com/ryugen04/sango/internal/process"
	"github.com/ryugen04/sango/internal/worktree"
	"github.com/spf13/cobra"
)

var snapshotJSON bool

var getPortHolder = port.GetPortHolder

type snapshotOutput struct {
	machineMeta
	Project          snapshotProject           `json:"project"`
	Repos            []snapshotRepo            `json:"repos"`
	Services         []snapshotService         `json:"services"`
	WorktreeSets     []snapshotWorktreeSet     `json:"worktree_sets"`
	ServiceInstances []snapshotServiceInstance `json:"service_instances"`
}

type snapshotProject struct {
	Name              string `json:"name"`
	Root              string `json:"root"`
	ActiveWorktreeSet string `json:"active_worktree_set"`
}

type snapshotRepo struct {
	ID            string   `json:"id"`
	Path          string   `json:"path"`
	DefaultBranch string   `json:"default_branch"`
	Services      []string `json:"services"`
}

type snapshotService struct {
	ID        string   `json:"id"`
	RepoID    string   `json:"repo_id,omitempty"`
	Type      string   `json:"type"`
	Shared    bool     `json:"shared"`
	PortBase  int      `json:"port_base,omitempty"`
	DependsOn []string `json:"depends_on,omitempty"`
}

type snapshotWorktreeSet struct {
	ID            string                 `json:"id"`
	Active        bool                   `json:"active"`
	RepoWorktrees []snapshotRepoWorktree `json:"repo_worktrees"`
}

type snapshotRepoWorktree struct {
	ID            string               `json:"id"`
	RepoID        string               `json:"repo_id"`
	WorktreeSetID string               `json:"worktree_set_id"`
	Path          string               `json:"path"`
	Branch        string               `json:"branch,omitempty"`
	Head          string               `json:"head,omitempty"`
	Exists        bool                 `json:"exists"`
	Dirty         snapshotDirtySummary `json:"dirty"`
}

type snapshotDirtySummary struct {
	Files     int `json:"files"`
	Staged    int `json:"staged"`
	Unstaged  int `json:"unstaged"`
	Untracked int `json:"untracked"`
}

type snapshotServiceInstance struct {
	ID            string         `json:"id"`
	ServiceID     string         `json:"service_id"`
	RepoID        string         `json:"repo_id,omitempty"`
	WorktreeSetID string         `json:"worktree_set_id"`
	Type          string         `json:"type"`
	Shared        bool           `json:"shared"`
	Status        string         `json:"status"`
	Health        snapshotHealth `json:"health"`
	PID           int            `json:"pid,omitempty"`
	Ports         []snapshotPort `json:"ports,omitempty"`
	DependsOn     []string       `json:"depends_on,omitempty"`
	RestartCount  int            `json:"restart_count,omitempty"`
	PortListening *bool          `json:"port_listening,omitempty"`
	ProcessAlive  *bool          `json:"process_alive,omitempty"`
	VerifiedAt    string         `json:"verified_at,omitempty"`
}

type snapshotPort struct {
	Name   string `json:"name"`
	Base   int    `json:"base"`
	Actual int    `json:"actual"`
	URL    string `json:"url,omitempty"`
	Open   bool   `json:"open"`
}

type snapshotHealth struct {
	Status    string `json:"status"`
	CheckedAt string `json:"checked_at,omitempty"`
	URL       string `json:"url,omitempty"`
	LastError string `json:"last_error,omitempty"`
}

var snapshotCmd = &cobra.Command{
	Use:   "snapshot",
	Short: "dashboard 向けのプロジェクト状態スナップショットを出力する",
	RunE: func(cmd *cobra.Command, args []string) error {
		if !snapshotJSON {
			return fmt.Errorf("snapshot は --json を指定してください")
		}
		cfg, err := loadConfig()
		if err != nil {
			return err
		}
		projectRoot, err := currentProjectRoot()
		if err != nil {
			return err
		}
		sangoDir := filepath.Join(projectRoot, ".sango")
		ws, err := worktree.Load(sangoDir)
		if err != nil {
			return fmt.Errorf("worktrees.jsonの読み込みに失敗: %w", err)
		}
		return writeJSON(cmd.OutOrStdout(), buildSnapshotOutput(projectRoot, sangoDir, cfg, ws))
	},
}

func buildSnapshotOutput(projectRoot, sangoDir string, cfg *config.Config, ws *worktree.WorktreeState) snapshotOutput {
	if ws == nil {
		ws = &worktree.WorktreeState{Worktrees: map[string]*worktree.WorktreeInfo{}, SharedServices: map[string]*worktree.SharedService{}}
	}
	active := ws.Active
	if active == "" {
		active = "main"
	}

	repos := collectSnapshotRepos(projectRoot, cfg)
	out := snapshotOutput{
		machineMeta: newMachineMeta(projectRoot),
		Project: snapshotProject{
			Name:              cfg.Name,
			Root:              projectRoot,
			ActiveWorktreeSet: active,
		},
		Repos:            repos,
		Services:         collectSnapshotServices(cfg),
		WorktreeSets:     []snapshotWorktreeSet{},
		ServiceInstances: []snapshotServiceInstance{},
	}

	wtNames := make([]string, 0, len(ws.Worktrees))
	for name := range ws.Worktrees {
		wtNames = append(wtNames, name)
	}
	sort.Strings(wtNames)

	for _, wtName := range wtNames {
		wt := ws.Worktrees[wtName]
		set := snapshotWorktreeSet{
			ID:     wtName,
			Active: wtName == active,
		}
		for _, repo := range repos {
			path := repoWorktreePath(projectRoot, cfg, wtName, repo.ID)
			exists, branch, head, dirty := inspectRepoWorktree(path)
			set.RepoWorktrees = append(set.RepoWorktrees, snapshotRepoWorktree{
				ID:            repoWorktreeID(wtName, repo.ID),
				RepoID:        repo.ID,
				WorktreeSetID: wtName,
				Path:          path,
				Branch:        branch,
				Head:          head,
				Exists:        exists,
				Dirty:         dirty,
			})
		}
		out.WorktreeSets = append(out.WorktreeSets, set)

		serviceIDs := sortedServiceIDs(wt.Services)
		for _, serviceID := range serviceIDs {
			svc := cfg.Services[serviceID]
			if svc == nil || svc.Shared || isRepoOnlyService(svc) {
				continue
			}
			out.ServiceInstances = append(out.ServiceInstances, buildServiceInstance(sangoDir, wtName, worktree.ToKey(wtName), wt.Offset, serviceID, svc))
		}
	}

	serviceIDs := make([]string, 0, len(cfg.Services))
	for serviceID, svc := range cfg.Services {
		if svc != nil && svc.Shared && !isRepoOnlyService(svc) {
			serviceIDs = append(serviceIDs, serviceID)
		}
	}
	sort.Strings(serviceIDs)
	for _, serviceID := range serviceIDs {
		svc := cfg.Services[serviceID]
		out.ServiceInstances = append(out.ServiceInstances, buildServiceInstance(sangoDir, "shared", "shared", 0, serviceID, svc))
	}

	sort.Slice(out.ServiceInstances, func(i, j int) bool {
		return out.ServiceInstances[i].ID < out.ServiceInstances[j].ID
	})
	return out
}

func collectSnapshotRepos(projectRoot string, cfg *config.Config) []snapshotRepo {
	repoMap := make(map[string]*snapshotRepo)

	serviceIDs := make([]string, 0, len(cfg.Services))
	for serviceID := range cfg.Services {
		serviceIDs = append(serviceIDs, serviceID)
	}
	sort.Strings(serviceIDs)

	for _, serviceID := range serviceIDs {
		svc := cfg.Services[serviceID]
		repoID := serviceRepoID(serviceID, svc)
		if repoID == "" {
			continue
		}
		repo := repoMap[repoID]
		if repo == nil {
			repoSvc := cfg.Services[repoID]
			if repoSvc == nil {
				repoSvc = svc
			}
			repo = &snapshotRepo{
				ID:            repoID,
				Path:          repoPath(projectRoot, repoID, repoSvc),
				DefaultBranch: defaultBranch(cfg, repoSvc),
			}
			repoMap[repoID] = repo
		}
		repo.Services = append(repo.Services, serviceID)
	}

	repoIDs := make([]string, 0, len(repoMap))
	for repoID := range repoMap {
		repoIDs = append(repoIDs, repoID)
	}
	sort.Strings(repoIDs)

	repos := make([]snapshotRepo, 0, len(repoIDs))
	for _, repoID := range repoIDs {
		repo := repoMap[repoID]
		sort.Strings(repo.Services)
		repos = append(repos, *repo)
	}
	return repos
}

func collectSnapshotServices(cfg *config.Config) []snapshotService {
	serviceIDs := make([]string, 0, len(cfg.Services))
	for serviceID := range cfg.Services {
		serviceIDs = append(serviceIDs, serviceID)
	}
	sort.Strings(serviceIDs)

	services := make([]snapshotService, 0, len(serviceIDs))
	for _, serviceID := range serviceIDs {
		svc := cfg.Services[serviceID]
		services = append(services, snapshotService{
			ID:        serviceID,
			RepoID:    serviceRepoID(serviceID, svc),
			Type:      svc.Type,
			Shared:    svc.Shared,
			PortBase:  svc.Port,
			DependsOn: sortedServiceIDs(svc.DependsOn),
		})
	}
	return services
}

func buildServiceInstance(sangoDir string, worktreeSetID, pidWorktree string, offset int, serviceID string, svc *config.Service) snapshotServiceInstance {
	resolvedPort := 0
	portOpen := false
	if svc.Port > 0 {
		resolvedPort = port.ResolvePort(svc.Port, offset, svc.Shared)
		portOpen = isPortOpen(resolvedPort)
	}

	pid := 0
	if p, err := process.ReadPID(sangoDir, pidWorktree, serviceID); err == nil && process.IsProcessRunning(p) {
		pid = p
	}
	status := normalizeProcessStatus(pid, portOpen)
	state := process.ReadState(sangoDir, pidWorktree, serviceID)

	instance := snapshotServiceInstance{
		ID:            serviceInstanceID(worktreeSetID, serviceID),
		ServiceID:     serviceID,
		RepoID:        serviceRepoID(serviceID, svc),
		WorktreeSetID: worktreeSetID,
		Type:          svc.Type,
		Shared:        svc.Shared,
		Status:        status,
		Health:        buildSnapshotHealth(svc, state, status, resolvedPort),
		PID:           pid,
		DependsOn:     sortedServiceIDs(svc.DependsOn),
		RestartCount:  state.RestartCount,
		PortListening: state.PortListening,
		ProcessAlive:  state.ProcessAlive,
		VerifiedAt:    state.VerifiedAt,
	}
	if svc.Port > 0 {
		instance.Ports = []snapshotPort{{
			Name:   "default",
			Base:   svc.Port,
			Actual: resolvedPort,
			URL:    serviceURL(svc, resolvedPort),
			Open:   portOpen,
		}}
	}
	return instance
}

func serviceRepoID(serviceID string, svc *config.Service) string {
	if svc == nil {
		return ""
	}
	if svc.RepoName != "" {
		return svc.RepoName
	}
	if svc.Repo != "" || svc.RepoPath != "" || svc.WorkingDir != "" {
		return serviceID
	}
	return ""
}

func isRepoOnlyService(svc *config.Service) bool {
	return svc != nil && svc.Repo != "" && svc.Command == ""
}

func repoPath(projectRoot, repoID string, svc *config.Service) string {
	switch {
	case svc != nil && svc.RepoPath != "":
		return absUnderRoot(projectRoot, svc.RepoPath)
	case svc != nil && svc.WorkingDir != "":
		return absUnderRoot(projectRoot, svc.WorkingDir)
	default:
		return filepath.Join(projectRoot, repoID)
	}
}

func repoWorktreePath(projectRoot string, cfg *config.Config, worktreeSetID, repoID string) string {
	base := cfg.Worktree.WorktreeDir(worktreeSetID)
	if !filepath.IsAbs(base) {
		base = filepath.Join(projectRoot, base)
	}
	return filepath.Join(base, repoID)
}

func absUnderRoot(projectRoot, path string) string {
	if filepath.IsAbs(path) {
		return path
	}
	return filepath.Join(projectRoot, path)
}

func defaultBranch(cfg *config.Config, svc *config.Service) string {
	if svc != nil && svc.DefaultBranch != "" {
		return svc.DefaultBranch
	}
	if cfg.Worktree.DefaultBranch != "" {
		return cfg.Worktree.DefaultBranch
	}
	return "main"
}

func serviceInstanceID(worktreeSetID, serviceID string) string {
	return worktreeSetID + ":" + serviceID
}

func repoWorktreeID(worktreeSetID, repoID string) string {
	return worktreeSetID + ":" + repoID
}

func normalizeProcessStatus(pid int, portOpen bool) string {
	if pid > 0 || portOpen {
		return "running"
	}
	return "stopped"
}

func buildSnapshotHealth(svc *config.Service, state *process.ServiceState, status string, resolvedPort int) snapshotHealth {
	health := snapshotHealth{Status: "unchecked"}
	if svc.Healthcheck == nil {
		return health
	}

	health.Status = "unknown"
	health.CheckedAt = state.VerifiedAt
	health.URL = resolveHealthURL(svc.Healthcheck.URL, resolvedPort)
	switch state.HealthStatus {
	case "healthy", "ok":
		health.Status = "ok"
	case "unhealthy", "failed", "fail":
		health.Status = "fail"
	case "warn", "warning":
		health.Status = "warn"
	default:
		if status == "stopped" {
			health.Status = "unknown"
		}
	}
	return health
}

func resolveHealthURL(raw string, portNumber int) string {
	if raw == "" {
		return ""
	}
	return strings.ReplaceAll(raw, "${port}", fmt.Sprintf("%d", portNumber))
}

func serviceURL(svc *config.Service, resolvedPort int) string {
	if svc.OpenURL != "" {
		return strings.ReplaceAll(svc.OpenURL, "${port}", fmt.Sprintf("%d", resolvedPort))
	}
	if resolvedPort > 0 {
		return fmt.Sprintf("http://localhost:%d", resolvedPort)
	}
	return ""
}

func isPortOpen(portNumber int) bool {
	if portNumber <= 0 {
		return false
	}
	pid, err := getPortHolder(portNumber)
	return err == nil && pid > 0
}

func inspectRepoWorktree(path string) (bool, string, string, snapshotDirtySummary) {
	info, err := os.Stat(path)
	if err != nil || !info.IsDir() {
		return false, "", "", snapshotDirtySummary{}
	}

	branch := strings.TrimSpace(runGit(path, "rev-parse", "--abbrev-ref", "HEAD"))
	head := strings.TrimSpace(runGit(path, "rev-parse", "--short", "HEAD"))
	dirty := parseDirtySummary(runGit(path, "status", "--porcelain"))
	return true, branch, head, dirty
}

func runGit(dir string, args ...string) string {
	cmd := exec.Command("git", args...)
	cmd.Dir = dir
	out, err := cmd.Output()
	if err != nil {
		return ""
	}
	return string(out)
}

func parseDirtySummary(status string) snapshotDirtySummary {
	lines := strings.Split(strings.TrimSpace(status), "\n")
	if len(lines) == 1 && strings.TrimSpace(lines[0]) == "" {
		return snapshotDirtySummary{}
	}

	var dirty snapshotDirtySummary
	for _, line := range lines {
		if line == "" {
			continue
		}
		dirty.Files++
		if strings.HasPrefix(line, "??") {
			dirty.Untracked++
			continue
		}
		if len(line) > 0 && line[0] != ' ' {
			dirty.Staged++
		}
		if len(line) > 1 && line[1] != ' ' {
			dirty.Unstaged++
		}
	}
	return dirty
}

func sortedServiceIDs(ids []string) []string {
	out := append([]string(nil), ids...)
	sort.Strings(out)
	return out
}

func init() {
	snapshotCmd.Flags().BoolVar(&snapshotJSON, "json", false, "JSONで出力する")
	rootCmd.AddCommand(snapshotCmd)
}
