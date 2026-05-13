package cmd

import (
	"fmt"
	"os"
	"path/filepath"
	"sort"
	"strings"
	"time"

	"github.com/charmbracelet/huh"
	"github.com/ryugen04/sango/internal/config"
	"github.com/ryugen04/sango/internal/worktree"
	"github.com/spf13/cobra"
)

var (
	wtCreateServices string
	wtCreateNoSetup  bool
	wtCreateFrom     string
	wtCreateFetch    bool
)

var worktreeCreateCmd = &cobra.Command{
	Use:   "create [branch]",
	Short: "新しいワークツリーを作成する",
	Args:  cobra.MaximumNArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		cfg, err := loadConfig()
		if err != nil {
			return err
		}
		targetServices, err := resolveWorktreeServices(cfg, wtCreateServices)
		if err != nil {
			return err
		}
		if wtCreateFetch || cfg.Worktree.Create.AutoFetch {
			if err := fetchWorktreeCreateBranchCandidates(worktree.DefaultDir(), cfg, targetServices); err != nil {
				return err
			}
		}
		remoteBranches := collectCommonRemoteBranches(worktree.DefaultDir(), cfg, targetServices)
		branch, err := resolveWorktreeCreateBranch(args, remoteBranches)
		if err != nil {
			return err
		}
		if err := maybePromptWorktreeCreateOptions(cfg, len(args) == 0, cmd.Flags().Changed("from"), cmd.Flags().Changed("no-setup")); err != nil {
			return err
		}
		return runWorktreeCreate(cfg, branch, targetServices)
	},
}

func init() {
	worktreeCreateCmd.Flags().StringVar(&wtCreateServices, "services", "", "対象サービス（カンマ区切り）")
	worktreeCreateCmd.Flags().BoolVar(&wtCreateNoSetup, "no-setup", false, "セットアップをスキップする")
	worktreeCreateCmd.Flags().StringVar(&wtCreateFrom, "from", "", "ベースブランチ名")
	worktreeCreateCmd.Flags().BoolVar(&wtCreateFetch, "fetch", false, "ブランチ候補表示前に対象リポジトリを fetch する")
	worktreeCmd.AddCommand(worktreeCreateCmd)
}

