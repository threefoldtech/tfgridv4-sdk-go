package server

import "github.com/gin-gonic/gin"

// ResponseMsg holds messages and needed data
type ResponseMsg struct {
	Message string      `json:"msg"`
	Data    interface{} `json:"data,omitempty"`
}

func Response(c *gin.Context, status int, message string, data interface{}) {
	c.JSON(status, ResponseMsg{
		Message: message,
		Data:    data,
	})
}

func AbortResponse(c *gin.Context, status int, message string, data interface{}) {
	c.AbortWithStatusJSON(status, ResponseMsg{
		Message: message,
		Data:    data,
	})
}
