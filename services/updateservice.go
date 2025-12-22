package services

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"
	"os/exec"
	"path/filepath"
	"runtime"
	"strings"
)

// UpdateService 处理应用自动更新
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

func (us *UpdateService) Start() error { return nil }
func (us *UpdateService) Stop() error  { return nil }

// CheckForUpdates 检查是否有新版本
func (us *UpdateService) CheckForUpdates() (*UpdateInfo, error) {
	apiURL := fmt.Sprintf("https://api.github.com/repos/%s/%s/releases/latest", us.repoOwner, us.repoName)

	req, err := http.NewRequest("GET", apiURL, nil)
	if err != nil {
		return nil, err
	}
	req.Header.Set("Accept", "application/vnd.github+json")

	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

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

	// 获取文件名
	fileName := filepath.Base(downloadURL)
	destPath := filepath.Join(tempDir, fileName)

	// 下载文件
	resp, err := http.Get(downloadURL)
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
	defer out.Close()

	// 复制内容
	_, err = io.Copy(out, resp.Body)
	if err != nil {
		return "", err
	}

	return destPath, nil
}

// InstallUpdate 安装更新（Windows）
func (us *UpdateService) InstallUpdate(downloadedPath string) error {
	if runtime.GOOS != "windows" {
		return fmt.Errorf("auto-install only supported on Windows")
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

	// 创建更新批处理脚本
	batchScript := us.createUpdateScript(currentExe, downloadedPath)
	batchPath := filepath.Join(os.TempDir(), "code-relay-update.bat")

	if err := os.WriteFile(batchPath, []byte(batchScript), 0755); err != nil {
		return err
	}

	// 启动批处理脚本（在后台运行）
	cmd := exec.Command("cmd", "/C", "start", "/B", batchPath)
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
	installDir := filepath.Dir(currentExe)

	// 使用 PowerShell 脚本来处理更新，支持管理员权限提升
	return fmt.Sprintf(`@echo off
chcp 65001 >nul
setlocal enabledelayedexpansion

echo ========================================
echo   Code Relay 自动更新
echo ========================================
echo.

set "NEW_EXE=%s"
set "CURRENT_EXE=%s"
set "INSTALL_DIR=%s"
set "EXE_NAME=%s"

echo 等待程序退出...
timeout /t 2 /nobreak >nul

:waitloop
tasklist /FI "IMAGENAME eq %%EXE_NAME%%" 2>NUL | find /I "%%EXE_NAME%%" >NUL
if not errorlevel 1 (
    echo 程序仍在运行，等待中...
    timeout /t 1 /nobreak >nul
    goto waitloop
)

echo 程序已退出，开始更新...
echo.

:: 尝试直接复制
echo 正在复制文件...
copy /Y "%%NEW_EXE%%" "%%CURRENT_EXE%%" >nul 2>&1
if not errorlevel 1 (
    echo 更新成功！
    goto success
)

:: 如果直接复制失败，尝试使用 PowerShell 提升权限
echo 需要管理员权限，正在请求...
powershell -Command "Start-Process -FilePath 'cmd.exe' -ArgumentList '/c copy /Y \"%s\" \"%s\" && echo 更新成功' -Verb RunAs -Wait" 2>nul
if errorlevel 1 (
    echo.
    echo ========================================
    echo   更新失败！
    echo ========================================
    echo.
    echo 请手动将以下文件复制到安装目录：
    echo   源文件: %%NEW_EXE%%
    echo   目标: %%CURRENT_EXE%%
    echo.
    pause
    exit /b 1
)

:success
echo.
echo ========================================
echo   更新完成！
echo ========================================
echo.
echo 正在启动新版本...
timeout /t 1 /nobreak >nul
start "" "%%CURRENT_EXE%%"
exit /b 0
`, newExe, currentExe, installDir, exeName, newExe, currentExe)
}

// findPlatformAsset 查找对应平台的资源文件
func (us *UpdateService) findPlatformAsset(assets []Asset) *Asset {
	var targetSuffix string

	switch runtime.GOOS {
	case "windows":
		targetSuffix = "windows-amd64.exe"
	case "darwin":
		if runtime.GOARCH == "arm64" {
			targetSuffix = "darwin-arm64"
		} else {
			targetSuffix = "darwin-amd64"
		}
	case "linux":
		targetSuffix = "linux-amd64"
	}

	for i := range assets {
		name := strings.ToLower(assets[i].Name)
		if strings.Contains(name, strings.ToLower(targetSuffix)) {
			return &assets[i]
		}
		// 备选：简单匹配
		if runtime.GOOS == "windows" && strings.HasSuffix(name, ".exe") {
			return &assets[i]
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
