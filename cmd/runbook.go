package cmd

import (
	"fmt"
	"sort"
	"strings"

	"github.com/ryugen04/sango/internal/config"
	"github.com/ryugen04/sango/internal/runbook"
	"github.com/spf13/cobra"
)

var runbookServiceFilter string
var runbookJSON bool

type runbookSearchOutput struct {
	machineMeta
	Keyword string                 `json:"keyword"`
	Results []runbook.SearchResult `json:"results"`
}

type runbookListService struct {
	ServiceName string                `json:"service_name"`
	Entries     []config.RunbookEntry `json:"entries"`
}

type runbookListOutput struct {
	machineMeta
	Services []runbookListService `json:"services"`
}

var runbookCmd = &cobra.Command{
	Use:   "runbook",
	Short: "サービスのRunbookを検索・一覧表示する",
}

var runbookSearchCmd = &cobra.Command{
	Use:   "search <keyword>",
	Short: "キーワードでRunbookを検索する",
	Args:  cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		cfg, err := loadConfig()
		if err != nil {
			return err
		}

		keyword := args[0]
		results := runbook.Search(cfg.Services, keyword)
		sort.Slice(results, func(i, j int) bool {
			if results[i].ServiceName == results[j].ServiceName {
				return results[i].Entry.Title < results[j].Entry.Title
			}
			return results[i].ServiceName < results[j].ServiceName
		})

		if runbookJSON {
			projectRoot, err := currentProjectRoot()
			if err != nil {
				return err
			}
			return writeJSON(cmd.OutOrStdout(), runbookSearchOutput{
				machineMeta: newMachineMeta(projectRoot),
				Keyword:     keyword,
				Results:     emptyRunbookResults(results),
			})
		}

		fmt.Printf("[sango] runbook検索: %q\n\n", keyword)

		if len(results) == 0 {
			fmt.Println("該当するエントリが見つかりませんでした")
			return nil
		}

		for _, r := range results {
			fmt.Printf("  [%s] %s\n", r.ServiceName, r.Entry.Title)
			if len(r.Entry.Symptoms) > 0 {
				fmt.Printf("    症状: %s\n", strings.Join(r.Entry.Symptoms, ", "))
			}
			if r.Entry.Cause != "" {
				fmt.Printf("    原因: %s\n", r.Entry.Cause)
			}
			if len(r.Entry.Steps) > 0 {
				fmt.Println("    手順:")
				for i, step := range r.Entry.Steps {
					fmt.Printf("      %d. %s\n", i+1, step)
				}
			}
			if len(r.Entry.Tags) > 0 {
				fmt.Printf("    タグ: %s\n", strings.Join(r.Entry.Tags, ", "))
			}
			fmt.Println()
		}

		fmt.Printf("%d件見つかりました\n", len(results))
		return nil
	},
}

var runbookListCmd = &cobra.Command{
	Use:   "list",
	Short: "Runbookを一覧表示する",
	RunE: func(cmd *cobra.Command, args []string) error {
		cfg, err := loadConfig()
		if err != nil {
			return err
		}

		services := buildRunbookList(cfg.Services, runbookServiceFilter)
		if runbookJSON {
			projectRoot, err := currentProjectRoot()
			if err != nil {
				return err
			}
			return writeJSON(cmd.OutOrStdout(), runbookListOutput{
				machineMeta: newMachineMeta(projectRoot),
				Services:    services,
			})
		}

		fmt.Println("[sango] runbook一覧")
		fmt.Println()

		if len(services) == 0 {
			fmt.Println("  Runbookが定義されているサービスがありません")
			return nil
		}

		for _, svc := range services {
			fmt.Printf("  %s:\n", svc.ServiceName)
			for _, entry := range svc.Entries {
				tagStr := ""
				if len(entry.Tags) > 0 {
					tagStr = fmt.Sprintf(" [%s]", strings.Join(entry.Tags, ", "))
				}
				fmt.Printf("    - %s%s\n", entry.Title, tagStr)
			}
		}

		return nil
	},
}

func buildRunbookList(services map[string]*config.Service, filter string) []runbookListService {
	names := make([]string, 0, len(services))
	for name, svc := range services {
		if filter != "" && name != filter {
			continue
		}
		if len(svc.Runbook) == 0 {
			continue
		}
		names = append(names, name)
	}
	sort.Strings(names)

	result := make([]runbookListService, 0, len(names))
	for _, name := range names {
		result = append(result, runbookListService{
			ServiceName: name,
			Entries:     services[name].Runbook,
		})
	}
	return result
}

func emptyRunbookResults(results []runbook.SearchResult) []runbook.SearchResult {
	if results == nil {
		return []runbook.SearchResult{}
	}
	return results
}

func init() {
	runbookListCmd.Flags().StringVar(&runbookServiceFilter, "service", "", "サービス名で絞り込み")
	runbookListCmd.Flags().BoolVar(&runbookJSON, "json", false, "JSONで出力する")
	runbookSearchCmd.Flags().BoolVar(&runbookJSON, "json", false, "JSONで出力する")
	runbookCmd.AddCommand(runbookSearchCmd)
	runbookCmd.AddCommand(runbookListCmd)
	rootCmd.AddCommand(runbookCmd)
}
