package common

import (
	"context"
	"errors"
	"log"

	"github.com/gin-gonic/gin"
)

// Envelope is the JSON shape returned by every endpoint.
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
	var apiErr *ApiError
	if errors.As(err, &apiErr) {
		if apiErr.Status >= 500 {
			log.Printf("api error (%d): %s", apiErr.Status, apiErr.Message)
		}
		c.JSON(apiErr.Status, Envelope{Success: false, Message: apiErr.Message})
		return
	}
	switch {
	case errors.Is(err, context.DeadlineExceeded):
		c.JSON(504, Envelope{Success: false, Message: "Request timed out"})
		return
	case errors.Is(err, context.Canceled):
		c.JSON(499, Envelope{Success: false, Message: "Request canceled"})
		return
	}
	// Do not leak SQL/provider details to the client, but keep them in the log.
	log.Printf("api error: %v", err)
	c.JSON(500, Envelope{Success: false, Message: "Internal error"})
}
