package cmd

import (
	"fmt"
	"sort"
	"strings"

	"github.com/charmbracelet/huh"
	"github.com/charmbracelet/lipgloss"
	"github.com/ryugen04/sango/internal/config"
	"github.com/ryugen04/sango/internal/service"
	"github.com/ryugen04/sango/internal/worktree"
)

var stdinIsTerminal = isTerminal

func sangoPromptTheme() *huh.Theme {
	t := huh.ThemeBase()
	var (
		primary = lipgloss.AdaptiveColor{Light: "#006D75", Dark: "#5CCFE6"}
		success = lipgloss.AdaptiveColor{Light: "#007A5A", Dark: "#8BD49C"}
		muted   = lipgloss.AdaptiveColor{Light: "242", Dark: "244"}
		text    = lipgloss.AdaptiveColor{Light: "235", Dark: "252"}
		err     = lipgloss.AdaptiveColor{Light: "#B42318", Dark: "#FF7A7A"}
	)

	t.Focused.Base = t.Focused.Base.BorderForeground(primary)
	t.Focused.Title = t.Focused.Title.Foreground(primary).Bold(true)
	t.Focused.Description = t.Focused.Description.Foreground(muted)
	t.Focused.ErrorIndicator = t.Focused.ErrorIndicator.Foreground(err)
	t.Focused.ErrorMessage = t.Focused.ErrorMessage.Foreground(err)
	t.Focused.SelectSelector = t.Focused.SelectSelector.Foreground(primary)
	t.Focused.MultiSelectSelector = t.Focused.MultiSelectSelector.Foreground(primary)
	t.Focused.Option = t.Focused.Option.Foreground(text)
	t.Focused.SelectedOption = t.Focused.SelectedOption.Foreground(success)
	t.Focused.SelectedPrefix = lipgloss.NewStyle().Foreground(success).SetString("✓ ")
	t.Focused.UnselectedOption = t.Focused.UnselectedOption.Foreground(text)
	t.Focused.UnselectedPrefix = lipgloss.NewStyle().Foreground(muted).SetString("• ")
	t.Focused.FocusedButton = t.Focused.FocusedButton.Foreground(lipgloss.Color("0")).Background(primary)
	t.Focused.BlurredButton = t.Focused.BlurredButton.Foreground(text).Background(lipgloss.AdaptiveColor{Light: "252", Dark: "237"})
	t.Focused.TextInput.Cursor = t.Focused.TextInput.Cursor.Foreground(primary)
	t.Focused.TextInput.Placeholder = t.Focused.TextInput.Placeholder.Foreground(muted)
	t.Focused.TextInput.Prompt = t.Focused.TextInput.Prompt.Foreground(primary)

	t.Blurred = t.Focused
	t.Blurred.Base = t.Focused.Base.BorderStyle(lipgloss.HiddenBorder())
	t.Blurred.NextIndicator = lipgloss.NewStyle()
	t.Blurred.PrevIndicator = lipgloss.NewStyle()
	t.Blurred.MultiSelectSelector = lipgloss.NewStyle().SetString("  ")
	t.Group.Title = t.Focused.Title
	t.Group.Description = t.Focused.Description
	return t
}

var promptWorktreeNameInput = func() (string, error) {
	return promptWorktreeNameInputWithDescription("")
}

func promptWorktreeNameInputWithDescription(description string) (string, error) {
	var branch string
	input := huh.NewInput().
		Title("作成する worktree 名を入力してください").
		Placeholder("feature/my-branch").
		Value(&branch).
		Validate(func(value string) error {
			if strings.TrimSpace(value) == "" {
				return fmt.Errorf("worktree 名は必須です")
			}
			return nil
		})
	if description != "" {
		input = input.Description(description)
	}
	form := huh.NewForm(huh.NewGroup(input)).WithTheme(sangoPromptTheme())
	if err := form.Run(); err != nil {
		return "", fmt.Errorf("worktree 名入力がキャンセルされました: %w", err)
	}
	return strings.TrimSpace(branch), nil
}