func runWorktreeCreate(cfg *config.Config, branch string, targetServices []string) error {
	sangoDir := worktree.DefaultDir()
	wtDir := cfg.Worktree.WorktreeDir(branch)
	absWtRoot, err := filepath.Abs(wtDir)
	if err != nil {
		return fmt.Errorf("ワークツリーパスの解決に失敗: %w", err)
	}

	// ロック取得
	lock, err := worktree.AcquireLock(sangoDir, "worktree-op")
	if err != nil {
		return fmt.Errorf("ロックの取得に失敗: %w", err)
	}
	defer lock.Release()

	ws, err := worktree.Load(sangoDir)
	if err != nil {
		return fmt.Errorf("worktrees.jsonの読み込みに失敗: %w", err)
	}

	// 既存チェック
	if _, exists := ws.Worktrees[branch]; exists {
		return fmt.Errorf("ワークツリー %q は既に存在します", branch)
	}

	// オフセット割り当て
	baseOffset := cfg.Ports.BaseOffset
	if baseOffset == 0 {
		baseOffset = 100
	}
	offset := ws.AllocateOffset(baseOffset)

	// グローバルベースブランチ
	globalBaseBranch := wtCreateFrom
	if globalBaseBranch == "" {
		globalBaseBranch = cfg.Worktree.DefaultBranch
		if globalBaseBranch == "" {
			globalBaseBranch = "main"
		}
	}

	// ロールバック用: worktreeパスとサービス名のペアを追跡
	type createdEntry struct {
		path        string
		serviceName string
	}
	var created []createdEntry
	var allServiceNames []string

	// 各サービスについてworktree作成
	for _, name := range targetServices {
		svc, ok := cfg.Services[name]
		if !ok {
			return fmt.Errorf("サービス %q は設定に存在しません", name)
		}

		if svc.Type == "docker" || svc.Repo == "" {
			allServiceNames = append(allServiceNames, name)
			continue
		}

		// サービスごとのベースブランチ（--from 未指定時のみサービス設定を優先）
		baseBranch := globalBaseBranch
		if wtCreateFrom == "" && svc.DefaultBranch != "" {
			baseBranch = svc.DefaultBranch
		}

		absWtPath, err := filepath.Abs(filepath.Join(wtDir, name))
		if err != nil {
			return fmt.Errorf("ワークツリーパスの解決に失敗: %w", err)
		}
		// ベアリポジトリで最新を取得
		bareDir := worktree.BareRepoDir(sangoDir, name)
		fmt.Printf("[sango] fetch: %s\n", name)
		if err := worktree.FetchOrigin(bareDir); err != nil {
			return fmt.Errorf("サービス %s の git fetch に失敗: %w", name, err)
		}

		fmt.Printf("[sango] create: %s (branch: %s, from: %s)\n", name, branch, baseBranch)

		// まず新規ブランチ作成を試み、既存ブランチなら既存ブランチでworktree追加
		if err := worktree.WorktreeAddNewBranch(sangoDir, name, absWtPath, branch, baseBranch); err != nil {
			// 既存ブランチの場合はWorktreeAddにフォールバック
			if err2 := worktree.WorktreeAdd(sangoDir, name, absWtPath, branch); err2 != nil {
				// 両方失敗した場合のみロールバック
				for _, e := range created {
					fmt.Printf("[sango] ロールバック: %s を削除中...\n", e.serviceName)
					if rbErr := worktree.WorktreeRemove(sangoDir, e.serviceName, e.path, true); rbErr != nil {
						fmt.Printf("[sango] ロールバック警告: %s の削除に失敗: %v\n", e.serviceName, rbErr)
					}
				}
				return fmt.Errorf("サービス %s のワークツリー作成に失敗: %w", name, err2)
			}
		}

		created = append(created, createdEntry{path: absWtPath, serviceName: name})
		allServiceNames = append(allServiceNames, name)
	}

	// 後処理を実行（include展開、setup、hooks）
	fmt.Println("[sango] postprocess")
	ppResult := worktree.RunPostProcess(cfg, sangoDir, branch, allServiceNames, offset, worktree.PostProcessOptions{
		SkipSetup: wtCreateNoSetup,
		Quiet:     true,
	})

	// include展開エラーの処理
	if ppResult.IncludeResult != nil && ppResult.IncludeResult.HasErrors() {
		// ロールバック実行
		for _, e := range created {
			fmt.Printf("[sango] ロールバック: %s を削除中...\n", e.serviceName)
			if rbErr := worktree.WorktreeRemove(sangoDir, e.serviceName, e.path, true); rbErr != nil {
				fmt.Printf("[sango] ロールバック警告: %s の削除に失敗: %v\n", e.serviceName, rbErr)
			}
		}
		return fmt.Errorf("必須includeエントリの展開に失敗: %w", ppResult.IncludeResult.CriticalError())
	}
	if ppResult.IncludeResult != nil {
		if warning := ppResult.IncludeResult.WarningError(); warning != nil {
			fmt.Printf("[sango] include展開で警告: %v\n", warning)
		}
	}

	// setupエラーの警告
	for _, e := range ppResult.SetupErrors {
		fmt.Printf("[sango] セットアップ警告: %v\n", e)
	}

	// hooksエラーの警告
	for _, e := range ppResult.HookErrors {
		fmt.Printf("[sango] フック警告: %v\n", e)
	}

	// worktrees.json更新
	ws.AddWorktree(branch, &worktree.WorktreeInfo{
		Offset:     offset,
		CreatedAt:  time.Now(),
		Services:   allServiceNames,
		FromBranch: globalBaseBranch,
	})

	if err := ws.Save(sangoDir); err != nil {
		return fmt.Errorf("worktrees.jsonの保存に失敗: %w", err)
	}

	fmt.Printf("[sango] created: %s (offset: %d)\n", branch, offset)
	fmt.Printf("[sango] path: %s\n", absWtRoot)
	return nil
}

