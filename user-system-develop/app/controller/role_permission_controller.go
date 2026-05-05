package controller

import (
	"net/http"

	"user-system/app/dto"
	"user-system/app/exceptions"
	"user-system/app/mapper"
	"user-system/app/service"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
)

type RolePermissionController struct {
	service service.RolePermissionService
}

func NewRolePermissionController(svc service.RolePermissionService) *RolePermissionController {
	return &RolePermissionController{service: svc}
}

func (c *RolePermissionController) AssignPermission(ctx *gin.Context) {
	var req dto.AssignPermissionRequest
	if err := ctx.ShouldBindJSON(&req); err != nil {
		ctx.JSON(http.StatusBadRequest, dto.ErrorResponse{Error: err.Error()})
		return
	}

	roleID, _ := uuid.Parse(req.RoleID)
	permissionID, _ := uuid.Parse(req.PermissionID)

	if err := c.service.AssignPermissionToRole(roleID, permissionID); err != nil {
		c.handleError(ctx, err)
		return
	}

	ctx.JSON(http.StatusOK, dto.SuccessResponse{
		Message: "Permission assigned successfully",
		Data:    dto.RolePermissionAssignmentResponse{RoleID: roleID, PermissionID: permissionID},
	})
}

func (c *RolePermissionController) RemovePermission(ctx *gin.Context) {
	roleID, _ := uuid.Parse(ctx.Param("roleId"))
	permissionID, _ := uuid.Parse(ctx.Param("permissionId"))

	if err := c.service.RemovePermissionFromRole(roleID, permissionID); err != nil {
		c.handleError(ctx, err)
		return
	}

	ctx.JSON(http.StatusOK, dto.SuccessResponse{
		Message: "Permission removed successfully",
		Data:    dto.RolePermissionAssignmentResponse{RoleID: roleID, PermissionID: permissionID},
	})
}

func (c *RolePermissionController) GetRolePermissions(ctx *gin.Context) {
	roleID, _ := uuid.Parse(ctx.Param("roleId"))
	permissions, err := c.service.GetPermissionsForRole(roleID)
	if err != nil {
		c.handleError(ctx, err)
		return
	}

	ctx.JSON(http.StatusOK, dto.SuccessResponse{
		Message: "Role permissions retrieved successfully",
		Data:    mapper.PermissionsToDTOs(permissions),
	})
}

func (c *RolePermissionController) GetPermissionRoles(ctx *gin.Context) {
	permissionID, _ := uuid.Parse(ctx.Param("permissionId"))
	roles, err := c.service.GetRolesWithPermission(permissionID)
	if err != nil {
		c.handleError(ctx, err)
		return
	}

	ctx.JSON(http.StatusOK, dto.SuccessResponse{
		Message: "Permission roles retrieved successfully",
		Data:    mapper.RolesToDTOs(roles),
	})
}

func (c *RolePermissionController) CheckPermission(ctx *gin.Context) {
	roleID, _ := uuid.Parse(ctx.Param("roleId"))
	permissionID, _ := uuid.Parse(ctx.Param("permissionId"))

	exists, err := c.service.CheckPermissionAssignment(roleID, permissionID)
	if err != nil {
		c.handleError(ctx, err)
		return
	}

	ctx.JSON(http.StatusOK, dto.SuccessResponse{
		Message: "Permission check completed",
		Data:    dto.CheckPermissionResponse{Assigned: exists},
	})
}

func (c *RolePermissionController) BulkAssignPermissions(ctx *gin.Context) {
	roleID, _ := uuid.Parse(ctx.Param("roleId"))
	var req dto.BulkPermissionsRequest

	if err := ctx.ShouldBindJSON(&req); err != nil {
		ctx.JSON(http.StatusBadRequest, dto.ErrorResponse{Error: err.Error()})
		return
	}

	permissionIDs := make([]uuid.UUID, len(req.PermissionIDs))
	for i, pid := range req.PermissionIDs {
		permissionIDs[i], _ = uuid.Parse(pid)
	}

	if err := c.service.BulkAssignPermissions(roleID, permissionIDs); err != nil {
		c.handleError(ctx, err)
		return
	}

	ctx.JSON(http.StatusOK, dto.SuccessResponse{Message: "Permissions bulk assigned successfully"})
}

func (c *RolePermissionController) BulkRemovePermissions(ctx *gin.Context) {
	roleID, _ := uuid.Parse(ctx.Param("roleId"))
	var req dto.BulkPermissionsRequest

	if err := ctx.ShouldBindJSON(&req); err != nil {
		ctx.JSON(http.StatusBadRequest, dto.ErrorResponse{Error: err.Error()})
		return
	}

	permissionIDs := make([]uuid.UUID, len(req.PermissionIDs))
	for i, pid := range req.PermissionIDs {
		permissionIDs[i], _ = uuid.Parse(pid)
	}

	if err := c.service.BulkRemovePermissions(roleID, permissionIDs); err != nil {
		c.handleError(ctx, err)
		return
	}

	ctx.JSON(http.StatusOK, dto.SuccessResponse{Message: "Permissions bulk removed successfully"})
}

func (c *RolePermissionController) handleError(ctx *gin.Context, err error) {
	switch e := err.(type) {
	case *exceptions.SystemRoleModificationError:
		ctx.JSON(http.StatusForbidden, dto.ErrorResponse{Error: e.Error()})
	case *exceptions.PermissionNotFoundError:
		ctx.JSON(http.StatusNotFound, dto.ErrorResponse{Error: e.Error()})
	case *exceptions.PermissionAlreadyAssignedError:
		ctx.JSON(http.StatusConflict, dto.ErrorResponse{Error: e.Error()})
	case *exceptions.PermissionNotAssignedError:
		ctx.JSON(http.StatusNotFound, dto.ErrorResponse{Error: e.Error()})
	default:
		ctx.JSON(http.StatusInternalServerError, dto.ErrorResponse{Error: "Internal server error"})
	}
}
