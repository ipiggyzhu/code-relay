package services

import (
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"
	"net/url"
	"os"
	"os/exec"
	"path/filepath"
	"runtime"
	"strings"
	"time"
)

// UpdateService 处理应用更新检测和安装
type UpdateService struct {
	currentVersion string
	repoOwner      string
	repoName       string
}

// ReleaseInfo GitHub Release 信息
type ReleaseInfo struct {
	TagName string  `json:"tag_name"`
	Assets  []Asset `json:"assets"`
	HTMLURL string  `json:"html_url"`
}

// Asset Release 资源文件
type Asset struct {
	Name               string `json:"name"`
	BrowserDownloadURL string `json:"browser_download_url"`
	Size               int64  `json:"size"`
}

// UpdateInfo 更新信息
type UpdateInfo struct {
	HasUpdate      bool   `json:"hasUpdate"`
	CurrentVersion string `json:"currentVersion"`
	LatestVersion  string `json:"latestVersion"`
	DownloadURL    string `json:"downloadUrl"`
	ReleaseURL     string `json:"releaseUrl"`
	FileName       string `json:"fileName"`
	FileSize       int64  `json:"fileSize"`
}

// DownloadProgress 下载进度
type DownloadProgress struct {
	Downloaded int64   `json:"downloaded"`
	Total      int64   `json:"total"`
	Percent    float64 `json:"percent"`
}

func NewUpdateService(currentVersion string) *UpdateService {
	return &UpdateService{
		currentVersion: currentVersion,
		repoOwner:      "ipiggyzhu",
		repoName:       "code-relay",
	}
}

// CheckForUpdates 检查是否有新版本
func (us *UpdateService) CheckForUpdates() (*UpdateInfo, error) {
	apiURL := fmt.Sprintf("https://api.github.com/repos/%s/%s/releases/latest", us.repoOwner, us.repoName)

	req, err := http.NewRequest("GET", apiURL, nil)
	if err != nil {
		return nil, err
	}
	req.Header.Set("Accept", "application/vnd.github+json")

	client := &http.Client{Timeout: 30 * time.Second}
	resp, err := client.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	if resp.StatusCode == http.StatusForbidden {
		return nil, fmt.Errorf("GitHub API rate limit exceeded, please try again later")
	}
	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("GitHub API returned status %d", resp.StatusCode)
	}

	var release ReleaseInfo
	if err := json.NewDecoder(resp.Body).Decode(&release); err != nil {
		return nil, err
	}

	info := &UpdateInfo{
		CurrentVersion: us.currentVersion,
		LatestVersion:  release.TagName,
		ReleaseURL:     release.HTMLURL,
	}

	// 比较版本
	if us.compareVersions(us.currentVersion, release.TagName) < 0 {
		info.HasUpdate = true

		// 查找对应平台的下载文件
		asset := us.findPlatformAsset(release.Assets)
		if asset != nil {
			info.DownloadURL = asset.BrowserDownloadURL
			info.FileName = asset.Name
			info.FileSize = asset.Size
		} else {
			log.Printf("[UpdateService] 未找到平台 %s/%s 的发布文件，release: %s", runtime.GOOS, runtime.GOARCH, info.ReleaseURL)
		}
	}

	return info, nil
}

// DownloadUpdate 下载更新文件
func (us *UpdateService) DownloadUpdate(downloadURL string) (string, error) {
	// 创建临时目录
	tempDir := filepath.Join(os.TempDir(), "code-relay-update")
	if err := os.MkdirAll(tempDir, 0755); err != nil {
		return "", err
	}

	// 解析 URL 获取文件名（处理可能的查询参数）
	parsedURL, err := url.Parse(downloadURL)
	if err != nil {
		return "", fmt.Errorf("invalid download URL: %w", err)
	}
	fileName := filepath.Base(parsedURL.Path)
	if fileName == "" || fileName == "." || fileName == "/" {
		return "", fmt.Errorf("cannot extract filename from URL: %s", downloadURL)
	}
	destPath := filepath.Join(tempDir, fileName)

	// 下载文件
	client := &http.Client{Timeout: 10 * time.Minute}
	resp, err := client.Get(downloadURL)
	if err != nil {
		return "", err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return "", fmt.Errorf("download failed with status %d", resp.StatusCode)
	}

	// 创建目标文件
	out, err := os.Create(destPath)
	if err != nil {
		return "", err
	}

	// 复制内容
	_, err = io.Copy(out, resp.Body)
	out.Close() // 先关闭文件再判断错误
	if err != nil {
		os.Remove(destPath) // 清理不完整的下载文件
		return "", err
	}

	return destPath, nil
}