// repoInfo はリポジトリの表示用情報
type repoInfo struct {
	Name    string   // リポジトリ名（サービス名）
	Servers []string // このリポジトリに紐づくサーバー名
}

// collectRepos はConfigからリポジトリ一覧を収集する
func collectRepos(cfg *config.Config) []repoInfo {
	// repo フィールドを持つサービス = リポジトリ
	repos := make(map[string]*repoInfo)
	var repoOrder []string

	for name, svc := range cfg.Services {
		if svc.Repo != "" {
			repos[name] = &repoInfo{Name: name}
			repoOrder = append(repoOrder, name)
		}
	}
	sort.Strings(repoOrder)

	// repo_name で参照しているサーバーを紐づける
	for name, svc := range cfg.Services {
		if svc.RepoName != "" {
			if ri, ok := repos[svc.RepoName]; ok {
				ri.Servers = append(ri.Servers, name)
			}
		}
	}

	// サーバー名をソート
	for _, ri := range repos {
		sort.Strings(ri.Servers)
	}

	var result []repoInfo
	for _, name := range repoOrder {
		result = append(result, *repos[name])
	}
	return result
}

// reposToServices は選択されたリポジトリ名から対象サービスリストを返す
// リポジトリ自体 + repo_name で参照するサーバー + shared サービスを含む
func reposToServices(cfg *config.Config, selectedRepos []string) []string {
	selected := make(map[string]bool)
	for _, r := range selectedRepos {
		selected[r] = true
	}

	var services []string
	for name, svc := range cfg.Services {
		// sharedサービスは常に含める
		if svc.Shared {
			services = append(services, name)
			continue
		}
		// 選択されたリポジトリ自体
		if selected[name] {
			services = append(services, name)
			continue
		}
		// repo_name が選択されたリポジトリを参照している
		if svc.RepoName != "" && selected[svc.RepoName] {
			services = append(services, name)
		}
	}
	sort.Strings(services)
	return services
}

// resolveWorktreeServices は対象サービスリストを返す
func resolveWorktreeServices(cfg *config.Config, servicesFlag string) ([]string, error) {
	// --services フラグが指定された場合はそのまま返す
	if servicesFlag != "" {
		return splitServiceList(servicesFlag), nil
	}

	if len(cfg.Worktree.Create.DefaultServices) > 0 && !stdinIsTerminal() {
		return resolveDefaultWorktreeServices(cfg), nil
	}

	// 非インタラクティブ環境ではエラーにする
	if !stdinIsTerminal() {
		return nil, fmt.Errorf("非インタラクティブ環境では --services フラグ、または worktree.create.default_services を指定してください")
	}

	// インタラクティブにリポジトリを選択
	repos := collectRepos(cfg)
	if len(repos) == 0 {
		// リポジトリがない場合は全サービスを返す
		var names []string
		for name := range cfg.Services {
			names = append(names, name)
		}
		return names, nil
	}

	defaultRepos := defaultSelectedRepos(cfg)

	var options []huh.Option[string]
	for _, ri := range repos {
		desc := ri.Name
		if len(ri.Servers) > 0 {
			desc = fmt.Sprintf("%s (%s)", ri.Name, strings.Join(ri.Servers, ", "))
		} else {
			desc = fmt.Sprintf("%s (サーバーなし)", ri.Name)
		}
		selected := len(defaultRepos) == 0 || defaultRepos[ri.Name]
		options = append(options, huh.NewOption(desc, ri.Name).Selected(selected))
	}

	var selectedRepos []string
	form := huh.NewForm(
		huh.NewGroup(
			huh.NewMultiSelect[string]().
				Title("対象リポジトリを選択してください").
				Options(options...).
				Value(&selectedRepos),
		),
	).WithTheme(sangoPromptTheme())

	if err := form.Run(); err != nil {
		return nil, fmt.Errorf("リポジトリ選択がキャンセルされました: %w", err)
	}

	if len(selectedRepos) == 0 {
		return nil, fmt.Errorf("リポジトリが選択されていません")
	}

	return reposToServices(cfg, selectedRepos), nil
}

