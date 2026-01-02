//go:build !windows

package services

import (
	"fmt"
)

// InstallUpdate 安装更新（非 Windows 平台）
func (us *UpdateService) InstallUpdate(downloadedPath string) error {
	return fmt.Errorf("auto-install only supported on Windows")
}
