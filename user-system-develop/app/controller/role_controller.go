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

type RoleController struct {
	roleService service.RoleService
}

func NewRoleController(roleService service.RoleService) *RoleController {
	return &RoleController{roleService: roleService}
}

func (c *RoleController) CreateRole(ctx *gin.Context) {
	var roleDTO dto.CreateRoleRequest
	if err := ctx.ShouldBindJSON(&roleDTO); err != nil {
		ctx.JSON(http.StatusBadRequest, dto.ErrorResponse{Error: err.Error()})
		return
	}

	role := mapper.CreateRequestToModel(roleDTO)

	if err := c.roleService.CreateRole(&role); err != nil {
		c.handleRoleError(ctx, err)
		return
	}

	ctx.JSON(http.StatusCreated, dto.SuccessResponse{
		Message: "Role created successfully",
		Data:    mapper.RoleToDTO(&role),
	})
}

func (c *RoleController) GetRole(ctx *gin.Context) {
	id, err := uuid.Parse(ctx.Param("id"))
	if err != nil {
		ctx.JSON(http.StatusBadRequest, dto.ErrorResponse{Error: "Invalid role ID"})
		return
	}

	role, err := c.roleService.GetRole(id)
	if err != nil {
		c.handleRoleError(ctx, err)
		return
	}

	ctx.JSON(http.StatusOK, dto.SuccessResponse{
		Message: "Role retrieved successfully",
		Data:    mapper.RoleToDTO(role),
	})
}

func (c *RoleController) GetCompanyRoles(ctx *gin.Context) {
	companyID, err := uuid.Parse(ctx.Param("companyId"))
	if err != nil {
		ctx.JSON(http.StatusBadRequest, dto.ErrorResponse{Error: "Invalid company ID"})
		return
	}

	roles, err := c.roleService.GetCompanyRoles(companyID)
	if err != nil {
		c.handleRoleError(ctx, err)
		return
	}

	ctx.JSON(http.StatusOK, dto.SuccessResponse{
		Message: "Company roles retrieved successfully",
		Data:    mapper.RolesToDTOs(roles),
	})
}

func (c *RoleController) UpdateRole(ctx *gin.Context) {
	id, err := uuid.Parse(ctx.Param("id"))
	if err != nil {
		ctx.JSON(http.StatusBadRequest, dto.ErrorResponse{Error: "Invalid role ID"})
		return
	}

	var roleDTO dto.UpdateRoleRequest
	if err := ctx.ShouldBindJSON(&roleDTO); err != nil {
		ctx.JSON(http.StatusBadRequest, dto.ErrorResponse{Error: err.Error()})
		return
	}

	existingRole, err := c.roleService.GetRole(id)
	if err != nil {
		c.handleRoleError(ctx, err)
		return
	}

	if roleDTO.Name != "" {
		existingRole.Name = roleDTO.Name
	}
	if roleDTO.Description != "" {
		existingRole.Description = roleDTO.Description
	}

	if err := c.roleService.UpdateRole(existingRole); err != nil {
		c.handleRoleError(ctx, err)
		return
	}

	ctx.JSON(http.StatusOK, dto.SuccessResponse{
		Message: "Role updated successfully",
		Data:    mapper.RoleToDTO(existingRole),
	})
}

func (c *RoleController) DeleteRole(ctx *gin.Context) {
	id, err := uuid.Parse(ctx.Param("id"))
	if err != nil {
		ctx.JSON(http.StatusBadRequest, dto.ErrorResponse{Error: "Invalid role ID"})
		return
	}

	if err := c.roleService.DeleteRole(id); err != nil {
		c.handleRoleError(ctx, err)
		return
	}

	ctx.Status(http.StatusNoContent)
}

func (c *RoleController) handleRoleError(ctx *gin.Context, err error) {
	switch e := err.(type) {
	case *exceptions.RoleNotFoundError:
		ctx.JSON(http.StatusNotFound, dto.ErrorResponse{Error: e.Error()})
	case *exceptions.CompanyNotFoundError:
		ctx.JSON(http.StatusNotFound, dto.ErrorResponse{Error: e.Error()})
	case *exceptions.RoleAlreadyExistsError:
		ctx.JSON(http.StatusConflict, dto.ErrorResponse{Error: e.Error()})
	case *exceptions.SystemRoleModificationError:
		ctx.JSON(http.StatusForbidden, dto.ErrorResponse{Error: e.Error()})
	case *exceptions.ValidationError:
		ctx.JSON(http.StatusBadRequest, dto.ErrorResponse{Error: e.Error()})
	default:
		ctx.JSON(http.StatusInternalServerError, dto.ErrorResponse{Error: "Internal server error"})
	}
}