// InstallUpdate 安装更新（Windows）
func (us *UpdateService) InstallUpdate(downloadedPath string) error {
	if runtime.GOOS != "windows" {
		return fmt.Errorf("auto-install only supported on Windows")
	}

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
	// 使用 /C 执行命令后关闭cmd
	// start 命令的第一个带引号的参数会被当作窗口标题，所以用空引号 "" 作为标题
	// /WAIT 参数不使用，让脚本在后台运行
	cmd := exec.Command("cmd", "/C", "start", `""`, batchPath)
	cmd.Dir = os.TempDir() // 设置工作目录
	if err := cmd.Start(); err != nil {
		return err
	}

	return nil
}

// GetCurrentExePath 获取当前可执行文件路径
func (us *UpdateService) GetCurrentExePath() (string, error) {
	exePath, err := os.Executable()
	if err != nil {
		return "", err
	}
	return filepath.EvalSymlinks(exePath)
}

// createUpdateScript 创建 Windows 更新批处理脚本
func (us *UpdateService) createUpdateScript(currentExe, newExe string) string {
	exeName := filepath.Base(currentExe)

	// 转义路径中的特殊字符
	// 在批处理中 % 是特殊字符，需要转义为 %%
	// ! 在 enabledelayedexpansion 模式下是特殊字符，需要转义为 ^^!
	currentExeEscaped := strings.ReplaceAll(currentExe, "%", "%%")
	currentExeEscaped = strings.ReplaceAll(currentExeEscaped, "!", "^^!")
	newExeEscaped := strings.ReplaceAll(newExe, "%", "%%")
	newExeEscaped = strings.ReplaceAll(newExeEscaped, "!", "^^!")

	// 更新脚本：显示窗口 -> 等待程序退出 -> 替换文件 -> 重启程序
	// 注意：路径使用双引号包裹，支持包含空格的路径
	return fmt.Sprintf(`@echo off
chcp 65001 >nul
setlocal enabledelayedexpansion
title Code Relay 更新程序

set "NEW_EXE=%s"
set "CURRENT_EXE=%s"
set "EXE_NAME=%s"

echo ========================================
echo        Code Relay 更新程序
echo ========================================
echo.
echo 新版本文件: %%NEW_EXE%%
echo 目标路径: %%CURRENT_EXE%%
echo.

:: 记录更新信息到日志文件
echo [%%date%% %%time%%] 开始更新... >> "%%TEMP%%\code-relay-update.log"
echo [%%date%% %%time%%] 新版本路径: %%NEW_EXE%% >> "%%TEMP%%\code-relay-update.log"
echo [%%date%% %%time%%] 目标路径: %%CURRENT_EXE%% >> "%%TEMP%%\code-relay-update.log"

:: 获取新版本文件大小用于验证
for %%%%A in ("%%NEW_EXE%%") do set "NEW_SIZE=%%%%~zA"
echo [%%date%% %%time%%] 新版本文件大小: !NEW_SIZE! >> "%%TEMP%%\code-relay-update.log"

echo 正在等待 Code Relay 退出...
echo.

:: 等待程序完全退出（最多等待 30 秒）
set /a count=0
:waitloop
timeout /t 1 /nobreak >nul
tasklist /FI "IMAGENAME eq %%EXE_NAME%%" 2>NUL | find /I "%%EXE_NAME%%" >NUL
if not errorlevel 1 (
    set /a count+=1
    if !count! geq 30 (
        echo 等待超时，程序可能未正常退出。
        echo 请手动关闭 Code Relay 后重试。
        echo [%%date%% %%time%%] 等待超时 >> "%%TEMP%%\code-relay-update.log"
        pause
        exit /b 1
    )
    echo 等待中... (!count!/30)
    goto waitloop
)

echo 程序已退出，开始更新...
echo.
echo [%%date%% %%time%%] 程序已退出，开始复制文件... >> "%%TEMP%%\code-relay-update.log"

:: 尝试直接复制
echo 正在复制新版本文件...
copy /Y "%%NEW_EXE%%" "%%CURRENT_EXE%%" >nul 2>&1
if not errorlevel 1 (
    :: 验证复制是否成功（比较文件大小）
    for %%%%A in ("%%CURRENT_EXE%%") do set "COPIED_SIZE=%%%%~zA"
    if "!COPIED_SIZE!"=="!NEW_SIZE!" (
        echo 更新成功！
        echo [%%date%% %%time%%] 直接复制成功，文件大小: !COPIED_SIZE! >> "%%TEMP%%\code-relay-update.log"
        goto success
    )
    echo 复制后文件大小不匹配，尝试提升权限...
    echo [%%date%% %%time%%] 文件大小不匹配: !COPIED_SIZE! vs !NEW_SIZE! >> "%%TEMP%%\code-relay-update.log"
)

echo 直接复制失败，尝试提升权限...
echo [%%date%% %%time%%] 直接复制失败，尝试提升权限... >> "%%TEMP%%\code-relay-update.log"

:: 如果直接复制失败，尝试使用 PowerShell 提升权限
:: 使用单独的批处理文件来执行复制，避免引号转义问题
echo @echo off > "%%TEMP%%\code-relay-copy.bat"
echo copy /Y "%%NEW_EXE%%" "%%CURRENT_EXE%%" >> "%%TEMP%%\code-relay-copy.bat"
powershell -Command "Start-Process -FilePath '%%TEMP%%\code-relay-copy.bat' -Verb RunAs -Wait" >nul 2>&1

:: 验证提升权限复制是否成功（比较文件大小）
for %%%%A in ("%%CURRENT_EXE%%") do set "COPIED_SIZE=%%%%~zA"
if "!COPIED_SIZE!"=="!NEW_SIZE!" (
    echo 提升权限复制成功！
    echo [%%date%% %%time%%] 提升权限复制成功，文件大小: !COPIED_SIZE! >> "%%TEMP%%\code-relay-update.log"
    goto success
)

echo 更新失败！文件大小不匹配。
echo 期望大小: !NEW_SIZE!
echo 实际大小: !COPIED_SIZE!
echo 请手动复制文件。
echo 源文件: %%NEW_EXE%%
echo 目标: %%CURRENT_EXE%%
echo [%%date%% %%time%%] 更新失败，文件大小不匹配 >> "%%TEMP%%\code-relay-update.log"
pause
exit /b 1

:success
echo.
echo ========================================
echo           更新完成！
echo ========================================
echo.

:: 清理下载的文件和临时批处理
del /Q "%%NEW_EXE%%" >nul 2>&1
del /Q "%%TEMP%%\code-relay-copy.bat" >nul 2>&1

echo 正在启动新版本...
echo [%%date%% %%time%%] 启动新版本: %%CURRENT_EXE%% >> "%%TEMP%%\code-relay-update.log"

:: 启动新版本
start "" "%%CURRENT_EXE%%"

echo.
echo 此窗口将在 3 秒后自动关闭...
timeout /t 3 /nobreak >nul
exit /b 0
`, newExeEscaped, currentExeEscaped, exeName)
}