func splitServiceList(raw string) []string {
	parts := strings.Split(raw, ",")
	services := make([]string, 0, len(parts))
	for _, part := range parts {
		name := strings.TrimSpace(part)
		if name != "" {
			services = append(services, name)
		}
	}
	return services
}

func resolveDefaultWorktreeServices(cfg *config.Config) []string {
	defaults := cfg.Worktree.Create.DefaultServices
	if len(defaults) == 0 {
		return nil
	}

	selectedRepos := make([]string, 0, len(defaults))
	extra := make(map[string]bool)
	for _, name := range defaults {
		svc := cfg.Services[name]
		if svc == nil {
			continue
		}
		switch {
		case svc.RepoName != "":
			selectedRepos = append(selectedRepos, svc.RepoName)
		case svc.Repo != "":
			selectedRepos = append(selectedRepos, name)
		default:
			extra[name] = true
		}
	}

	resultSet := make(map[string]bool)
	for _, name := range reposToServices(cfg, selectedRepos) {
		resultSet[name] = true
	}
	for name := range extra {
		resultSet[name] = true
	}

	result := make([]string, 0, len(resultSet))
	for name := range resultSet {
		result = append(result, name)
	}
	sort.Strings(result)
	return result
}

func defaultSelectedRepos(cfg *config.Config) map[string]bool {
	if len(cfg.Worktree.Create.DefaultServices) == 0 {
		return nil
	}
	selected := make(map[string]bool)
	for _, name := range cfg.Worktree.Create.DefaultServices {
		svc := cfg.Services[name]
		if svc == nil {
			continue
		}
		if svc.RepoName != "" {
			selected[svc.RepoName] = true
			continue
		}
		if svc.Repo != "" {
			selected[name] = true
		}
	}
	return selected
}

func collectCommonRemoteBranches(sangoDir string, cfg *config.Config, services []string) []string {
	repoNames := repoServicesFromTargets(cfg, services)
	if len(repoNames) == 0 {
		return nil
	}

	var common map[string]bool
	for _, name := range repoNames {
		branches, err := worktree.ListRemoteBranches(sangoDir, name)
		if err != nil {
			return nil
		}
		current := make(map[string]bool, len(branches))
		for _, branch := range branches {
			current[branch] = true
		}
		if common == nil {
			common = current
			continue
		}
		for branch := range common {
			if !current[branch] {
				delete(common, branch)
			}
		}
	}

	branches := make([]string, 0, len(common))
	for branch := range common {
		branches = append(branches, branch)
	}
	sort.Strings(branches)
	return branches
}

func fetchWorktreeCreateBranchCandidates(sangoDir string, cfg *config.Config, services []string) error {
	for _, name := range repoServicesFromTargets(cfg, services) {
		fmt.Printf("[sango] fetch branches: %s\n", name)
		bareDir := worktree.BareRepoDir(sangoDir, name)
		if err := worktree.FetchOrigin(bareDir); err != nil {
			return fmt.Errorf("サービス %s のブランチ候補 fetch に失敗: %w", name, err)
		}
	}
	return nil
}

func repoServicesFromTargets(cfg *config.Config, services []string) []string {
	seen := make(map[string]bool)
	for _, name := range services {
		svc := cfg.Services[name]
		if svc == nil {
			continue
		}
		repoName := ""
		if svc.Repo != "" {
			repoName = name
		} else if svc.RepoName != "" {
			repoName = svc.RepoName
		}
		if repoName != "" {
			seen[repoName] = true
		}
	}
	repos := make([]string, 0, len(seen))
	for name := range seen {
		repos = append(repos, name)
	}
	sort.Strings(repos)
	return repos
}

// isTerminal は標準入力がターミナルかどうかを判定する
func isTerminal() bool {
	fi, err := os.Stdin.Stat()
	if err != nil {
		return false
	}
	return fi.Mode()&os.ModeCharDevice != 0
}