var promptExistingWorktreeSelection = func(title string, names []string, active string) (string, error) {
	options := make([]huh.Option[string], 0, len(names))
	for _, name := range names {
		label := name
		if name == active && active != "" {
			label = fmt.Sprintf("%s (active)", name)
		}
		option := huh.NewOption(label, name)
		if name == active && active != "" {
			option = option.Selected(true)
		}
		options = append(options, option)
	}

	var selected string
	form := huh.NewForm(
		huh.NewGroup(
			huh.NewSelect[string]().
				Title(title).
				Options(options...).
				Value(&selected),
		),
	).WithTheme(sangoPromptTheme())
	if err := form.Run(); err != nil {
		return "", fmt.Errorf("worktree 選択がキャンセルされました: %w", err)
	}
	return selected, nil
}

var promptCreateBaseBranch = func(defaultBranch string) (string, error) {
	options := []huh.Option[string]{
		huh.NewOption(fmt.Sprintf("既定を使う (%s)", defaultBranch), defaultBranch).Selected(true),
		huh.NewOption("カスタム入力", "__custom__"),
	}

	var choice string
	form := huh.NewForm(
		huh.NewGroup(
			huh.NewSelect[string]().
				Title("分岐元ブランチを選択してください").
				Options(options...).
				Value(&choice),
		),
	).WithTheme(sangoPromptTheme())
	if err := form.Run(); err != nil {
		return "", fmt.Errorf("分岐元ブランチ選択がキャンセルされました: %w", err)
	}
	if choice != "__custom__" {
		return choice, nil
	}

	var custom string
	customForm := huh.NewForm(
		huh.NewGroup(
			huh.NewInput().
				Title("分岐元ブランチ名を入力してください").
				Placeholder(defaultBranch).
				Value(&custom).
				Validate(func(value string) error {
					if strings.TrimSpace(value) == "" {
						return fmt.Errorf("ブランチ名は必須です")
					}
					return nil
				}),
		),
	).WithTheme(sangoPromptTheme())
	if err := customForm.Run(); err != nil {
		return "", fmt.Errorf("分岐元ブランチ入力がキャンセルされました: %w", err)
	}
	return strings.TrimSpace(custom), nil
}

var promptCreateBranchSelection = func(remoteBranches []string) (string, error) {
	if len(remoteBranches) == 0 {
		return promptWorktreeNameInputWithDescription("リモートブランチ候補が見つかりません。bare cache が古い場合は --fetch または worktree.create.auto_fetch: true を使ってください。既存ブランチ名を直接入力することもできます。")
	}

	const customChoice = "__custom__"
	const searchChoice = "__search__"
	candidates := remoteBranches
	title := "作成する worktree を選択してください"
	description := "既存のリモートブランチ、または新しい名前を選べます。そのまま入力して候補を絞り込めます"

	for {
		choice, err := promptCreateBranchSelectionOnce(title, description, candidates, searchChoice, customChoice)
		if err != nil {
			return "", err
		}
		switch choice {
		case customChoice:
			return promptWorktreeNameInput()
		case searchChoice:
			query, err := promptCreateBranchSearchQuery()
			if err != nil {
				return "", err
			}
			matches := fuzzyFilterRemoteBranches(remoteBranches, query)
			if len(matches) == 0 {
				return "", fmt.Errorf("ブランチ候補 %q が見つかりません。bare cache が古い場合は --fetch または worktree.create.auto_fetch: true を使って再実行してください。remote に存在しない場合はブランチ名を直接指定して新規作成できます", query)
			}
			candidates = matches
			title = fmt.Sprintf("%q の検索結果から選択してください", query)
			description = "検索結果を表示しています。さらに入力して候補を絞り込めます"
		default:
			return choice, nil
		}
	}
}

