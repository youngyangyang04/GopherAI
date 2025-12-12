package image

import (
	"GopherAI/common/code"
	"GopherAI/controller"
	"GopherAI/service/image"
	"log"
	"net/http"

	"github.com/gin-gonic/gin"
)

type (
	RecognizeImageResponse struct {
		ClassName string `json:"class_name,omitempty"` // AI回答
		controller.Response
	}
)

func RecognizeImage(c *gin.Context) {
	res := new(RecognizeImageResponse)
	fileHeader, err := c.FormFile("image")
	if err != nil {
		log.Println("FormFile fail ", err)
		c.JSON(http.StatusOK, res.CodeOf(code.CodeInvalidParams))
		return
	}

	file, err := fileHeader.Open()
	if err != nil {
		log.Println("Open file fail ", err)
		c.JSON(http.StatusOK, res.CodeOf(code.CodeServerBusy))
		return
	}

	defer file.Close()

	className, err := image.RecognizeImage(c.Request.Context(), file)


	res.Success()
	res.ClassName = className
	c.JSON(http.StatusOK, res)
}
