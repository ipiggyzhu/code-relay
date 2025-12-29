package services

import (
	"encoding/json"
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"strings"
)

const (
	geminiSettingsDir      = ".gemini"
	geminiEnvFileName      = ".env"
	geminiBackupEnvName    = "cc-studio.back.env"
	geminiSettingsFileName = "settings.json"
	geminiBackupSettingsName = "cc-studio.back.settings.json"
	geminiAuthTokenValue   = "code-relay"
)

type GeminiSettingsService struct {
	relayAddr           string
	commonConfigService *CommonConfigService
}

func NewGeminiSettingsService(relayAddr string, commonConfigService *CommonConfigService) *GeminiSettingsService {
	return &GeminiSettingsService{
		relayAddr:           relayAddr,
		commonConfigService: commonConfigService,
	}
}

func (gss *GeminiSettingsService) ProxyStatus() (ClaudeProxyStatus, error) {
	status := ClaudeProxyStatus{Enabled: false, BaseURL: gss.baseURL()}
	envPath, _, err := gss.envPaths()
	if err != nil {
		return status, err
	}
	data, err := os.ReadFile(envPath)
	if err != nil {
		if errors.Is(err, os.ErrNotExist) {
			return status, nil
		}
		return status, err
	}
	content := string(data)
	apiKeyEnabled := strings.Contains(content, "GEMINI_API_KEY="+geminiAuthTokenValue) || 
		strings.Contains(content, "GOOGLE_GEMINI_API_KEY="+geminiAuthTokenValue)
	
	baseURL := gss.baseURL()
	baseURLMatched := strings.Contains(content, "GOOGLE_GEMINI_BASE_URL="+baseURL)
	
	status.Enabled = apiKeyEnabled && baseURLMatched
	return status, nil
}

func (gss *GeminiSettingsService) EnableProxy() error {
	envPath, envBackup, err := gss.envPaths()
	if err != nil {
		return err
	}
	settingsPath, settingsBackup, err := gss.settingsPaths()
	if err != nil {
		return err
	}

	if err := os.MkdirAll(filepath.Dir(envPath), 0o755); err != nil {
		return err
	}

	// 1. 备份并写入 .env
	if _, err := os.Stat(envPath); err == nil {
		content, _ := os.ReadFile(envPath)
		_ = os.WriteFile(envBackup, content, 0o600)
	}

	// 读取通用配置
	commonConfig, err := gss.commonConfigService.GetCommonConfig("gemini")
	if err != nil {
		return err
	}

	var envLines []string
	envLines = append(envLines, "GEMINI_API_KEY="+geminiAuthTokenValue)
	envLines = append(envLines, "GOOGLE_GEMINI_API_KEY="+geminiAuthTokenValue)
	envLines = append(envLines, "GOOGLE_GEMINI_BASE_URL="+gss.baseURL())
	
	// 合并通用配置中的环境变量（如果存在）
	// 注意：gemini 通用配置可能也会包含其他的环境变量
	for k, v := range commonConfig {
		// 简单的将所有配置导出为环境变量
		envLines = append(envLines, fmt.Sprintf("%s=%v", strings.ToUpper(k), v))
	}

	if err := os.WriteFile(envPath, []byte(strings.Join(envLines, "\n")), 0o600); err != nil {
		return err
	}

	// 2. 备份并写入 settings.json 以切换到 api-key 模式
	// 读取现有配置以保留 mcpServers 等其他配置
	existingSettings := make(map[string]interface{})
	if _, err := os.Stat(settingsPath); err == nil {
		content, readErr := os.ReadFile(settingsPath)
		if readErr == nil && len(content) > 0 {
			_ = json.Unmarshal(content, &existingSettings)
		}
		_ = os.WriteFile(settingsBackup, content, 0o600)
	}

	// 设置 security 配置，保留其他配置（如 mcpServers）
	existingSettings["security"] = map[string]interface{}{
		"auth": map[string]interface{}{
			"selectedType": "api-key",
		},
	}
	settingsData, _ := json.MarshalIndent(existingSettings, "", "  ")
	return os.WriteFile(settingsPath, settingsData, 0o600)
}

func (gss *GeminiSettingsService) DisableProxy() error {
	envPath, envBackup, err := gss.envPaths()
	if err != nil {
		return err
	}
	settingsPath, settingsBackup, err := gss.settingsPaths()
	if err != nil {
		return err
	}

	// 还原 .env
	_ = os.Remove(envPath)
	if _, err := os.Stat(envBackup); err == nil {
		_ = os.Rename(envBackup, envPath)
	}

	// 还原 settings.json
	_ = os.Remove(settingsPath)
	if _, err := os.Stat(settingsBackup); err == nil {
		_ = os.Rename(settingsBackup, settingsPath)
	}

	return nil
}

func (gss *GeminiSettingsService) envPaths() (string, string, error) {
	home, err := os.UserHomeDir()
	if err != nil {
		return "", "", err
	}
	dir := filepath.Join(home, geminiSettingsDir)
	return filepath.Join(dir, geminiEnvFileName), filepath.Join(dir, geminiBackupEnvName), nil
}

func (gss *GeminiSettingsService) settingsPaths() (string, string, error) {
	home, err := os.UserHomeDir()
	if err != nil {
		return "", "", err
	}
	dir := filepath.Join(home, geminiSettingsDir)
	return filepath.Join(dir, geminiSettingsFileName), filepath.Join(dir, geminiBackupSettingsName), nil
}

func (gss *GeminiSettingsService) baseURL() string {
	addr := strings.TrimSpace(gss.relayAddr)
	if addr == "" {
		addr = ":18100"
	}
	host := addr
	if strings.HasPrefix(host, ":") {
		host = "127.0.0.1" + host
	}
	if !strings.Contains(host, "://") {
		host = "http://" + host
	}
	return host
}
