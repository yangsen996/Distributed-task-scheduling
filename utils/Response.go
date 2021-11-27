package utils

import (
	"net/http"

	"github.com/gin-gonic/gin"
)

type ResponseData struct {
	Code ResCode     `json:"code"`
	Data interface{} `json:"data"`
	Msg  interface{} `json:"msg"`
}

func ResponseSuccess(c *gin.Context, data interface{}) {
	da := &ResponseData{
		Code: CodeSuccess,
		Data: data,
		Msg:  CodeSuccess.Msg(),
	}
	c.JSON(http.StatusOK, da)
}

func ResponseErr(c *gin.Context, code ResCode) {
	data := &ResponseData{
		Code: code,
		Data: nil,
		Msg:  code.Msg(),
	}
	c.JSON(http.StatusOK, data)
}
