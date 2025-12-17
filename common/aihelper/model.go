package aihelper

import (
	"GopherAI/common/tools"
	"context"
	"fmt"
	"io"
	"log"
	"os"
	"strings"

	"github.com/cloudwego/eino-ext/components/model/ollama"
	"github.com/cloudwego/eino-ext/components/model/openai"
	"github.com/cloudwego/eino/callbacks"
	"github.com/cloudwego/eino/components"
	"github.com/cloudwego/eino/components/model"
	"github.com/cloudwego/eino/components/tool"
	"github.com/cloudwego/eino/compose"
	"github.com/cloudwego/eino/schema"
)

type StreamCallback func(msg string)

// AIModel 定义AI模型接口
type AIModel interface {
	GenerateResponse(ctx context.Context, messages []*schema.Message, opts ...ToolOption) (*schema.Message, error)
	StreamResponse(ctx context.Context, messages []*schema.Message, cb StreamCallback) (string, error)
	GetModelType() string
}

// =================== OpenAI 实现 ===================
type OpenAIModel struct {
	llm model.ToolCallingChatModel
}

// TODO: 重构文件
// 工具选择
type ToolOptions struct {
	usingGoogle bool
	usingRAG    bool
}

func defaultToolOptions() *ToolOptions {
	out := &ToolOptions{
		usingGoogle: false,
		usingRAG:    false,
	}
	return out
}

type ToolOption func(opts *ToolOptions)

func WithGoogleTool() ToolOption {
	return func(opts *ToolOptions) {
		opts.usingGoogle = true
	}
}

func WithRAGTool() ToolOption {
	return func(opts *ToolOptions) {
		opts.usingRAG = true
	}
}

// NOTE: 测试代码
func AddTodoFunc(_ context.Context, params string) (string, error) {
	// Mock处理逻辑
	return `{"msg": "add todo success"}`, nil
}

func NewOpenAIModel(ctx context.Context) (*OpenAIModel, error) {
	key := os.Getenv("OPENAI_API_KEY")
	modelName := os.Getenv("OPENAI_MODEL_NAME")
	baseURL := os.Getenv("OPENAI_BASE_URL")

	llm, err := openai.NewChatModel(ctx, &openai.ChatModelConfig{
		BaseURL: baseURL,
		Model:   modelName,
		APIKey:  key,
	})
	if err != nil {
		return nil, fmt.Errorf("create openai model failed: %v", err)
	}
	return &OpenAIModel{llm: llm}, nil
}

func (o *OpenAIModel) GenerateResponse(ctx context.Context, messages []*schema.Message, opts ...ToolOption) (*schema.Message, error) {
	// 处理可选参数
	var options *ToolOptions
	options = defaultToolOptions()
	for _, opt := range opts {
		opt(options)
	}

	// 使用 Google 能力进行回复
	if options.usingGoogle {
		return o.GenerateResponseWithGoogle(ctx, messages)
	}

	if options.usingRAG {
		return o.GenerateResponseWithRAG(ctx, messages)
	}

	// 如果不需要工具，直接使用 chat 生成回复
	resp, err := o.llm.Generate(ctx, messages)
	if err != nil {
		return nil, fmt.Errorf("openai generate failed: %v", err)
	}
	return resp, nil

}

func (o *OpenAIModel) GenerateResponseWithRAG(ctx context.Context, messages []*schema.Message) (*schema.Message, error) {
	ragTool, err := tools.GetTools().GetVikingDBRetriever(ctx)
	if err != nil {
		return nil, fmt.Errorf("get RAG tool failed: %v", err)
	}

	// 使用 RAG 工具查找资料
	query := messages[len(messages)-1].Content
	docs, err := ragTool.Retrieve(ctx, query)
	if err != nil {
		return nil, fmt.Errorf("RAG tool retrieve failed: %v", err)
	}

	// 将检索到的资料添加到消息中
	for _, doc := range docs {
		messages = append(messages, &schema.Message{
			Role:    schema.Assistant,
			Content: fmt.Sprintf("参考资料：%s", doc.Content),
		})
	}

	// 使用增强后的消息生成回复
	resp, err := o.llm.Generate(ctx, messages)
	if err != nil {
		return nil, fmt.Errorf("openai generate with RAG failed: %v", err)
	}

	// 在 ## 参考资料 ## 部分添加引用信息
	var references strings.Builder
	references.WriteString("## 参考资料 ##\n")
	for i, doc := range docs {
		references.WriteString(fmt.Sprintf("[%d] 文档ID: %s, 内容: %s\n", i+1, doc.ID, doc.Content))
	}
	resp.Content += "\n" + references.String()

	return resp, nil
}

