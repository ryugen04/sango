package cmd

import (
	"encoding/json"
	"fmt"
	"io"
	"os"
	"path/filepath"

	"github.com/ryugen04/sango/internal/audit"
	"github.com/spf13/cobra"
)

var (
	auditRootFlag         string
	auditInventoryOutput  string
	auditInventoryFormat  string
	auditInventoryRuntime bool
)

var auditCmd = &cobra.Command{
	Use:   "audit",
	Short: "設定・運用資産を監査する",
}

var auditInventoryCmd = &cobra.Command{
	Use:   "inventory",
	Short: "設定・workflow・runtime 資産の棚卸しレポートを生成する",
	RunE: func(cmd *cobra.Command, args []string) error {
		return runAuditInventory(auditRootFlag, auditInventoryOutput, auditInventoryFormat, auditInventoryRuntime, cmd.OutOrStdout(), cmd.ErrOrStderr())
	},
}

func init() {
	auditInventoryCmd.Flags().StringVar(&auditRootFlag, "root", ".", "監査対象ルート")
	auditInventoryCmd.Flags().StringVar(&auditInventoryOutput, "output", ".sango/audit/inventory.json", "JSON出力先")
	auditInventoryCmd.Flags().StringVar(&auditInventoryFormat, "format", "both", "出力形式: text|json|both")
	auditInventoryCmd.Flags().BoolVar(&auditInventoryRuntime, "include-runtime", false, "runtimeディレクトリも含める")
	auditCmd.AddCommand(auditInventoryCmd)
	rootCmd.AddCommand(auditCmd)
}

func runAuditInventory(root, output, format string, includeRuntime bool, stdout, stderr io.Writer) error {
	if format != "text" && format != "json" && format != "both" {
		return fmt.Errorf("unsupported format: %s", format)
	}

	report, err := audit.CollectInventory(root, audit.Options{
		IncludeRuntime: includeRuntime,
	})
	if err != nil {
		return err
	}

	if format == "text" || format == "both" {
		_, _ = fmt.Fprintln(stdout, audit.RenderText(report))
	}

	for _, warning := range report.Warnings {
		_, _ = fmt.Fprintf(stderr, "warning: %s\n", warning)
	}

	if format == "json" || format == "both" {
		outputPath := resolveAuditOutputPath(report.Root, output)
		if err := writeAuditJSON(outputPath, report); err != nil {
			return err
		}
		if format == "both" {
			_, _ = fmt.Fprintf(stdout, "json written: %s\n", outputPath)
		}
	}

	return nil
}

func resolveAuditOutputPath(root, output string) string {
	if filepath.IsAbs(output) {
		return output
	}
	return filepath.Join(root, output)
}

func writeAuditJSON(path string, report audit.Report) error {
	if err := os.MkdirAll(filepath.Dir(path), 0o755); err != nil {
		return err
	}
	raw, err := json.MarshalIndent(report, "", "  ")
	if err != nil {
		return err
	}
	raw = append(raw, '\n')
	return os.WriteFile(path, raw, 0o644)
}
