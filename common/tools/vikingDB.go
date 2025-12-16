package tools

import (
	"context"

	"github.com/cloudwego/eino-ext/components/retriever/volc_vikingdb"
)

func ptrOf[T any](v T) *T {
	return &v
}

func InitVikingDBRetriever(
	ctx context.Context,
	AK string,
	SK string,
	collection string,
	index string,
) (*volc_vikingdb.Retriever, error) {
	// 初始化检索器
	return volc_vikingdb.NewRetriever(ctx, &volc_vikingdb.RetrieverConfig{
		Host:           "api-vikingdb.volces.com",
		Region:         "cn-beijing",
		AK:             AK,
		SK:             SK,
		Collection:     collection,
		Index:          index,
		WithMultiModal: true,
		TopK:           ptrOf(3),
	})
}
