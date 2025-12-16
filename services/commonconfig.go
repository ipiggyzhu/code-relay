package services

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"strings"
)

// CommonConfigService 管理 Claude Code 和 Codex 的通用配置
type CommonConfigService struct{}

// NewCommonConfigService 创建通用配置服务实例
func NewCommonConfigService() *CommonConfigService {
	return &CommonConfigService{}
}

func (ccs *CommonConfigService) Start() error { return nil }
func (ccs *CommonConfigService) Stop() error  { return nil }

// getConfigPath 获取通用配置文件路径
func (ccs *CommonConfigService) getConfigPath(kind string) (string, error) {
	home, err := os.UserHomeDir()
	if err != nil {
		return "", err
	}

	dir := filepath.Join(home, ".code-switch")
	if err := os.MkdirAll(dir, 0o755); err != nil {
		return "", err
	}

	var filename string
	switch strings.ToLower(kind) {
	case "claude", "claude-code", "claude_code":
		filename = "claude-common-config.json"
	case "codex":
		filename = "codex-common-config.json"
	default:
		return "", fmt.Errorf("unknown config type: %s (expected 'claude' or 'codex')", kind)
	}

	return filepath.Join(dir, filename), nil
}

// GetCommonConfig 获取通用配置
func (ccs *CommonConfigService) GetCommonConfig(kind string) (map[string]interface{}, error) {
	path, err := ccs.getConfigPath(kind)
	if err != nil {
		return nil, err
	}

	data, err := os.ReadFile(path)
	if err != nil {
		if os.IsNotExist(err) {
			// 文件不存在，返回空配置
			return make(map[string]interface{}), nil
		}
		return nil, err
	}

	if len(data) == 0 {
		return make(map[string]interface{}), nil
	}

	var config map[string]interface{}
	if err := json.Unmarshal(data, &config); err != nil {
		return nil, fmt.Errorf("failed to parse config: %w", err)
	}

	return config, nil
}

// SaveCommonConfig 保存通用配置（原子写入）
func (ccs *CommonConfigService) SaveCommonConfig(kind string, config map[string]interface{}) error {
	path, err := ccs.getConfigPath(kind)
	if err != nil {
		return err
	}

	// 序列化配置
	data, err := json.MarshalIndent(config, "", "  ")
	if err != nil {
		return fmt.Errorf("failed to serialize config: %w", err)
	}

	// 原子写入（先写临时文件，再重命名）
	tmp := path + ".tmp"
	if err := os.WriteFile(tmp, data, 0o644); err != nil {
		return fmt.Errorf("failed to write temp file: %w", err)
	}

	if err := os.Rename(tmp, path); err != nil {
		os.Remove(tmp) // 清理临时文件
		return fmt.Errorf("failed to rename temp file: %w", err)
	}

	return nil
}