func promptCreateBranchSelectionOnce(title, description string, remoteBranches []string, searchChoice, customChoice string) (string, error) {
	options := make([]huh.Option[string], 0, len(remoteBranches)+1)
	options = append(options, huh.NewOption("新しい worktree 名を入力", customChoice).Selected(true))
	options = append(options, huh.NewOption("ブランチ候補を検索して絞り込む", searchChoice))
	for _, branch := range remoteBranches {
		options = append(options, huh.NewOption(fmt.Sprintf("%s をcheckout", branch), branch))
	}

	var choice string
	form := huh.NewForm(
		huh.NewGroup(
			huh.NewSelect[string]().
				Title(title).
				Description(description).
				Filtering(true).
				Height(15).
				Options(options...).
				Value(&choice),
		),
	).WithTheme(sangoPromptTheme())
	if err := form.Run(); err != nil {
		return "", fmt.Errorf("worktree 選択がキャンセルされました: %w", err)
	}
	return choice, nil
}

func promptCreateBranchSearchQuery() (string, error) {
	var query string
	form := huh.NewForm(
		huh.NewGroup(
			huh.NewInput().
				Title("ブランチ候補の検索語を入力してください").
				Placeholder("user ticket keyword").
				Description("prefix、チケット番号、ブランチ名の一部で検索できます").
				Value(&query).
				Validate(func(value string) error {
					if strings.TrimSpace(value) == "" {
						return fmt.Errorf("検索語は必須です")
					}
					return nil
				}),
		),
	).WithTheme(sangoPromptTheme())
	if err := form.Run(); err != nil {
		return "", fmt.Errorf("ブランチ候補検索がキャンセルされました: %w", err)
	}
	return strings.TrimSpace(query), nil
}

var promptCreateRunSetup = func(autoSetup bool) (bool, error) {
	runSetup := autoSetup
	description := "既定設定に従います"
	if autoSetup {
		description = "既定では実行されます"
	} else {
		description = "既定では実行されません"
	}
	form := huh.NewForm(
		huh.NewGroup(
			huh.NewConfirm().
				Title("セットアップを実行しますか？").
				Description(description).
				Affirmative("実行する").
				Negative("スキップする").
				Value(&runSetup),
		),
	).WithTheme(sangoPromptTheme())
	if err := form.Run(); err != nil {
		return false, fmt.Errorf("セットアップ設定がキャンセルされました: %w", err)
	}
	return runSetup, nil
}

type upTargetMode string

const (
	upTargetModeAll      upTargetMode = "all"
	upTargetModeProfile  upTargetMode = "profile"
	upTargetModeRepos    upTargetMode = "repos"
	upTargetModeServices upTargetMode = "services"
)

type upInteractiveSelection struct {
	Mode         upTargetMode
	Profile      string
	Repos        []string
	Services     []string
	DefaultPorts bool
}

