package image

import (
	"GopherAI/utils"
	"bufio"
	"bytes"
	"context"
	"encoding/base64"
	"errors"
	"fmt"
	"image"
	_ "image/gif"
	"image/jpeg"
	_ "image/jpeg"
	_ "image/png"
	"os"
	"path/filepath"
	"strings"
)

// TODO: 换成一个基于 qwen 的能力，因为 onnx 在 Mac 上这个库跑不了

type ImageRecognizer struct {
	model *OpenAIImageModel
}

// NewImageRecognizer 创建识别器（自动使用默认 input/output 名称）
func NewImageRecognizer(ctx context.Context) (*ImageRecognizer, error) {
	ir_llm, err := NewOpenAIImageModel(ctx)
	if err != nil {
		return nil, err
	}
	return &ImageRecognizer{model: ir_llm}, nil
}

func (r *ImageRecognizer) Close() {
	// 目前没有需要关闭的资源
}

func (r *ImageRecognizer) PredictFromFile(ctx context.Context, img image.Image) (string, error) {
	return r.PredictFromImage(ctx, img)
}

func (r *ImageRecognizer) PredictFromBuffer(ctx context.Context, buf []byte) (string, error) {
	img, _, err := image.Decode(bytes.NewReader(buf))
	if err != nil {
		return "", fmt.Errorf("failed to decode image from buffer: %w", err)
	}
	return r.PredictFromImage(ctx, img)
}

func (r *ImageRecognizer) PredictFromImage(ctx context.Context, img image.Image) (string, error) {
	// image.Image -> JPEG
	var buf bytes.Buffer
	if err := jpeg.Encode(&buf, img, &jpeg.Options{Quality: 85}); err != nil {
		return "", err
	}

	// Base64 + data URL
	b64 := base64.StdEncoding.EncodeToString(buf.Bytes())
	dataURL := "data:image/jpeg;base64," + b64

	msg := utils.ConvertToSchemaImageRequests(dataURL)
	om, err := NewOpenAIImageModel(ctx)
	if err != nil {
		return "", err
	}

	// 4. 调用 OpenAI
	resp, err := om.GenerateImageDescription(ctx, msg)
	if err != nil {
		return "", err
	}

	// 5. 取模型输出文本
	result := strings.TrimSpace(resp.Content)
	if result == "" {
		return "", errors.New("empty response from model")
	}

	return result, nil
}

func loadLabels(path string) ([]string, error) {
	f, err := os.Open(filepath.Clean(path))
	if err != nil {
		return nil, fmt.Errorf("open label file failed: %w", err)
	}
	defer f.Close()

	var labels []string
	sc := bufio.NewScanner(f)
	for sc.Scan() {
		line := sc.Text()
		if line != "" {
			labels = append(labels, line)
		}
	}
	if err := sc.Err(); err != nil {
		return nil, fmt.Errorf("read labels failed: %w", err)
	}
	if len(labels) == 0 {
		return nil, fmt.Errorf("no labels found in %s", path)
	}
	return labels, nil
}
