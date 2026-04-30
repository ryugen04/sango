package cmd

import (
	"fmt"
	"os/exec"
	"sort"

	"github.com/ryugen04/sango/internal/troubleshoot"
	"github.com/spf13/cobra"
)

var troubleshootFix bool
var troubleshootJSON bool

type troubleshootTarget struct {
	ServiceName string                     `json:"service_name"`
	Results     []troubleshoot.CheckResult `json:"results"`
}

type troubleshootSummary struct {
	Passed int `json:"passed"`
	Failed int `json:"failed"`
}

type troubleshootReport struct {
	Targets []troubleshootTarget `json:"targets"`
	Summary troubleshootSummary  `json:"summary"`
}

var troubleshootCmd = &cobra.Command{
	Use:   "troubleshoot [service]",
	Short: "サービスのトラブルシュートチェックを実行する",
	RunE: func(cmd *cobra.Command, args []string) error {
		cfg, err := loadConfig()
		if err != nil {
			return err
		}

		var targets []troubleshootTarget

		if len(args) > 0 {
			// 指定サービスのみ
			svcName := args[0]
			svc, ok := cfg.Services[svcName]
			if !ok {
				return fmt.Errorf("サービス %q が見つかりません", svcName)
			}
			if len(svc.Troubleshoot) == 0 {
				if troubleshootJSON {
					return writeJSON(cmd.OutOrStdout(), troubleshootReport{})
				}
				fmt.Printf("[sango] %s にトラブルシュートチェックが定義されていません\n", svcName)
				return nil
			}
			results := troubleshoot.Run(svc.Troubleshoot)
			targets = append(targets, troubleshootTarget{ServiceName: svcName, Results: results})
		} else {
			// 全サービス
			names := make([]string, 0, len(cfg.Services))
			for name, svc := range cfg.Services {
				if len(svc.Troubleshoot) == 0 {
					continue
				}
				names = append(names, name)
			}
			sort.Strings(names)
			for _, name := range names {
				results := troubleshoot.Run(cfg.Services[name].Troubleshoot)
				targets = append(targets, troubleshootTarget{ServiceName: name, Results: results})
			}
		}

		if len(targets) == 0 {
			if troubleshootJSON {
				return writeJSON(cmd.OutOrStdout(), buildTroubleshootReport(nil))
			}
			fmt.Println("[sango] トラブルシュートチェックが定義されているサービスがありません")
			return nil
		}

		report := buildTroubleshootReport(targets)

		if troubleshootJSON {
			return writeJSON(cmd.OutOrStdout(), report)
		}

		for _, t := range report.Targets {
			fmt.Printf("[sango] %s のトラブルシュート実行中...\n", t.ServiceName)
			for _, r := range t.Results {
				switch r.Status {
				case troubleshoot.StatusPass:
					fmt.Printf("  %s %s - %s\n", passStyle.Render("[pass]"), r.Name, r.Output)
				case troubleshoot.StatusFail:
					fmt.Printf("  %s %s - %s\n", failStyle.Render("[fail]"), r.Name, r.Output)
					if r.Fix != "" {
						fmt.Printf("    修復: %s\n", r.Fix)
					}
				}
			}
			fmt.Println()
		}

		fmt.Printf("結果: %d passed, %d failed\n", report.Summary.Passed, report.Summary.Failed)

		if troubleshootFix {
			for _, t := range report.Targets {
				for _, r := range t.Results {
					if r.Status == troubleshoot.StatusFail && r.Fix != "" {
						fmt.Printf("\n[fix] %s/%s: %s\n", t.ServiceName, r.Name, r.Fix)
						out, err := exec.Command("sh", "-c", r.Fix).CombinedOutput()
						if err != nil {
							fmt.Printf("[fix] 失敗: %s\n", string(out))
						} else {
							fmt.Printf("[fix] 成功: %s\n", string(out))
						}
					}
				}
			}
		}

		return nil
	},
}

func buildTroubleshootReport(targets []troubleshootTarget) troubleshootReport {
	report := troubleshootReport{Targets: targets}
	if report.Targets == nil {
		report.Targets = []troubleshootTarget{}
	}
	for _, t := range targets {
		for _, r := range t.Results {
			switch r.Status {
			case troubleshoot.StatusPass:
				report.Summary.Passed++
			case troubleshoot.StatusFail:
				report.Summary.Failed++
			}
		}
	}
	return report
}

func init() {
	troubleshootCmd.Flags().BoolVar(&troubleshootFix, "fix", false, "失敗したチェックの修復コマンドを実行する")
	troubleshootCmd.Flags().BoolVar(&troubleshootJSON, "json", false, "JSONで出力する")
	rootCmd.AddCommand(troubleshootCmd)
}
