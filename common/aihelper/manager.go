package aihelper

import (
	"GopherAI/model"
	"context"
	"sync"
)

var ctx = context.Background()

// AIHelperManager AI助手管理器，管理用户-会话-AIHelper的映射关系
type AIHelperManager struct {
	helpers map[string]map[string]*AIHelper // map[用户账号（唯一）]map[会话ID]*AIHelper
	mu      sync.RWMutex
}

// NewAIHelperManager 创建新的管理器实例
func NewAIHelperManager() *AIHelperManager {
	return &AIHelperManager{
		helpers: make(map[string]map[string]*AIHelper),
	}
}

func (m *AIHelperManager) GetOrCreateAIHelper(
	userName string,
	sessionID string,
	modelType string,
	config map[string]interface{},
	titleOpt ...string, // 可选参数
) (*AIHelper, error) {

	m.mu.Lock()
	defer m.mu.Unlock()

	// 获取用户的会话映射
	userHelpers, exists := m.helpers[userName]
	if !exists {
		userHelpers = make(map[string]*AIHelper)
		m.helpers[userName] = userHelpers
	}

	// 检查会话是否已存在
	helper, exists := userHelpers[sessionID]
	if exists {
		return helper, nil
	}

	// 解析可选 title
	title := ""
	if len(titleOpt) > 0 {
		title = titleOpt[0]
	}

	// 创建新的 AIHelper
	factory := GetGlobalFactory()
	helper, err := factory.CreateAIHelper(ctx, modelType, sessionID, config, title)
	if err != nil {
		return nil, err
	}

	userHelpers[sessionID] = helper
	return helper, nil
}

// 获取指定用户的指定会话的AIHelper
func (m *AIHelperManager) GetAIHelper(userName string, sessionID string) (*AIHelper, bool) {
	m.mu.RLock()
	defer m.mu.RUnlock()

	userHelpers, exists := m.helpers[userName]
	if !exists {
		return nil, false
	}

	helper, exists := userHelpers[sessionID]
	return helper, exists
}

// 移除指定用户的指定会话的AIHelper
func (m *AIHelperManager) RemoveAIHelper(userName string, sessionID string) {
	m.mu.Lock()
	defer m.mu.Unlock()

	userHelpers, exists := m.helpers[userName]
	if !exists {
		return
	}

	delete(userHelpers, sessionID)

	// 如果用户没有会话了，清理用户映射
	if len(userHelpers) == 0 {
		delete(m.helpers, userName)
	}
}

// 获取指定用户的所有会话ID
func (m *AIHelperManager) GetUserSessions(userName string) []model.SessionInfo {
	m.mu.RLock()
	defer m.mu.RUnlock()

	userHelpers, exists := m.helpers[userName]
	if !exists {
		return []model.SessionInfo{}
	}

	sessionIDs := make([]model.SessionInfo, 0, len(userHelpers))
	for sessionID, helper := range userHelpers {
		sessionIDs = append(sessionIDs, model.SessionInfo{
			SessionID: sessionID,
			Title:     helper.Title,
			ModelType: helper.model.GetModelType(),
		})
	}

	return sessionIDs
}

// 全局管理器实例
var globalManager *AIHelperManager
var once sync.Once

// GetGlobalManager 获取全局管理器实例
func GetGlobalManager() *AIHelperManager {
	once.Do(func() {
		globalManager = NewAIHelperManager()
	})
	return globalManager
}
