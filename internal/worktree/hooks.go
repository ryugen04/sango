package worktree

import (
	"fmt"
	"os/exec"
	"path/filepath"
	"strings"

	"github.com/ryugen04/sango/internal/config"
)

type HookRunOptions struct {
	Quiet bool
}

// RunHooks はフックエントリを実行する
// per_service=trueなら各サービスディレクトリで実行
func RunHooks(hooks []config.HookEntry, branchDir string, serviceNames []string) error {
	return RunHooksWithOptions(hooks, branchDir, serviceNames, HookRunOptions{})
}

func RunHooksWithOptions(hooks []config.HookEntry, branchDir string, serviceNames []string, opts HookRunOptions) error {
	for _, hook := range hooks {
		if hook.PerService {
			for _, name := range serviceNames {
				dir := filepath.Join(branchDir, name)
				if opts.Quiet {
					fmt.Printf("[sango] hook: %s\n", name)
				} else {
					fmt.Printf("[sango] フック実行 (%s): %s\n", name, hook.Command)
				}
				c := exec.Command("sh", "-c", hook.Command)
				c.Dir = dir
				if out, err := c.CombinedOutput(); err != nil {
					fmt.Printf("[sango] フック警告 (%s): %v%s\n", name, err, formatCommandOutput(out))
				}
			}
		} else {
			if opts.Quiet {
				fmt.Println("[sango] hook: post_create")
			} else {
				fmt.Printf("[sango] フック実行: %s\n", hook.Command)
			}
			c := exec.Command("sh", "-c", hook.Command)
			c.Dir = branchDir
			if out, err := c.CombinedOutput(); err != nil {
				fmt.Printf("[sango] フック警告: %v%s\n", err, formatCommandOutput(out))
			}
		}
	}
	return nil
}

func formatCommandOutput(out []byte) string {
	text := strings.TrimSpace(string(out))
	if text == "" {
		return ""
	}
	return "\n" + text
}
