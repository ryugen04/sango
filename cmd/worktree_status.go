package cmd

import (
	"fmt"
	"path/filepath"
	"sort"
	"strconv"

	"github.com/charmbracelet/lipgloss"
	"github.com/charmbracelet/lipgloss/table"
	"github.com/ryugen04/sango/internal/config"
	"github.com/ryugen04/sango/internal/process"
	"github.com/ryugen04/sango/internal/worktree"
	"github.com/spf13/cobra"
)

var worktreeStatusJSON bool

type worktreeStatusService struct {
	Name   string `json:"name"`
	Port   int    `json:"port,omitempty"`
	Status string `json:"status"`
	PID    int    `json:"pid,omitempty"`
	Shared bool   `json:"shared,omitempty"`
}

type worktreeStatusEntry struct {
	Name     string                  `json:"name"`
	Offset   int                     `json:"offset"`
	Active   bool                    `json:"active"`
	Services []worktreeStatusService `json:"services,omitempty"`
}

type worktreeSharedStatusEntry struct {
	Name   string `json:"name"`
	Port   int    `json:"port"`
	Status string `json:"status"`
	PID    int    `json:"pid,omitempty"`
}

type worktreeStatusRepo struct {
	ID   string `json:"id"`
	Path string `json:"path"`
}

type worktreeStatusRepoWorktree struct {
	ID         string `json:"id"`
	Exists     bool   `json:"exists"`
	Path       string `json:"path"`
	Branch     string `json:"branch,omitempty"`
	Head       string `json:"head,omitempty"`
	DirtyFiles int    `json:"dirty_files"`
}

type worktreeStatusSet struct {
	ID            string                                `json:"id"`
	Active        bool                                  `json:"active"`
	RepoWorktrees map[string]worktreeStatusRepoWorktree `json:"repo_worktrees"`
}

type worktreeStatusOutput struct {
	machineMeta
	Active            string                      `json:"active"`
	ActiveWorktreeSet string                      `json:"active_worktree_set"`
	Repos             []worktreeStatusRepo        `json:"repos"`
	WorktreeSets      []worktreeStatusSet         `json:"worktree_sets"`
	Worktrees         []worktreeStatusEntry       `json:"worktrees"`
	SharedServices    []worktreeSharedStatusEntry `json:"shared_services,omitempty"`
}

