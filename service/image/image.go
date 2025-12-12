package image

import (
	image_recognizer "GopherAI/common/image"
	"context"
	"image"
	"io"
)

func RecognizeImage(ctx context.Context, r io.Reader) (string, error) {
	img, _, err := image.Decode(r)
	if err != nil {
		return "", err
	}

	// TODO: 目前去除了 Buffer 之后补全
	ri, err := image_recognizer.NewImageRecognizer(ctx)
	if err != nil {
		return "", err
	}
	defer ri.Close()

	return ri.PredictFromFile(ctx, img)
}