// findPlatformAsset 查找对应平台的资源文件
func (us *UpdateService) findPlatformAsset(assets []Asset) *Asset {
	patterns := us.platformAssetPatterns(runtime.GOOS, runtime.GOARCH)
	return findAssetByPatterns(assets, patterns)
}

// platformAssetPatterns 返回按优先级排序的资产匹配规则
func (us *UpdateService) platformAssetPatterns(goos, goarch string) []string {
	switch goos {
	case "windows":
		return []string{
			"windows-amd64.exe",
			"windows-x86_64.exe",
			"win64.exe",
			"win-amd64.exe",
			"windows.exe",
			"-amd64.exe",
			"_amd64.exe",
		}
	case "darwin":
		if goarch == "arm64" {
			return []string{"darwin-arm64", "macos-arm64", "mac-arm64", "mac_arm64"}
		}
		return []string{"darwin-amd64", "macos-amd64", "mac-amd64", "mac_x64", "mac"}
	case "linux":
		return []string{"linux-amd64", "linux-x86_64", "linux"}
	default:
		return nil
	}
}

// findAssetByPatterns 根据候选模式查找匹配的资产
func findAssetByPatterns(assets []Asset, patterns []string) *Asset {
	if len(assets) == 0 || len(patterns) == 0 {
		return nil
	}

	for _, pattern := range patterns {
		p := strings.ToLower(pattern)
		for i := range assets {
			name := strings.ToLower(assets[i].Name)
			if strings.Contains(name, "installer") || strings.Contains(name, "setup") {
				continue
			}
			if strings.Contains(name, p) {
				return &assets[i]
			}
		}
	}

	// 兜底：仅在 Windows 平台选择非安装器的 .exe 文件
	if runtime.GOOS == "windows" {
		for i := range assets {
			name := strings.ToLower(assets[i].Name)
			if strings.Contains(name, "installer") || strings.Contains(name, "setup") {
				continue
			}
			// 排除其他平台的文件
			if strings.Contains(name, "linux") || strings.Contains(name, "mac") || strings.Contains(name, "darwin") {
				continue
			}
			if strings.HasSuffix(name, ".exe") {
				return &assets[i]
			}
		}
	}

	return nil
}

// compareVersions 比较版本号，返回 -1 (current < remote), 0 (equal), 1 (current > remote)
func (us *UpdateService) compareVersions(current, remote string) int {
	current = strings.TrimPrefix(strings.ToLower(current), "v")
	remote = strings.TrimPrefix(strings.ToLower(remote), "v")

	curParts := strings.Split(current, ".")
	remParts := strings.Split(remote, ".")

	maxLen := len(curParts)
	if len(remParts) > maxLen {
		maxLen = len(remParts)
	}

	for i := 0; i < maxLen; i++ {
		var cur, rem int
		if i < len(curParts) {
			fmt.Sscanf(curParts[i], "%d", &cur)
		}
		if i < len(remParts) {
			fmt.Sscanf(remParts[i], "%d", &rem)
		}

		if cur < rem {
			return -1
		}
		if cur > rem {
			return 1
		}
	}

	return 0
}
