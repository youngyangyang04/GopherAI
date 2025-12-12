package image

import (
	"GopherAI/config"
	"context"
	"fmt"

	"github.com/cloudwego/eino-ext/components/model/openai"
	"github.com/cloudwego/eino/components/model"
	"github.com/cloudwego/eino/schema"
)

type AIImageModel interface {
	GenerateImageDescription(ctx context.Context, messages []*schema.Message) (*schema.Message, error)
	GetModelType() (string, error)
}

type OpenAIImageModel struct {
	llm model.ToolCallingChatModel
}

// 生成一个新的图像 model
func NewOpenAIImageModel(ctx context.Context) (*OpenAIImageModel, error) {
	config := config.GetConfig()
	key := config.ImageAIConfig.Key
	modelName := config.ImageAIConfig.ModelName
	baseURL := config.ImageAIConfig.BaseURL
	llm, err := openai.NewChatModel(ctx, &openai.ChatModelConfig{
		BaseURL: baseURL,
		Model:   modelName,
		APIKey:  key,
	})
	if err != nil {
		return nil, fmt.Errorf("create openai model failed: %v", err)
	}
	return &OpenAIImageModel{llm: llm}, nil
}

// 使用 model 进行推理
func (o *OpenAIImageModel) GenerateImageDescription(ctx context.Context, messages []*schema.Message) (*schema.Message, error) {
	resp, err := o.llm.Generate(ctx, messages)
	if err != nil {
		return nil, fmt.Errorf("openai generate failed: %v", err)
	}
	return resp, nil
}

// 获取模型类型
func (o *OpenAIImageModel) GetModelType() (string, error) {
	config := config.GetConfig()
	return config.ImageAIConfig.ModelName, nil
}
