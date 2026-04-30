//go:build linux

package doctor

import (
	"fmt"
	"os"
	"os/exec"
	"strings"
)

// CheckLinuxSandbox はLinuxのuser namespace / bubblewrapの可用性をチェックする
func CheckLinuxSandbox() []CheckResult {
	var results []CheckResult

	// user namespace チェック
	data, err := os.ReadFile("/proc/sys/kernel/unprivileged_userns_clone")
	if err == nil {
		val := strings.TrimSpace(string(data))
		if val == "1" {
			results = append(results, CheckResult{
				Name:    "Linux sandbox: user namespace",
				Status:  StatusPass,
				Message: "unprivileged_userns_clone = 1",
			})
		} else {
			results = append(results, CheckResult{
				Name:    "Linux sandbox: user namespace",
				Status:  StatusFail,
				Message: fmt.Sprintf("unprivileged_userns_clone = %s", val),
				Fix:     "sudo sysctl -w kernel.unprivileged_userns_clone=1",
			})
		}
	} else {
		results = append(results, CheckResult{
			Name:    "Linux sandbox: user namespace",
			Status:  StatusWarn,
			Message: "sysctl kernel.unprivileged_userns_clone not available",
		})
	}

	// bubblewrap チェック
	if _, err := exec.LookPath("bwrap"); err == nil {
		results = append(results, CheckResult{
			Name:    "Linux sandbox: bubblewrap",
			Status:  StatusPass,
			Message: "bwrap found",
		})
	} else {
		results = append(results, CheckResult{
			Name:    "Linux sandbox: bubblewrap",
			Status:  StatusWarn,
			Message: "bwrap not found",
			Fix:     "sudo apt-get install -y bubblewrap",
		})
	}

	return results
}
