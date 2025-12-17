//go:build !windows

package services

import "os/exec"

// hideWindow 在非 Windows 平台上是空操作
func hideWindow(cmd *exec.Cmd) {
	// no-op on non-Windows platforms
}
