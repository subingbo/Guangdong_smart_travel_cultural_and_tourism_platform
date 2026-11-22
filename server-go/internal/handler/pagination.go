package handler

import (
	"strconv"

	"github.com/gin-gonic/gin"

	"github.com/travelgd/server-go/internal/common"
)

// pageParams reads page/pageSize; the service clamps the values.
func pageParams(c *gin.Context) (int, int) {
	page, _ := strconv.Atoi(c.DefaultQuery("page", "1"))
	pageSize, _ := strconv.Atoi(c.DefaultQuery("pageSize", "10"))
	return page, pageSize
}

// idParam parses a positive int64 path parameter or writes a 400.
func idParam(c *gin.Context, name string) (int64, bool) {
	id, err := strconv.ParseInt(c.Param(name), 10, 64)
	if err != nil || id <= 0 {
		common.Fail(c, common.BadRequest("invalid "+name))
		return 0, false
	}
	return id, true
}
