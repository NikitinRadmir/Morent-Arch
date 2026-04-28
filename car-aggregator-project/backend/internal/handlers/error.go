package handlers

import (
	"net/http"

	"github.com/gin-gonic/gin"

	"car-aggregator/internal/errors"
)

func RenderError(c *gin.Context, err error) {
	if de, ok := errors.IsDomain(err); ok {
		c.JSON(statusFromDomain(de.Code), gin.H{
			"error": gin.H{
				"code":    string(de.Code),
				"message": de.Message,
			},
		})
		return
	}

	c.JSON(http.StatusBadRequest, gin.H{
		"error": gin.H{
			"code":    string(errors.CodeValidation),
			"message": "invalid request",
		},
	})
}

func statusFromDomain(code errors.Code) int {
	switch code {
	case errors.CodeNotFound:
		return http.StatusNotFound
	default:
		return http.StatusBadRequest
	}
}
