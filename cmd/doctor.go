package cmd

import (
	"fmt"
	"os/exec"
	"sort"

	"github.com/charmbracelet/lipgloss"
	"github.com/ryugen04/sango/internal/doctor"
	"github.com/ryugen04/sango/internal/service"
	"github.com/spf13/cobra"
)

var doctorFix bool
var doctorJSON bool

type doctorSummary struct {
	Passed int `json:"passed"`
	Failed int `json:"failed"`
	Warned int `json:"warned"`
}

type doctorReport struct {
	machineMeta
	Results []doctor.CheckResult `json:"results"`
	Summary doctorSummary        `json:"summary"`
}

var (
	passStyle = lipgloss.NewStyle().Foreground(lipgloss.Color("2")) // 緑
	failStyle = lipgloss.NewStyle().Foreground(lipgloss.Color("1")) // 赤
	warnStyle = lipgloss.NewStyle().Foreground(lipgloss.Color("3")) // 黄
)

var doctorCmd = &cobra.Command{
	Use:   "doctor",
	Short: "開発環境の状態をチェックする",
	RunE: func(cmd *cobra.Command, args []string) error {
		results, err := collectDoctorResults()
		if err != nil {
			return err
		}
		report := buildDoctorReport(results)

		if doctorJSON {
			projectRoot, err := currentProjectRoot()
			if err != nil {
				return err
			}
			report.machineMeta = newMachineMeta(projectRoot)
			return writeJSON(cmd.OutOrStdout(), report)
		}

		// ヘッダー出力
		fmt.Println("Sango Doctor")
		fmt.Println("============")
		fmt.Println()

		for _, r := range report.Results {
			switch r.Status {
			case doctor.StatusPass:
				fmt.Printf("%s %s - %s\n", passStyle.Render("[✓]"), r.Name, r.Message)
			case doctor.StatusFail:
				fmt.Printf("%s %s - %s\n", failStyle.Render("[✗]"), r.Name, r.Message)
				if r.Fix != "" {
					fmt.Printf("    Fix: %s\n", r.Fix)
				}
			case doctor.StatusWarn:
				fmt.Printf("%s %s - %s\n", warnStyle.Render("[!]"), r.Name, r.Message)
				if r.Fix != "" {
					fmt.Printf("    Fix: %s\n", r.Fix)
				}
			}
		}

		// サマリー出力
		fmt.Println()
		summary := fmt.Sprintf("Results: %d passed", report.Summary.Passed)
		if report.Summary.Failed > 0 {
			summary += fmt.Sprintf(", %d failed", report.Summary.Failed)
		}
		if report.Summary.Warned > 0 {
			summary += fmt.Sprintf(", %d warned", report.Summary.Warned)
		}
		fmt.Println(summary)

		// --fix オプション: failしたチェックのfixコマンドを実行
		if doctorFix {
			for _, r := range report.Results {
				if r.Status == doctor.StatusFail && r.Fix != "" {
					fmt.Printf("\n[fix] %s: %s\n", r.Name, r.Fix)
					out, err := exec.Command("sh", "-c", r.Fix).CombinedOutput()
					if err != nil {
						fmt.Printf("[fix] 失敗: %s\n", string(out))
					} else {
						fmt.Printf("[fix] 成功: %s\n", string(out))
					}
				}
			}
		}

		return nil
	},
}

func collectDoctorResults() ([]doctor.CheckResult, error) {
	cfg, err := loadConfig()
	if err != nil {
		return nil, err
	}

	checks := make([]doctor.Check, len(cfg.Doctor.Checks))
	for i, c := range cfg.Doctor.Checks {
		checks[i] = doctor.Check{
			Name:    c.Name,
			Command: c.Command,
			Expect:  c.Expect,
			Fix:     c.Fix,
		}
	}

	results := doctor.Run(checks)
	results = append(results, doctor.CheckLinuxSandbox()...)

	sangoDir, err := currentSangoDir()
	if err != nil {
		return nil, err
	}
	wtName := service.ResolveActiveWorktreeWithBaseDir(sangoDir, worktreeFlag, cfg.Worktree.ResolveBaseDir())
	orch, orchErr := service.NewOrchestratorWithWorktree(cfg, cfgFile, service.OrchestratorOptions{WorktreeFlag: worktreeFlag})
	if orchErr == nil {
		ports := orch.ResolveServicePorts()
		portResults := doctor.CheckPortConflicts(ports, sangoDir)
		sort.Slice(portResults, func(i, j int) bool {
			return portResults[i].Name < portResults[j].Name
		})
		results = append(results, portResults...)
	} else {
		results = append(results, doctor.CheckResult{
			Name:    fmt.Sprintf("ポート競合チェック (worktree: %s)", wtName),
			Status:  doctor.StatusWarn,
			Message: fmt.Sprintf("worktree情報の取得に失敗: %v", orchErr),
		})
	}

	return results, nil
}

func buildDoctorReport(results []doctor.CheckResult) doctorReport {
	report := doctorReport{Results: results}
	for _, r := range results {
		switch r.Status {
		case doctor.StatusPass:
			report.Summary.Passed++
		case doctor.StatusFail:
			report.Summary.Failed++
		case doctor.StatusWarn:
			report.Summary.Warned++
		}
	}
	return report
}

func init() {
	doctorCmd.Flags().BoolVar(&doctorFix, "fix", false, "失敗したチェックの修復コマンドを実行する")
	doctorCmd.Flags().BoolVar(&doctorJSON, "json", false, "JSONで出力する")
	rootCmd.AddCommand(doctorCmd)
}
