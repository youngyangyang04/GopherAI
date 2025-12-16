package tools

import (
	"GopherAI/config"
	"context"

	"github.com/cloudwego/eino-ext/components/retriever/volc_vikingdb"
	"github.com/cloudwego/eino/components/tool"
)

var toolsInstance *Tools

type Tools struct {
	googleConfig   *config.GoogleConfig
	vikingdbConfig *config.VikingDBConfig
}

func InitTools() error {
	config := config.GetConfig()
	toolsInstance = &Tools{
		googleConfig:   &config.GoogleConfig,
		vikingdbConfig: &config.VikingDBConfig,
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

// 获取 VikingDB 检索器
func (t *Tools) GetVikingDBRetriever(ctx context.Context) (*volc_vikingdb.Retriever, error) {
	return InitVikingDBRetriever(ctx, t.vikingdbConfig.AK, t.vikingdbConfig.SK, t.vikingdbConfig.Collection, t.vikingdbConfig.Index)
}
