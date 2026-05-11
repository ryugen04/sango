package cmd

import (
	"fmt"
	"os"
	"path/filepath"

	"github.com/ryugen04/sango/internal/config"
	"github.com/ryugen04/sango/internal/dag"
	"github.com/ryugen04/sango/internal/service"
	"github.com/spf13/cobra"
)

var cfgFile string
var rootDir string
var worktreeFlag string

var rootCmd = &cobra.Command{
	Use:   "sango",
	Short: "複数 repo / worktree の開発環境を迷わず扱うための CLI",
}

func init() {
	rootCmd.PersistentFlags().StringVar(&cfgFile, "config", "", "設定ファイルパス")
	rootCmd.PersistentFlags().StringVar(&rootDir, "root", "", "プロジェクトルート")
	rootCmd.PersistentFlags().StringVar(&worktreeFlag, "worktree", "", "ワークツリー名（省略時はアクティブ）")
}

// Execute はルートコマンドを実行する
func Execute() error {
	return rootCmd.Execute()
}

// loadConfig は設定ファイルの読み込み・検証・変数展開をまとめて行う
func loadConfig() (*config.Config, error) {
	ctx, err := resolveProjectContext("")
	if err != nil {
		return nil, err
	}
	cfgFile = ctx.ConfigPath
	return service.LoadAndValidateConfig(ctx.ConfigPath)
}

// resolveActiveWorktree は使用するworktree名を解決する
func resolveActiveWorktree(sangoDir string) string {
	return service.ResolveActiveWorktree(sangoDir, worktreeFlag)
}

func resolveActiveWorktreeWithConfig(sangoDir string, cfg *config.Config) string {
	return service.ResolveActiveWorktreeWithBaseDir(sangoDir, worktreeFlag, cfg.Worktree.ResolveBaseDir())
}

// buildDAG は設定からDAGを構築する
func buildDAG(cfg *config.Config) *dag.DAG {
	return service.BuildDAG(cfg)
}

type projectContext struct {
	Root       string
	ConfigPath string
	Name       string
}

func resolveProjectContext(from string) (*projectContext, error) {
	if cfgFile != "" {
		return projectContextFromConfig(cfgFile)
	}
	if rootDir != "" {
		return projectContextFromRoot(rootDir)
	}

	start := from
	if start == "" {
		var err error
		start, err = os.Getwd()
		if err != nil {
			return nil, err
		}
	}
	return findProjectContext(start)
}

func projectContextFromConfig(path string) (*projectContext, error) {
	absConfig, err := filepath.Abs(path)
	if err != nil {
		return nil, fmt.Errorf("設定ファイルパスの解決に失敗: %w", err)
	}
	root := filepath.Dir(absConfig)
	return loadProjectContext(root, absConfig)
}

func projectContextFromRoot(root string) (*projectContext, error) {
	absRoot, err := filepath.Abs(root)
	if err != nil {
		return nil, fmt.Errorf("プロジェクトルートの解決に失敗: %w", err)
	}
	return loadProjectContext(absRoot, filepath.Join(absRoot, "sango.yaml"))
}

func findProjectContext(start string) (*projectContext, error) {
	absStart, err := filepath.Abs(start)
	if err != nil {
		return nil, fmt.Errorf("探索開始ディレクトリの解決に失敗: %w", err)
	}
	if info, err := os.Stat(absStart); err == nil && !info.IsDir() {
		absStart = filepath.Dir(absStart)
	}

	dir := absStart
	for {
		candidate := filepath.Join(dir, "sango.yaml")
		if _, err := os.Stat(candidate); err == nil {
			return loadProjectContext(dir, candidate)
		}
		parent := filepath.Dir(dir)
		if parent == dir {
			break
		}
		dir = parent
	}
	return nil, fmt.Errorf("sango.yaml が見つかりません: %s から上方向に探索しました", absStart)
}

func loadProjectContext(root, configPath string) (*projectContext, error) {
	cfg, err := config.Load(configPath)
	if err != nil {
		return nil, err
	}
	return &projectContext{
		Root:       root,
		ConfigPath: configPath,
		Name:       cfg.Name,
	}, nil
}

func currentProjectRoot() (string, error) {
	ctx, err := resolveProjectContext("")
	if err != nil {
		return "", err
	}
	return ctx.Root, nil
}

func currentSangoDir() (string, error) {
	root, err := currentProjectRoot()
	if err != nil {
		return "", err
	}
	return filepath.Join(root, ".sango"), nil
}
