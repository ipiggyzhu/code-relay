package services

import (
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"
	"os"
	"os/exec"
	"path/filepath"
	"runtime"
	"strings"
	"sync"
	"time"
)

// UpdateService 处理应用自动更新
type UpdateService struct {
	currentVersion    string
	repoOwner         string
	repoName          string
	pendingUpdatePath string       // 已下载待安装的更新文件路径
	pendingVersion    string       // 待安装的版本号
	mu                sync.RWMutex // 保护共享状态
	stopChan          chan struct{}
	checkInterval     time.Duration
	autoCheckEnabled  bool
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

// PendingUpdateInfo 待安装更新信息
type PendingUpdateInfo struct {
	HasPendingUpdate bool   `json:"hasPendingUpdate"`
	Version          string `json:"version"`
	FilePath         string `json:"filePath"`
}

// DownloadProgress 下载进度
type DownloadProgress struct {
	Downloaded int64   `json:"downloaded"`
	Total      int64   `json:"total"`
	Percent    float64 `json:"percent"`
}

func NewUpdateService(currentVersion string) *UpdateService {
	return &UpdateService{
		currentVersion:   currentVersion,
		repoOwner:        "ipiggyzhu",
		repoName:         "code-relay",
		stopChan:         make(chan struct{}),
		checkInterval:    1 * time.Hour, // 每小时检查一次
		autoCheckEnabled: true,
	}
}

// Start 启动后台自动更新检测
func (us *UpdateService) Start() error {
	go us.autoCheckLoop()
	return nil
}

// Stop 停止后台检测
func (us *UpdateService) Stop() error {
	close(us.stopChan)
	return nil
}

// autoCheckLoop 后台自动检测循环
func (us *UpdateService) autoCheckLoop() {
	// 启动后延迟 30 秒再开始第一次检测，避免影响启动速度
	select {
	case <-time.After(30 * time.Second):
	case <-us.stopChan:
		return
	}

	// 第一次检测
	us.silentCheckAndDownload()

	// 定期检测
	ticker := time.NewTicker(us.checkInterval)
	defer ticker.Stop()

	for {
		select {
		case <-ticker.C:
			us.silentCheckAndDownload()
		case <-us.stopChan:
			return
		}
	}
}

// silentCheckAndDownload 静默检测并下载更新
func (us *UpdateService) silentCheckAndDownload() {
	// 如果已有待安装的更新，跳过
	us.mu.RLock()
	if us.pendingUpdatePath != "" {
		us.mu.RUnlock()
		return
	}
	us.mu.RUnlock()

	// 检测更新
	info, err := us.CheckForUpdates()
	if err != nil {
		log.Printf("[UpdateService] 检测更新失败: %v", err)
		return
	}

	if !info.HasUpdate || info.DownloadURL == "" {
		return
	}

	log.Printf("[UpdateService] 发现新版本 %s，开始后台下载...", info.LatestVersion)

	// 后台下载
	downloadedPath, err := us.DownloadUpdate(info.DownloadURL)
	if err != nil {
		log.Printf("[UpdateService] 下载更新失败: %v", err)
		return
	}

	// 保存待安装信息
	us.mu.Lock()
	us.pendingUpdatePath = downloadedPath
	us.pendingVersion = info.LatestVersion
	us.mu.Unlock()

	log.Printf("[UpdateService] 更新已下载到 %s，将在程序关闭时自动安装", downloadedPath)
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
	defer out.Close()

	// 复制内容
	_, err = io.Copy(out, resp.Body)
	if err != nil {
		return "", err
	}

	return destPath, nil
}

// GetPendingUpdate 获取待安装的更新信息
func (us *UpdateService) GetPendingUpdate() *PendingUpdateInfo {
	us.mu.RLock()
	defer us.mu.RUnlock()

	return &PendingUpdateInfo{
		HasPendingUpdate: us.pendingUpdatePath != "",
		Version:          us.pendingVersion,
		FilePath:         us.pendingUpdatePath,
	}
}

// HasPendingUpdate 检查是否有待安装的更新
func (us *UpdateService) HasPendingUpdate() bool {
	us.mu.RLock()
	defer us.mu.RUnlock()
	return us.pendingUpdatePath != ""
}

// ApplyPendingUpdate 应用待安装的更新（在程序退出时调用）
func (us *UpdateService) ApplyPendingUpdate() error {
	us.mu.RLock()
	pendingPath := us.pendingUpdatePath
	us.mu.RUnlock()

	if pendingPath == "" {
		return nil // 没有待安装的更新
	}

	return us.InstallUpdate(pendingPath)
}

// SetPendingUpdate 手动设置待安装的更新（用于用户手动下载后）
func (us *UpdateService) SetPendingUpdate(path, version string) {
	us.mu.Lock()
	defer us.mu.Unlock()
	us.pendingUpdatePath = path
	us.pendingVersion = version
}

// ClearPendingUpdate 清除待安装的更新
func (us *UpdateService) ClearPendingUpdate() {
	us.mu.Lock()
	defer us.mu.Unlock()
	us.pendingUpdatePath = ""
	us.pendingVersion = ""
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

	// 启动批处理脚本（在新窗口中运行，这样用户可以看到进度）
	cmd := exec.Command("cmd", "/C", "start", "", batchPath)
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

// createUpdateScript 创建 Windows 更新批处理脚本（静默模式）
func (us *UpdateService) createUpdateScript(currentExe, newExe string) string {
	exeName := filepath.Base(currentExe)

	// 静默更新脚本：等待程序退出 -> 替换文件 -> 重启程序
	return fmt.Sprintf(`@echo off
chcp 65001 >nul
setlocal enabledelayedexpansion

set "NEW_EXE=%s"
set "CURRENT_EXE=%s"
set "EXE_NAME=%s"

:: 等待程序完全退出
:waitloop
timeout /t 1 /nobreak >nul
tasklist /FI "IMAGENAME eq %%EXE_NAME%%" 2>NUL | find /I "%%EXE_NAME%%" >NUL
if not errorlevel 1 (
    goto waitloop
)

:: 尝试直接复制
copy /Y "%%NEW_EXE%%" "%%CURRENT_EXE%%" >nul 2>&1
if not errorlevel 1 (
    goto success
)

:: 如果直接复制失败，尝试使用 PowerShell 提升权限
powershell -Command "Start-Process -FilePath 'cmd.exe' -ArgumentList '/c copy /Y \"%%NEW_EXE%%\" \"%%CURRENT_EXE%%\"' -Verb RunAs -Wait" >nul 2>&1

:success
:: 清理下载的文件
del /Q "%%NEW_EXE%%" >nul 2>&1

:: 启动新版本
start "" "%%CURRENT_EXE%%"
exit /b 0
`, newExe, currentExe, exeName)
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
