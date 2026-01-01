package services

import (
	"encoding/json"
	"errors"
	"os"
	"path/filepath"
	"sort"
	"strings"
	"sync"
	"time"

	"github.com/google/uuid"
)

const (
	promptStoreDir  = ".code-relay"
	promptStoreFile = "prompts.json"
)

// Prompt 提示词数据结构
type Prompt struct {
	ID        string    `json:"id"`
	Name      string    `json:"name"`
	Content   string    `json:"content"`
	Platform  string    `json:"platform"` // claude, codex, gemini
	IsActive  bool      `json:"is_active"`
	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`
}

// promptStore 提示词存储结构
type promptStore struct {
	Prompts []Prompt `json:"prompts"`
}

// PromptService 提示词管理服务
type PromptService struct {
	storePath string
	mu        sync.Mutex
}

// NewPromptService 创建提示词服务实例
func NewPromptService() *PromptService {
	home, err := os.UserHomeDir()
	if err != nil {
		home = "."
	}
	return &PromptService{
		storePath: filepath.Join(home, promptStoreDir, promptStoreFile),
	}
}

func (ps *PromptService) Start() error { return nil }
func (ps *PromptService) Stop() error  { return nil }

// ListPrompts 获取指定平台的所有提示词
func (ps *PromptService) ListPrompts(platform string) ([]Prompt, error) {
	ps.mu.Lock()
	defer ps.mu.Unlock()

	store, err := ps.loadStoreLocked()
	if err != nil {
		return nil, err
	}

	platform = strings.ToLower(strings.TrimSpace(platform))
	var result []Prompt
	for _, prompt := range store.Prompts {
		if platform == "" || strings.EqualFold(prompt.Platform, platform) {
			result = append(result, prompt)
		}
	}

	// 按更新时间倒序排列
	sort.SliceStable(result, func(i, j int) bool {
		return result[i].UpdatedAt.After(result[j].UpdatedAt)
	})

	return result, nil
}

// GetPrompt 获取单个提示词
func (ps *PromptService) GetPrompt(id string) (Prompt, error) {
	ps.mu.Lock()
	defer ps.mu.Unlock()

	store, err := ps.loadStoreLocked()
	if err != nil {
		return Prompt{}, err
	}

	for _, prompt := range store.Prompts {
		if prompt.ID == id {
			return prompt, nil
		}
	}

	return Prompt{}, errors.New("提示词不存在")
}

// CreatePrompt 创建新提示词
func (ps *PromptService) CreatePrompt(prompt Prompt) (Prompt, error) {
	ps.mu.Lock()
	defer ps.mu.Unlock()

	store, err := ps.loadStoreLocked()
	if err != nil {
		return Prompt{}, err
	}

	// 生成 ID 和时间戳
	prompt.ID = uuid.New().String()
	prompt.CreatedAt = time.Now()
	prompt.UpdatedAt = time.Now()
	prompt.Platform = normalizePromptPlatform(prompt.Platform)
	prompt.IsActive = false

	store.Prompts = append(store.Prompts, prompt)

	if err := ps.saveStoreLocked(store); err != nil {
		return Prompt{}, err
	}

	return prompt, nil
}

// UpdatePrompt 更新提示词
func (ps *PromptService) UpdatePrompt(prompt Prompt) error {
	ps.mu.Lock()
	defer ps.mu.Unlock()

	store, err := ps.loadStoreLocked()
	if err != nil {
		return err
	}

	found := false
	for i, p := range store.Prompts {
		if p.ID == prompt.ID {
			prompt.CreatedAt = p.CreatedAt
			prompt.UpdatedAt = time.Now()
			prompt.Platform = normalizePromptPlatform(prompt.Platform)
			prompt.IsActive = p.IsActive // 保持激活状态不变
			store.Prompts[i] = prompt
			found = true

			// 如果是激活状态，同步到文件
			if prompt.IsActive {
				if err := ps.syncToFile(prompt); err != nil {
					return err
				}
			}
			break
		}
	}

	if !found {
		return errors.New("提示词不存在")
	}

	return ps.saveStoreLocked(store)
}

// DeletePrompt 删除提示词
func (ps *PromptService) DeletePrompt(id string) error {
	ps.mu.Lock()
	defer ps.mu.Unlock()

	store, err := ps.loadStoreLocked()
	if err != nil {
		return err
	}

	newPrompts := make([]Prompt, 0, len(store.Prompts))
	var deletedPrompt *Prompt
	for _, p := range store.Prompts {
		if p.ID == id {
			deletedPrompt = &p
			continue
		}
		newPrompts = append(newPrompts, p)
	}

	if deletedPrompt == nil {
		return errors.New("提示词不存在")
	}

	store.Prompts = newPrompts

	// 如果删除的是激活状态的提示词，清空对应文件
	if deletedPrompt.IsActive {
		if err := ps.clearFile(deletedPrompt.Platform); err != nil {
			return err
		}
	}

	return ps.saveStoreLocked(store)
}

// ActivatePrompt 激活提示词
func (ps *PromptService) ActivatePrompt(id string) error {
	ps.mu.Lock()
	defer ps.mu.Unlock()

	store, err := ps.loadStoreLocked()
	if err != nil {
		return err
	}

	var targetPrompt *Prompt
	var targetIndex int
	for i := range store.Prompts {
		if store.Prompts[i].ID == id {
			targetPrompt = &store.Prompts[i]
			targetIndex = i
			break
		}
	}

	if targetPrompt == nil {
		return errors.New("提示词不存在")
	}

	// 取消同平台其他提示词的激活状态
	for i := range store.Prompts {
		if strings.EqualFold(store.Prompts[i].Platform, targetPrompt.Platform) {
			store.Prompts[i].IsActive = false
		}
	}

	// 激活目标提示词
	store.Prompts[targetIndex].IsActive = true

	// 同步到文件
	if err := ps.syncToFile(*targetPrompt); err != nil {
		return err
	}

	return ps.saveStoreLocked(store)
}

// DeactivatePrompt 取消激活提示词
func (ps *PromptService) DeactivatePrompt(id string) error {
	ps.mu.Lock()
	defer ps.mu.Unlock()

	store, err := ps.loadStoreLocked()
	if err != nil {
		return err
	}

	var targetPrompt *Prompt
	for i := range store.Prompts {
		if store.Prompts[i].ID == id {
			targetPrompt = &store.Prompts[i]
			store.Prompts[i].IsActive = false
			break
		}
	}

	if targetPrompt == nil {
		return errors.New("提示词不存在")
	}

	// 清空对应文件
	if err := ps.clearFile(targetPrompt.Platform); err != nil {
		return err
	}

	return ps.saveStoreLocked(store)
}

// GetActivePrompt 获取当前激活的提示词
func (ps *PromptService) GetActivePrompt(platform string) (*Prompt, error) {
	ps.mu.Lock()
	defer ps.mu.Unlock()

	store, err := ps.loadStoreLocked()
	if err != nil {
		return nil, err
	}

	platform = normalizePromptPlatform(platform)
	for _, prompt := range store.Prompts {
		if strings.EqualFold(prompt.Platform, platform) && prompt.IsActive {
			return &prompt, nil
		}
	}

	return nil, nil
}

// 内部方法

func (ps *PromptService) loadStoreLocked() (promptStore, error) {
	data, err := os.ReadFile(ps.storePath)
	if err != nil {
		if os.IsNotExist(err) {
			return promptStore{Prompts: []Prompt{}}, nil
		}
		return promptStore{}, err
	}

	if len(data) == 0 {
		return promptStore{Prompts: []Prompt{}}, nil
	}

	var store promptStore
	if err := json.Unmarshal(data, &store); err != nil {
		return promptStore{Prompts: []Prompt{}}, err
	}

	if store.Prompts == nil {
		store.Prompts = []Prompt{}
	}

	return store, nil
}

func (ps *PromptService) saveStoreLocked(store promptStore) error {
	if err := os.MkdirAll(filepath.Dir(ps.storePath), 0o755); err != nil {
		return err
	}

	data, err := json.MarshalIndent(store, "", "  ")
	if err != nil {
		return err
	}

	tmp := ps.storePath + ".tmp"
	if err := os.WriteFile(tmp, data, 0o644); err != nil {
		return err
	}

	return os.Rename(tmp, ps.storePath)
}

func (ps *PromptService) syncToFile(prompt Prompt) error {
	targetPath := ps.getTargetPath(prompt.Platform)
	if targetPath == "" {
		return nil
	}

	if err := os.MkdirAll(filepath.Dir(targetPath), 0o755); err != nil {
		return err
	}

	return os.WriteFile(targetPath, []byte(prompt.Content), 0o644)
}

func (ps *PromptService) clearFile(platform string) error {
	targetPath := ps.getTargetPath(platform)
	if targetPath == "" {
		return nil
	}

	// 只是清空内容，不删除文件
	if _, err := os.Stat(targetPath); err == nil {
		return os.WriteFile(targetPath, []byte{}, 0o644)
	}

	return nil
}

func (ps *PromptService) getTargetPath(platform string) string {
	home, err := os.UserHomeDir()
	if err != nil {
		return ""
	}

	switch strings.ToLower(platform) {
	case "claude":
		return filepath.Join(home, ".claude", "CLAUDE.md")
	case "codex":
		return filepath.Join(home, ".codex", "AGENTS.md")
	case "gemini":
		return filepath.Join(home, ".gemini", "GEMINI.md")
	default:
		return ""
	}
}

func normalizePromptPlatform(platform string) string {
	platform = strings.ToLower(strings.TrimSpace(platform))
	switch platform {
	case "claude", "claude-code", "claude_code":
		return "claude"
	case "codex", "openai":
		return "codex"
	case "gemini", "google":
		return "gemini"
	default:
		return "claude" // 默认 claude
	}
}
