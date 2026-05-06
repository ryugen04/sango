package cmd

import (
	"fmt"

	"github.com/spf13/cobra"
)

var (
	rootJSON bool
	rootFrom string
)

type rootOutput struct {
	machineMeta
	Root       string `json:"root"`
	ConfigPath string `json:"config_path"`
	Name       string `json:"name"`
}

var rootDetectCmd = &cobra.Command{
	Use:   "root",
	Short: "sango プロジェクトルートを表示する",
	RunE: func(cmd *cobra.Command, args []string) error {
		ctx, err := resolveProjectContext(rootFrom)
		if err != nil {
			return err
		}
		if rootJSON {
			out := rootOutput{
				machineMeta: newMachineMeta(ctx.Root),
				Root:        ctx.Root,
				ConfigPath:  ctx.ConfigPath,
				Name:        ctx.Name,
			}
			return writeJSON(cmd.OutOrStdout(), out)
		}
		fmt.Fprintln(cmd.OutOrStdout(), ctx.Root)
		return nil
	},
}

func init() {
	rootDetectCmd.Flags().BoolVar(&rootJSON, "json", false, "JSONで出力する")
	rootDetectCmd.Flags().StringVar(&rootFrom, "from", "", "探索開始ディレクトリ")
	rootCmd.AddCommand(rootDetectCmd)
}
