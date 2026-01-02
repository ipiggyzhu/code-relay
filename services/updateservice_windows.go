//go:build windows

package services

import (
	"fmt"
	"log"
	"os"
	"os/exec"
	"path/filepath"
	"syscall"
)

// InstallUpdate 安装更新（Windows 特定实现）
func (us *UpdateService) InstallUpdate(downloadedPath string) error {
	// 验证下载文件是否存在
	if _, err := os.Stat(downloadedPath); os.IsNotExist(err) {
		return fmt.Errorf("downloaded file does not exist: %s", downloadedPath)
	}

	// 获取当前可执行文件路径
	currentExe, err := os.Executable()
	if err != nil {
		return err
	}
	currentExe, err = filepath.EvalSymlinks(currentExe)
	if err != nil {
		return err
	}

	// 记录路径信息，便于调试
	log.Printf("[UpdateService] 当前程序路径: %s", currentExe)
	log.Printf("[UpdateService] 新版本文件路径: %s", downloadedPath)

	// 创建更新批处理脚本
	batchScript := us.createUpdateScript(currentExe, downloadedPath)
	batchPath := filepath.Join(os.TempDir(), "code-relay-update.bat")

	log.Printf("[UpdateService] 更新脚本路径: %s", batchPath)

	if err := os.WriteFile(batchPath, []byte(batchScript), 0755); err != nil {
		return err
	}

	// 启动批处理脚本
	// 使用 CREATE_NEW_CONSOLE (0x10) 在新窗口中运行批处理脚本
	// 不使用 start 命令，避免其复杂的参数解析问题
	cmd := exec.Command("cmd", "/C", batchPath)
	cmd.Dir = os.TempDir()
	cmd.SysProcAttr = &syscall.SysProcAttr{
		CreationFlags: 0x10, // CREATE_NEW_CONSOLE
	}
	if err := cmd.Start(); err != nil {
		return err
	}

	return nil
}
