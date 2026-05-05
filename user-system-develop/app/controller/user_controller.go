package controller

import (
	"net/http"
	"strconv"

	"user-system/app/dto"
	"user-system/app/service"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
)

type UserController struct {
	userService service.UserService
}

func NewUserController(userService service.UserService) *UserController {
	return &UserController{userService: userService}
}

func (c *UserController) GetUsers(ctx *gin.Context) {
	page, _ := strconv.Atoi(ctx.DefaultQuery("page", "1"))
	limit, _ := strconv.Atoi(ctx.DefaultQuery("limit", "20"))
	companyID := ctx.Query("company_id")
	isActive := ctx.Query("is_active")

	var companyUUID uuid.UUID
	if companyID != "" {
		var err error
		companyUUID, err = uuid.Parse(companyID)
		if err != nil {
			ctx.JSON(http.StatusBadRequest, gin.H{"error": "invalid company ID format"})
			return
		}
	}

	var isActiveBool *bool
	if isActive != "" {
		val := isActive == "true"
		isActiveBool = &val
	}

	users, total, err := c.userService.GetUsers(page, limit, companyUUID, isActiveBool)
	if err != nil {
		ctx.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	ctx.JSON(http.StatusOK, gin.H{
		"data": users,
		"meta": gin.H{
			"total": total,
			"page":  page,
			"limit": limit,
		},
	})
}

func (c *UserController) AdminDashboard(ctx *gin.Context) {
	companyID := ctx.Query("company_id")

	var companyUUID uuid.UUID
	if companyID != "" {
		var err error
		companyUUID, err = uuid.Parse(companyID)
		if err != nil {
			ctx.JSON(http.StatusBadRequest, gin.H{"error": "invalid company ID format"})
			return
		}
	}

	stats, err := c.userService.GetAdminDashboard(companyUUID)
	if err != nil {
		ctx.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	ctx.JSON(http.StatusOK, stats)
}

func (c *UserController) GetUserByID(ctx *gin.Context) {
	userID := ctx.Param("id")
	if userID == "" {
		ctx.JSON(http.StatusBadRequest, gin.H{"error": "user ID is required"})
		return
	}

	id, err := uuid.Parse(userID)
	if err != nil {
		ctx.JSON(http.StatusBadRequest, gin.H{"error": "invalid user ID format"})
		return
	}

	user, err := c.userService.GetUserByID(id)
	if err != nil {
		ctx.JSON(http.StatusNotFound, gin.H{"error": "user not found"})
		return
	}

	ctx.JSON(http.StatusOK, user)
}

func (c *UserController) CreateUser(ctx *gin.Context) {
	var req dto.CreateUserRequest
	if err := ctx.ShouldBindJSON(&req); err != nil {
		ctx.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	currentUserID, exists := ctx.Get("userID")
	if !exists {
		currentUserID = uuid.Nil
	}

	user, err := c.userService.CreateUser(req, currentUserID.(uuid.UUID))
	if err != nil {
		ctx.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	ctx.JSON(http.StatusCreated, user)
}

func (c *UserController) UpdateUser(ctx *gin.Context) {
	userID := ctx.Param("id")
	if userID == "" {
		ctx.JSON(http.StatusBadRequest, gin.H{"error": "user ID is required"})
		return
	}

	id, err := uuid.Parse(userID)
	if err != nil {
		ctx.JSON(http.StatusBadRequest, gin.H{"error": "invalid user ID format"})
		return
	}

	var req dto.UpdateUserRequest
	if err := ctx.ShouldBindJSON(&req); err != nil {
		ctx.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	currentUserID, exists := ctx.Get("userID")
	if !exists {
		ctx.JSON(http.StatusUnauthorized, gin.H{"error": "unauthorized"})
		return
	}

	user, err := c.userService.UpdateUser(id, req, currentUserID.(uuid.UUID))
	if err != nil {
		ctx.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	ctx.JSON(http.StatusOK, user)
}

func (c *UserController) DeleteUser(ctx *gin.Context) {
	userID := ctx.Param("id")
	if userID == "" {
		ctx.JSON(http.StatusBadRequest, gin.H{"error": "user ID is required"})
		return
	}

	id, err := uuid.Parse(userID)
	if err != nil {
		ctx.JSON(http.StatusBadRequest, gin.H{"error": "invalid user ID format"})
		return
	}

	currentUserID, exists := ctx.Get("userID")
	if !exists {
		ctx.JSON(http.StatusUnauthorized, gin.H{"error": "unauthorized"})
		return
	}

	err = c.userService.DeleteUser(id, currentUserID.(uuid.UUID))
	if err != nil {
		ctx.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	ctx.Status(http.StatusNoContent)
}

func (c *UserController) ChangePassword(ctx *gin.Context) {
	userID := ctx.Param("id")
	if userID == "" {
		ctx.JSON(http.StatusBadRequest, gin.H{"error": "user ID is required"})
		return
	}

	id, err := uuid.Parse(userID)
	if err != nil {
		ctx.JSON(http.StatusBadRequest, gin.H{"error": "invalid user ID format"})
		return
	}

	var req dto.ChangePasswordRequest
	if err := ctx.ShouldBindJSON(&req); err != nil {
		ctx.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	currentUserID, exists := ctx.Get("userID")
	if !exists {
		ctx.JSON(http.StatusUnauthorized, gin.H{"error": "unauthorized"})
		return
	}

	err = c.userService.ChangePassword(id, req.OldPassword, req.NewPassword, currentUserID.(uuid.UUID))
	if err != nil {
		ctx.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	ctx.JSON(http.StatusOK, gin.H{"message": "password changed successfully"})
}

func (c *UserController) GetUserProfile(ctx *gin.Context) {
	userID, exists := ctx.Get("userID")
	if !exists {
		ctx.JSON(http.StatusUnauthorized, gin.H{"error": "unauthorized"})
		return
	}

	user, err := c.userService.GetUserByID(userID.(uuid.UUID))
	if err != nil {
		ctx.JSON(http.StatusNotFound, gin.H{"error": "user not found"})
		return
	}

	ctx.JSON(http.StatusOK, user)
}

func (c *UserController) UpdateUserProfile(ctx *gin.Context) {
	userID, exists := ctx.Get("userID")
	if !exists {
		ctx.JSON(http.StatusUnauthorized, gin.H{"error": "unauthorized"})
		return
	}

	var req struct {
		FirstName  string `json:"first_name"`
		LastName   string `json:"last_name"`
		Phone      string `json:"phone"`
		Department string `json:"department"`
		Position   string `json:"position"`
	}

	if err := ctx.ShouldBindJSON(&req); err != nil {
		ctx.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	updateReq := dto.UpdateUserRequest{
		Email:      "",
		FirstName:  req.FirstName,
		LastName:   req.LastName,
		Phone:      req.Phone,
		Department: req.Department,
		Position:   req.Position,
	}

	user, err := c.userService.UpdateUser(userID.(uuid.UUID), updateReq, userID.(uuid.UUID))
	if err != nil {
		ctx.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	ctx.JSON(http.StatusOK, user)
}