var promptUpSelection = func(cfg *config.Config, currentDefaultPorts bool) (*upInteractiveSelection, error) {
	var mode upTargetMode
	modeOptions := []huh.Option[upTargetMode]{
		huh.NewOption("既定対象をそのまま起動", upTargetModeAll).Selected(true),
	}
	if len(cfg.Profiles) > 0 {
		modeOptions = append(modeOptions, huh.NewOption("プロファイルを選ぶ", upTargetModeProfile))
	}
	if len(collectRepos(cfg)) > 0 {
		modeOptions = append(modeOptions, huh.NewOption("リポジトリ単位で選ぶ", upTargetModeRepos))
	}
	modeOptions = append(modeOptions, huh.NewOption("サービス単位で選ぶ", upTargetModeServices))

	form := huh.NewForm(
		huh.NewGroup(
			huh.NewSelect[upTargetMode]().
				Title("起動対象の選び方を選択してください").
				Options(modeOptions...).
				Value(&mode),
		),
	)
	if err := form.Run(); err != nil {
		return nil, fmt.Errorf("起動対象選択がキャンセルされました: %w", err)
	}

	result := &upInteractiveSelection{Mode: mode}

	switch mode {
	case upTargetModeProfile:
		profiles := make([]string, 0, len(cfg.Profiles))
		for name := range cfg.Profiles {
			profiles = append(profiles, name)
		}
		sort.Strings(profiles)
		options := make([]huh.Option[string], 0, len(profiles))
		for i, name := range profiles {
			option := huh.NewOption(name, name)
			if i == 0 {
				option = option.Selected(true)
			}
			options = append(options, option)
		}
		var selected string
		profileForm := huh.NewForm(
			huh.NewGroup(
				huh.NewSelect[string]().
					Title("プロファイルを選択してください").
					Options(options...).
					Value(&selected),
			),
		)
		if err := profileForm.Run(); err != nil {
			return nil, fmt.Errorf("プロファイル選択がキャンセルされました: %w", err)
		}
		result.Profile = selected
	case upTargetModeRepos:
		repos := collectRepos(cfg)
		options := make([]huh.Option[string], 0, len(repos))
		for _, ri := range repos {
			desc := ri.Name
			if len(ri.Servers) > 0 {
				desc = fmt.Sprintf("%s (%s)", ri.Name, strings.Join(ri.Servers, ", "))
			}
			options = append(options, huh.NewOption(desc, ri.Name).Selected(true))
		}
		repoForm := huh.NewForm(
			huh.NewGroup(
				huh.NewMultiSelect[string]().
					Title("起動対象のリポジトリを選択してください").
					Options(options...).
					Value(&result.Repos),
			),
		)
		if err := repoForm.Run(); err != nil {
			return nil, fmt.Errorf("リポジトリ選択がキャンセルされました: %w", err)
		}
	case upTargetModeServices:
		serviceNames := collectRunnableServiceNames(cfg)
		options := make([]huh.Option[string], 0, len(serviceNames))
		for _, name := range serviceNames {
			options = append(options, huh.NewOption(name, name))
		}
		serviceForm := huh.NewForm(
			huh.NewGroup(
				huh.NewMultiSelect[string]().
					Title("起動対象のサービスを選択してください").
					Options(options...).
					Value(&result.Services),
			),
		)
		if err := serviceForm.Run(); err != nil {
			return nil, fmt.Errorf("サービス選択がキャンセルされました: %w", err)
		}
	}

	useDefaultPorts := currentDefaultPorts
	portsForm := huh.NewForm(
		huh.NewGroup(
			huh.NewConfirm().
				Title("デフォルトポート（offset=0）を使いますか？").
				Affirmative("使う").
				Negative("使わない").
				Value(&useDefaultPorts),
		),
	)
	if err := portsForm.Run(); err != nil {
		return nil, fmt.Errorf("ポート設定がキャンセルされました: %w", err)
	}
	result.DefaultPorts = useDefaultPorts

	return result, nil
}

func resolveWorktreeNameArg(args []string, ws *worktree.WorktreeState, fallback string, promptTitle string) (string, error) {
	if len(args) > 0 {
		return args[0], nil
	}
	if len(ws.Worktrees) == 0 {
		return "", fmt.Errorf("ワークツリーがありません")
	}
	if !stdinIsTerminal() {
		if fallback != "" {
			return fallback, nil
		}
		return "", fmt.Errorf("非インタラクティブ環境ではワークツリー名を指定してください")
	}
	return promptExistingWorktreeSelection(promptTitle, listWorktreeNames(ws), fallback)
}

func resolveWorktreeCreateBranch(args []string, remoteBranches []string) (string, error) {
	if len(args) > 0 {
		return args[0], nil
	}
	if !stdinIsTerminal() {
		return "", fmt.Errorf("非インタラクティブ環境では worktree 名を指定してください")
	}
	selected, err := promptCreateBranchSelection(remoteBranches)
	if err != nil {
		return "", err
	}
	if strings.HasPrefix(selected, "origin/") {
		wtCreateFrom = selected
		return strings.TrimPrefix(selected, "origin/"), nil
	}
	return selected, nil
}

