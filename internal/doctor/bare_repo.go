package doctor

import (
	"fmt"
	"os"
	"path/filepath"
	"sort"
	"strings"

	"github.com/ryugen04/sango/internal/worktree"
)

// CheckBareRepoFetchRefspec は .sango/bare/*.git に remote.origin.fetch があるか検査する。
func CheckBareRepoFetchRefspec(sangoDir string) []CheckResult {
	bareRoot := filepath.Join(sangoDir, "bare")
	entries, err := os.ReadDir(bareRoot)
	if err != nil {
		if os.IsNotExist(err) {
			return nil
		}
		return []CheckResult{{
			Name:    "bare repo fetch refspec",
			Status:  StatusWarn,
			Message: fmt.Sprintf("bare repo ディレクトリの読み込みに失敗: %v", err),
		}}
	}

	var missing []string
	for _, entry := range entries {
		if !entry.IsDir() || !strings.HasSuffix(entry.Name(), ".git") {
			continue
		}
		bareDir := filepath.Join(bareRoot, entry.Name())
		ok, err := worktree.HasBareRepoFetchRefspec(bareDir)
		if err != nil {
			return []CheckResult{{
				Name:    "bare repo fetch refspec",
				Status:  StatusWarn,
				Message: err.Error(),
			}}
		}
		if !ok {
			missing = append(missing, bareDir)
		}
	}

	if len(missing) == 0 {
		return []CheckResult{{
			Name:    "bare repo fetch refspec",
			Status:  StatusPass,
			Message: "OK",
		}}
	}

	sort.Strings(missing)
	return []CheckResult{{
		Name:    "bare repo fetch refspec",
		Status:  StatusFail,
		Message: fmt.Sprintf("remote.origin.fetch が未設定: %s", strings.Join(missing, ", ")),
		Fix:     bareRepoFetchRefspecFixCommand(sangoDir),
	}}
}

func bareRepoFetchRefspecFixCommand(sangoDir string) string {
	return fmt.Sprintf("for bare in %s; do git --git-dir=\"$bare\" config remote.origin.fetch '%s'; done",
		shellQuote(filepath.Join(sangoDir, "bare"))+"/*.git",
		worktree.BareRepoFetchRefspec,
	)
}

func shellQuote(s string) string {
	return "'" + strings.ReplaceAll(s, "'", "'\\''") + "'"
}
