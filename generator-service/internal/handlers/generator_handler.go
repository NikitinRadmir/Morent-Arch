package handlers

import (
	"generator-service/internal/generators"
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"
)

type GeneratorHandler struct{}

func NewGeneratorHandler() *GeneratorHandler {
	return &GeneratorHandler{}
}

// GeneratePassword автоматически генерирует пароль
func (h *GeneratorHandler) GeneratePassword(c *gin.Context) {
	password, mask, err := generators.GeneratePasswordAuto()
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"password": password,
		"mask":     mask,
		"length":   len(password),
	})
}

// GeneratePasswordByMask генерирует пароль по маске
func (h *GeneratorHandler) GeneratePasswordByMask(c *gin.Context) {
	mask := c.Query("mask")
	if mask == "" {
		c.JSON(http.StatusBadRequest, gin.H{
			"error":   "mask parameter is required",
			"example": "LLLdddss (L=upper, l=lower, d=digit, s=special)",
		})
		return
	}

	password, err := generators.GeneratePasswordByMask(mask)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"password": password,
		"mask":     mask,
		"length":   len(password),
	})
}

// GenerateQRCode генерирует QR-код
func (h *GeneratorHandler) GenerateQRCode(c *gin.Context) {
	data := c.Query("data")
	if data == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "data parameter is required"})
		return
	}

	sizeStr := c.DefaultQuery("size", "256")
	size, err := strconv.Atoi(sizeStr)
	if err != nil || size < 64 {
		size = 256
	}

	qrCode, err := generators.GenerateQR(data, size)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to generate QR code"})
		return
	}

	c.Data(http.StatusOK, "image/png", qrCode)
}
