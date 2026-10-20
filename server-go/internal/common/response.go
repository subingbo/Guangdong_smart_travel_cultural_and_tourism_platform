package common

import "github.com/gin-gonic/gin"

/** Unified response envelope: { success, data, message }. */
type Envelope struct {
	Success bool        `json:"success"`
	Data    interface{} `json:"data"`
	Message string      `json:"message"`
}

func OK(c *gin.Context, data interface{}) {
	if data == nil {
		c.JSON(200, Envelope{Success: true, Message: ""})
		return
	}
	c.JSON(200, Envelope{Success: true, Data: data})
}

func Created(c *gin.Context, data interface{}) {
	c.JSON(201, Envelope{Success: true, Data: data})
}

func Fail(c *gin.Context, err error) {
	if apiErr, ok := err.(*ApiError); ok {
		c.JSON(apiErr.Status, Envelope{Success: false, Message: apiErr.Message})
		return
	}
	c.JSON(500, Envelope{Success: false, Message: "Internal error: " + err.Error()})
}