func maybePromptWorktreeCreateOptions(cfg *config.Config, branchPrompted bool, fromChanged, noSetupChanged bool) error {
	if !branchPrompted || !stdinIsTerminal() {
		return nil
	}

	if !fromChanged && wtCreateFrom == "" {
		defaultBranch := cfg.Worktree.DefaultBranch
		if defaultBranch == "" {
			defaultBranch = "main"
		}
		selected, err := promptCreateBaseBranch(defaultBranch)
		if err != nil {
			return err
		}
		if selected != defaultBranch {
			wtCreateFrom = selected
		}
	}

	if !noSetupChanged {
		runSetup, err := promptCreateRunSetup(cfg.Worktree.AutoSetup)
		if err != nil {
			return err
		}
		wtCreateNoSetup = !runSetup
	}

	return nil
}

func fuzzyFilterRemoteBranches(branches []string, query string) []string {
	query = normalizeBranchSearchText(query)
	if query == "" {
		return branches
	}

	type match struct {
		branch string
		score  int
	}
	matches := make([]match, 0, len(branches))
	for _, branch := range branches {
		if score, ok := branchSearchScore(branch, query); ok {
			matches = append(matches, match{branch: branch, score: score})
		}
	}
	sort.SliceStable(matches, func(i, j int) bool {
		if matches[i].score != matches[j].score {
			return matches[i].score < matches[j].score
		}
		return matches[i].branch < matches[j].branch
	})

	result := make([]string, 0, len(matches))
	for _, match := range matches {
		result = append(result, match.branch)
	}
	return result
}

func branchSearchScore(branch, query string) (int, bool) {
	text := normalizeBranchSearchText(branch)
	text = strings.TrimPrefix(text, "origin/")

	if text == query {
		return 0, true
	}
	if strings.HasPrefix(text, query) {
		return 1, true
	}
	if strings.Contains(text, "/"+query) {
		return 2, true
	}
	if idx := strings.Index(text, query); idx >= 0 {
		return 10 + idx, true
	}
	if fuzzyContains(text, query) {
		return 100 + len(text) - len(query), true
	}
	return 0, false
}

func normalizeBranchSearchText(value string) string {
	value = strings.ToLower(strings.TrimSpace(value))
	value = strings.ReplaceAll(value, "_", "-")
	return value
}

func fuzzyContains(text, query string) bool {
	if query == "" {
		return true
	}
	queryRunes := []rune(query)
	pos := 0
	for _, r := range text {
		if r == queryRunes[pos] {
			pos++
			if pos == len(queryRunes) {
				return true
			}
		}
	}
	return false
}

func resolveUpTargets(cfg *config.Config, args []string, profile string, interactive bool) ([]string, error) {
	if len(args) > 0 || profile != "" || !interactive {
		return args, nil
	}

	selection, err := promptUpSelection(cfg, defaultPorts)
	if err != nil {
		return nil, err
	}
	defaultPorts = selection.DefaultPorts

	switch selection.Mode {
	case upTargetModeAll:
		return nil, nil
	case upTargetModeProfile:
		return service.ResolveTargets(cfg, nil, selection.Profile), nil
	case upTargetModeRepos:
		return filterRunnableServices(cfg, reposToServices(cfg, selection.Repos)), nil
	case upTargetModeServices:
		return selection.Services, nil
	default:
		return nil, nil
	}
}

func listWorktreeNames(ws *worktree.WorktreeState) []string {
	names := make([]string, 0, len(ws.Worktrees))
	for name := range ws.Worktrees {
		names = append(names, name)
	}
	sort.Strings(names)
	return names
}

func collectRunnableServiceNames(cfg *config.Config) []string {
	names := make([]string, 0, len(cfg.Services))
	for name, svc := range cfg.Services {
		if svc == nil {
			continue
		}
		if svc.Type == "docker" || svc.Command != "" {
			names = append(names, name)
		}
	}
	sort.Strings(names)
	return names
}

func filterRunnableServices(cfg *config.Config, names []string) []string {
	runnable := make([]string, 0, len(names))
	seen := make(map[string]bool)
	for _, name := range names {
		svc := cfg.Services[name]
		if svc == nil {
			continue
		}
		if svc.Type != "docker" && svc.Command == "" {
			continue
		}
		if seen[name] {
			continue
		}
		seen[name] = true
		runnable = append(runnable, name)
	}
	sort.Strings(runnable)
	return runnable
}
