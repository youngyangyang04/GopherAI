package tools

import (
	"GopherAI/config"
	"context"

	"github.com/cloudwego/eino/components/tool"
)

var toolsInstance *Tools

type Tools struct {
	googleConfig *config.GoogleConfig
}

func InitTools() error {
	config := config.GetConfig()
	toolsInstance = &Tools{
		googleConfig: &config.GoogleConfig,
	}
	return nil
}

// 获取工具实例
func GetTools() *Tools {
	return toolsInstance
}

// 获取 Google Search 工具
func (t *Tools) GetGoogleSearchTool(ctx context.Context) (tool.InvokableTool, error) {
	return InitGoogleSearchTool(ctx, t.googleConfig.GoogleAPIKey, t.googleConfig.GoogleSearchEngineID)
}