var worktreeStatusCmd = &cobra.Command{
	Use:   "status",
	Short: "全ワークツリーの状態を横断表示する",
	RunE: func(cmd *cobra.Command, args []string) error {
		cfg, err := loadConfig()
		if err != nil {
			return err
		}
		sangoDir, err := currentSangoDir()
		if err != nil {
			return err
		}
		ws, err := worktree.Load(sangoDir)
		if err != nil {
			return fmt.Errorf("worktrees.jsonの読み込みに失敗: %w", err)
		}

		output := buildWorktreeStatusOutput(sangoDir, ws, cfg)
		if worktreeStatusJSON {
			return writeJSON(cmd.OutOrStdout(), output)
		}

		green := lipgloss.NewStyle().Foreground(lipgloss.Color("2"))
		gray := lipgloss.NewStyle().Foreground(lipgloss.Color("8"))
		bold := lipgloss.NewStyle().Bold(true)
		dim := lipgloss.NewStyle().Foreground(lipgloss.Color("8"))

		headerStyle := lipgloss.NewStyle().Bold(true).Padding(0, 1)
		cellStyle := lipgloss.NewStyle().Padding(0, 1)

		// worktree名をソート
		wtNames := make([]string, 0, len(ws.Worktrees))
		for name := range ws.Worktrees {
			wtNames = append(wtNames, name)
		}
		sort.Strings(wtNames)

		for _, wtName := range wtNames {
			wt := ws.Worktrees[wtName]
			wtKey := worktree.ToKey(wtName)

			// ヘッダー
			header := wtName
			if wtName == ws.Active {
				header = "* " + green.Render(wtName) + " (active)"
			}
			fmt.Printf("%s  %s\n", bold.Render(header), dim.Render(fmt.Sprintf("offset:%d", wt.Offset)))

			// サーバーサービスのみ抽出
			serviceNames := make([]string, 0)
			for _, name := range wt.Services {
				svc := cfg.Services[name]
				if svc == nil {
					continue
				}
				// repo-onlyサービスをスキップ
				if svc.Repo != "" && svc.Command == "" {
					continue
				}
				serviceNames = append(serviceNames, name)
			}
			sort.Strings(serviceNames)

			if len(serviceNames) == 0 {
				fmt.Println()
				continue
			}

			rows := make([][]string, 0, len(serviceNames))
			for _, name := range serviceNames {
				svc := cfg.Services[name]

				status := "stopped"
				pidStr := dim.Render("-")

				pidWorktree := wtKey
				if svc.Shared {
					pidWorktree = "shared"
				}

				if pid, err := process.ReadPID(sangoDir, pidWorktree, name); err == nil {
					if process.IsProcessRunning(pid) {
						status = "running"
						pidStr = strconv.Itoa(pid)
					}
				}

				var styledStatus string
				if status == "running" {
					styledStatus = green.Render(status)
				} else {
					styledStatus = gray.Render(status)
				}

				portStr := dim.Render("-")
				if svc.Port > 0 {
					resolvedPort := svc.Port
					if !svc.Shared {
						resolvedPort += wt.Offset
					}
					portStr = strconv.Itoa(resolvedPort)
				}

				rows = append(rows, []string{name, portStr, styledStatus, pidStr})
			}

			t := table.New().
				Border(lipgloss.NormalBorder()).
				BorderStyle(dim).
				BorderRow(false).
				BorderColumn(false).
				BorderLeft(false).
				BorderRight(false).
				BorderTop(false).
				BorderBottom(false).
				BorderHeader(true).
				StyleFunc(func(row, col int) lipgloss.Style {
					if row == table.HeaderRow {
						return headerStyle
					}
					return cellStyle
				}).
				Headers("SERVICE", "PORT", "STATUS", "PID").
				Rows(rows...)

			fmt.Println(t)
			fmt.Println()
		}

		// sharedサービス表示
		if len(ws.SharedServices) > 0 {
			fmt.Println(bold.Render("Shared Services"))
			rows := make([][]string, 0, len(ws.SharedServices))
			for name, ss := range ws.SharedServices {
				status := "stopped"
				pidStr := dim.Render("-")
				if pid, err := process.ReadPID(sangoDir, "shared", name); err == nil {
					if process.IsProcessRunning(pid) {
						status = "running"
						pidStr = strconv.Itoa(pid)
					}
				}

				var styledStatus string
				if status == "running" {
					styledStatus = green.Render(status)
				} else {
					styledStatus = gray.Render(status)
				}

				rows = append(rows, []string{name, strconv.Itoa(ss.Port), styledStatus, pidStr})
			}

			t := table.New().
				Border(lipgloss.NormalBorder()).
				BorderStyle(dim).
				BorderRow(false).
				BorderColumn(false).
				BorderLeft(false).
				BorderRight(false).
				BorderTop(false).
				BorderBottom(false).
				BorderHeader(true).
				StyleFunc(func(row, col int) lipgloss.Style {
					if row == table.HeaderRow {
						return lipgloss.NewStyle().Bold(true).Padding(0, 1)
					}
					return lipgloss.NewStyle().Padding(0, 1)
				}).
				Headers("SERVICE", "PORT", "STATUS", "PID").
				Rows(rows...)

			fmt.Println(t)
		}

		return nil
	},
}