func (o *OpenAIModel) GenerateResponseWithGoogle(ctx context.Context, messages []*schema.Message) (*schema.Message, error) {
	// 接入 Google，使用 Google 能力进行回答
	googleTools, err := tools.GetTools().GetGoogleSearchTool(ctx)
	if err != nil {
		return nil, fmt.Errorf("get google search tool failed: %v", err)
	}
	todoTools := []tool.BaseTool{
		googleTools,
	}

	// 获取工具信息并绑定到 ChatModel
	toolInfos := make([]*schema.ToolInfo, 0, len(todoTools))
	for _, tool := range todoTools {
		info, err := tool.Info(ctx)
		if err != nil {
			log.Fatal(err)
		}
		toolInfos = append(toolInfos, info)
	}

	// 给模型增加工具
	llm_new, err := o.llm.WithTools(toolInfos)
	if err != nil {
		return nil, fmt.Errorf("openai add tools failed: %v", err)
	}

	// 创建 tools 节点
	todoToolsNode, err := compose.NewToolNode(context.Background(), &compose.ToolsNodeConfig{
		Tools: todoTools,
	})
	if err != nil {
		log.Fatal(err)
	}

	// 创建一个回调函数，收集中间消息
	var resp *schema.Message
	handler := callbacks.NewHandlerBuilder().
		OnEndFn(func(ctx context.Context, info *callbacks.RunInfo, output callbacks.CallbackOutput) context.Context {
			if info != nil && info.Component == components.ComponentOfChatModel {
				messages = append(messages, output.(*model.CallbackOutput).Message)
				// 暂时存储第一条回复，如果不需要调用工具则直接返回
				resp = output.(*model.CallbackOutput).Message
				log.Printf("Appended message from chat model to messages: %v", output.(*model.CallbackOutput).Message)
			} else if info != nil && info.Component == components.ComponentOfTool {
				messages = append(messages, &schema.Message{
					Role:       schema.Tool,
					ToolCallID: messages[len(messages)-1].ToolCallID,
					Content:    output.(string),
				})
			}
			return ctx
		}).
		Build()

	// 注册全局回调
	callbacks.AppendGlobalHandlers(handler)

	// 构建完整的处理链
	chain := compose.NewChain[[]*schema.Message, []*schema.Message]()
	chain.
		AppendChatModel(llm_new, compose.WithNodeName("chat_model")).
		AppendToolsNode(todoToolsNode, compose.WithNodeName("tools"))

	// 编译并运行 chain
	agent, err := chain.Compile(ctx)
	if err != nil {
		log.Fatal(err)
	}

	_, err = agent.Invoke(ctx, messages)
	if err != nil {
		log.Printf("chain invoke error: %v", err)
		if resp == nil {
			return nil, fmt.Errorf("openai generate failed: %v", err)
		}
		return resp, nil
		// return nil, fmt.Errorf("openai generate failed: %v", err)
	}

	log.Printf("current msg: %v", messages)
	resp, err = o.llm.Generate(ctx, messages)
	if err != nil {
		return nil, fmt.Errorf("openai generate summary failed: %v", err)
	}

	return resp, nil
}

func (o *OpenAIModel) StreamResponse(ctx context.Context, messages []*schema.Message, cb StreamCallback) (string, error) {
	stream, err := o.llm.Stream(ctx, messages)
	if err != nil {
		return "", fmt.Errorf("openai stream failed: %v", err)
	}
	defer stream.Close()

	var fullResp strings.Builder

	for {
		msg, err := stream.Recv()
		if err == io.EOF {
			break
		}
		if err != nil {
			return "", fmt.Errorf("openai stream recv failed: %v", err)
		}
		if len(msg.Content) > 0 {
			fullResp.WriteString(msg.Content) // 聚合

			cb(msg.Content) // 实时调用cb函数，方便主动发送给前端
		}
	}

	return fullResp.String(), nil //返回完整内容，方便后续存储
}

func (o *OpenAIModel) GetModelType() string { return "openai" }

// =================== Ollama 实现 ===================

// OllamaModel Ollama模型实现
type OllamaModel struct {
	llm model.ToolCallingChatModel
}

func NewOllamaModel(ctx context.Context, baseURL, modelName string) (*OllamaModel, error) {
	llm, err := ollama.NewChatModel(ctx, &ollama.ChatModelConfig{
		BaseURL: baseURL,
		Model:   modelName,
	})
	if err != nil {
		return nil, fmt.Errorf("create ollama model failed: %v", err)
	}
	return &OllamaModel{llm: llm}, nil
}

func (o *OllamaModel) GenerateResponse(ctx context.Context, messages []*schema.Message, opts ...ToolOption) (*schema.Message, error) {
	resp, err := o.llm.Generate(ctx, messages)
	if err != nil {
		return nil, fmt.Errorf("ollama generate failed: %v", err)
	}
	return resp, nil
}

// TODO: 流式响应回调函数存在问题，还是都生成了再统一回复
func (o *OllamaModel) StreamResponse(ctx context.Context, messages []*schema.Message, cb StreamCallback) (string, error) {
	stream, err := o.llm.Stream(ctx, messages)
	if err != nil {
		return "", fmt.Errorf("ollama stream failed: %v", err)
	}
	defer stream.Close()
	var fullResp strings.Builder
	for {
		msg, err := stream.Recv()
		if err == io.EOF {
			break
		}
		if err != nil {
			return "", fmt.Errorf("openai stream recv failed: %v", err)
		}
		if len(msg.Content) > 0 {
			fullResp.WriteString(msg.Content) // 聚合
			cb(msg.Content)                   // 实时调用cb函数，方便主动发送给前端
		}
	}
	return fullResp.String(), nil //返回完整内容，方便后续存储
}

func (o *OllamaModel) GetModelType() string { return "ollama" }
