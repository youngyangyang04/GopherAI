package aihelper

import (
	myconfig "GopherAI/config"
	"context"
	"fmt"
	"log"
	"sync"
)

// ModelCreator 定义模型创建函数类型（需要 context）
type ModelCreator func(ctx context.Context, config map[string]interface{}) (AIModel, error)

// AIModelFactory AI模型工厂
type AIModelFactory struct {
	creators map[string]ModelCreator
}

var (
	globalFactory *AIModelFactory
	factoryOnce   sync.Once
)

// GetGlobalFactory 获取全局单例
func GetGlobalFactory() *AIModelFactory {
	factoryOnce.Do(func() {
		globalFactory = &AIModelFactory{
			creators: make(map[string]ModelCreator),
		}
		globalFactory.registerCreators()
	})
	return globalFactory
}

// 注册模型
// FIXME: 整体封装有问题，config 没有使用到，而是直接读取了配置文件或环境变量
func (f *AIModelFactory) registerCreators() {
	//OpenAI
	f.creators["1"] = func(ctx context.Context, config map[string]interface{}) (AIModel, error) {
		return NewOpenAIModel(ctx)
	}

	//Ollama
	f.creators["2"] = func(ctx context.Context, config map[string]interface{}) (AIModel, error) {
		baseURL := myconfig.GetConfig().OllamaConfig.BaseURL
		modelName := myconfig.GetConfig().OllamaConfig.ModelName
		if baseURL == "" || modelName == "" {
			return nil, fmt.Errorf("Ollama model requires baseURL and modelName in config")
		}
		return NewOllamaModel(ctx, baseURL, modelName)
	}
}

// CreateAIModel 根据类型创建 AI 模型
func (f *AIModelFactory) CreateAIModel(ctx context.Context, modelType string, config map[string]interface{}) (AIModel, error) {
	creator, ok := f.creators[modelType]
	if !ok {
		return nil, fmt.Errorf("unsupported model type: %s", modelType)
	}
	return creator(ctx, config)
}

// CreateAIHelper 一键创建 AIHelper
func (f *AIModelFactory) CreateAIHelper(ctx context.Context, modelType string, SessionID string, config map[string]interface{}, title string) (*AIHelper, error) {
	log.Printf("Creating AIHelper with modelType: %s, SessionID: %s", modelType, SessionID)
	model, err := f.CreateAIModel(ctx, modelType, config)
	if err != nil {
		return nil, err
	}
	return NewAIHelper(model, SessionID, title), nil
}

// RegisterModel 可扩展注册
func (f *AIModelFactory) RegisterModel(modelType string, creator ModelCreator) {
	f.creators[modelType] = creator
}