func buildWorktreeStatusOutput(sangoDir string, ws *worktree.WorktreeState, cfg *config.Config) worktreeStatusOutput {
	projectRoot := filepath.Dir(sangoDir)
	active := ws.Active
	if active == "" {
		active = "main"
	}
	output := worktreeStatusOutput{
		machineMeta:       newMachineMeta(projectRoot),
		Active:            active,
		ActiveWorktreeSet: active,
		Repos:             []worktreeStatusRepo{},
		WorktreeSets:      []worktreeStatusSet{},
		Worktrees:         make([]worktreeStatusEntry, 0, len(ws.Worktrees)),
	}

	repos := collectSnapshotRepos(projectRoot, cfg)
	for _, repo := range repos {
		output.Repos = append(output.Repos, worktreeStatusRepo{ID: repo.ID, Path: repo.Path})
	}

	wtNames := make([]string, 0, len(ws.Worktrees))
	for name := range ws.Worktrees {
		wtNames = append(wtNames, name)
	}
	sort.Strings(wtNames)

	for _, wtName := range wtNames {
		wt := ws.Worktrees[wtName]
		entry := worktreeStatusEntry{
			Name:   wtName,
			Offset: wt.Offset,
			Active: wtName == active,
		}
		matrixSet := worktreeStatusSet{
			ID:            wtName,
			Active:        wtName == active,
			RepoWorktrees: make(map[string]worktreeStatusRepoWorktree, len(repos)),
		}
		for _, repo := range repos {
			path := repoWorktreePath(projectRoot, cfg, wtName, repo.ID)
			exists, branch, head, dirty := inspectRepoWorktree(path)
			matrixSet.RepoWorktrees[repo.ID] = worktreeStatusRepoWorktree{
				ID:         repoWorktreeID(wtName, repo.ID),
				Exists:     exists,
				Path:       path,
				Branch:     branch,
				Head:       head,
				DirtyFiles: dirty.Files,
			}
		}
		output.WorktreeSets = append(output.WorktreeSets, matrixSet)

		serviceNames := make([]string, 0)
		for _, name := range wt.Services {
			svc := cfg.Services[name]
			if svc == nil {
				continue
			}
			if svc.Repo != "" && svc.Command == "" {
				continue
			}
			serviceNames = append(serviceNames, name)
		}
		sort.Strings(serviceNames)

		wtKey := worktree.ToKey(wtName)
		for _, name := range serviceNames {
			svc := cfg.Services[name]
			status := "stopped"
			pid := 0

			pidWorktree := wtKey
			if svc.Shared {
				pidWorktree = "shared"
			}
			if p, err := process.ReadPID(sangoDir, pidWorktree, name); err == nil && process.IsProcessRunning(p) {
				status = "running"
				pid = p
			}

			resolvedPort := 0
			if svc.Port > 0 {
				resolvedPort = svc.Port
				if !svc.Shared {
					resolvedPort += wt.Offset
				}
			}

			entry.Services = append(entry.Services, worktreeStatusService{
				Name:   name,
				Port:   resolvedPort,
				Status: status,
				PID:    pid,
				Shared: svc.Shared,
			})
		}

		output.Worktrees = append(output.Worktrees, entry)
	}

	sharedNames := make([]string, 0, len(ws.SharedServices))
	for name := range ws.SharedServices {
		sharedNames = append(sharedNames, name)
	}
	sort.Strings(sharedNames)

	for _, name := range sharedNames {
		ss := ws.SharedServices[name]
		status := "stopped"
		pid := 0
		if p, err := process.ReadPID(sangoDir, "shared", name); err == nil && process.IsProcessRunning(p) {
			status = "running"
			pid = p
		}
		output.SharedServices = append(output.SharedServices, worktreeSharedStatusEntry{
			Name:   name,
			Port:   ss.Port,
			Status: status,
			PID:    pid,
		})
	}

	return output
}

func init() {
	worktreeStatusCmd.Flags().BoolVar(&worktreeStatusJSON, "json", false, "JSONで出力する")
	worktreeCmd.AddCommand(worktreeStatusCmd)
}
