//go:build !linux

package doctor

// CheckLinuxSandbox は非Linux環境では何もしない
func CheckLinuxSandbox() []CheckResult {
	return nil
}
