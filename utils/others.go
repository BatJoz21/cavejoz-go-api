package utils

import (
	"strconv"

	"github.com/gin-gonic/gin"
)

func GetOffsetForPagination(max int, context *gin.Context) int {
	// Get page
	page, err := strconv.Atoi(context.DefaultQuery("page", "1"))
	if err != nil {
		return 0
	}

	// Count offset
	offset := max * (page - 1)
	return offset
}
