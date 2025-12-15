package tools

import (
	"context"
	"log"

	"github.com/cloudwego/eino-ext/components/tool/googlesearch"
	"github.com/cloudwego/eino/components/tool"
)

func InitGoogleSearchTool(ctx context.Context, apiKey, searchEngineID string) (tool.InvokableTool, error) {
	googleSearchTool, err := googlesearch.NewTool(ctx, &googlesearch.Config{
		APIKey:         apiKey,
		SearchEngineID: searchEngineID,
	})
	// 异常处理
	if err != nil {
		log.Println("InitGoogleSearchTool NewTool error:", err)
		return nil, err
	}

	return googleSearchTool, nil
}
