package utils

import (
	"GopherAI/model"
	"crypto/md5"
	"encoding/hex"
	"math/rand"
	"strconv"
	"time"

	"github.com/cloudwego/eino/schema"
	"github.com/google/uuid"
)

// GetRandomNumbers 生成num位随机数字字符串
func GetRandomNumbers(num int) string {
	r := rand.New(rand.NewSource(time.Now().UnixNano()))

	code := ""
	for i := 0; i < num; i++ {
		// 0~9随机数
		digit := r.Intn(10)
		code += strconv.Itoa(digit)
	}
	return code
}

// MD5加密
func MD5(str string) string {
	m := md5.New()
	m.Write([]byte(str))
	return hex.EncodeToString(m.Sum(nil))
}

func GenerateUUID() string {
	return uuid.New().String()
}

// 将 schema 消息转换为数据库可存储的格式
func ConvertToModelMessage(sessionID string, userName string, msg *schema.Message) *model.Message {
	return &model.Message{
		SessionID: sessionID,
		UserName:  userName,
		Content:   msg.Content,
	}
}

// 将数据库消息转换为 schema 消息（供 AI 使用）
func ConvertToSchemaMessages(msgs []*model.Message) []*schema.Message {
	schemaMsgs := make([]*schema.Message, 0, len(msgs))
	for _, m := range msgs {
		role := schema.Assistant
		if m.IsUser == false {
			role = schema.User
		}
		schemaMsgs = append(schemaMsgs, &schema.Message{
			Role:    role,
			Content: m.Content,
		})
	}
	return schemaMsgs
}

func of[T any](a T) *T {
	return &a
}

// 转换图片请求信息到 schema 格式
func ConvertToSchemaImageRequests(b64OrDataURL string) []*schema.Message {
	// msg := []*schema.Message{
	// 	{
	// 		Role:    "user",
	// 		Content: "请进行简单的自我介绍",
	// 	},
	// }
	msg := &schema.Message{
		Role: schema.User,
		UserInputMultiContent: []schema.MessageInputPart{
			{
				Type: schema.ChatMessagePartTypeText,
				Text: "帮我对照片进行简单描述",
			},
			{
				Type: schema.ChatMessagePartTypeImageURL,
				Image: &schema.MessageInputImage{
					MessagePartCommon: schema.MessagePartCommon{
						Base64Data: &b64OrDataURL,
						MIMEType:   "image/jpeg",
					},
					Detail: schema.ImageURLDetailAuto,
				},
			},
		},
	}
	return []*schema.Message{msg}
}
