package rpc

import (
	"user-system/app/dto"
	"user-system/app/exceptions"
	"user-system/app/service"
	"encoding/json"

	"github.com/google/uuid"
)

type RPCMethods struct {
	companyService    service.CompanyService
	userService       service.UserService
	roleService       service.RoleService
	permissionService service.RolePermissionService
}

func NewRPCMethods(
	companyService service.CompanyService,
	userService service.UserService,
	roleService service.RoleService,
	permissionService service.RolePermissionService,
) *RPCMethods {
	return &RPCMethods{
		companyService:    companyService,
		userService:       userService,
		roleService:       roleService,
		permissionService: permissionService,
	}
}

func (m *RPCMethods) GetCompanyHierarchy(params json.RawMessage) (interface{}, error) {
	var p struct {
		CompanyID string `json:"company_id"`
	}

	if err := json.Unmarshal(params, &p); err != nil {
		return nil, &RPCError{Code: -32602, Message: "Invalid params: " + err.Error()}
	}

	if p.CompanyID == "" {
		return nil, &RPCError{Code: -32602, Message: "company_id is required"}
	}

	companyID, err := uuid.Parse(p.CompanyID)
	if err != nil {
		return nil, &RPCError{Code: -32602, Message: "Invalid company_id format"}
	}

	result, err := m.companyService.GetCompanyHierarchy(companyID)
	if err != nil {
		switch err.(type) {
		case *exceptions.CompanyNotFoundError:
			return nil, &RPCError{Code: -32000, Message: err.Error()}
		default:
			return nil, &RPCError{Code: -32603, Message: err.Error()}
		}
	}

	return result, nil
}

func (m *RPCMethods) GetUsersByRole(params json.RawMessage) (interface{}, error) {
	var p struct {
		RoleID string `json:"role_id"`
	}

	if err := json.Unmarshal(params, &p); err != nil {
		return nil, &RPCError{Code: -32602, Message: "Invalid params: " + err.Error()}
	}

	if p.RoleID == "" {
		return nil, &RPCError{Code: -32602, Message: "role_id is required"}
	}

	roleID, err := uuid.Parse(p.RoleID)
	if err != nil {
		return nil, &RPCError{Code: -32602, Message: "Invalid role_id format"}
	}

	result, err := m.userService.GetUsersByRole(roleID)
	if err != nil {
		switch err.(type) {
		case *exceptions.RoleNotFoundError:
			return nil, &RPCError{Code: -32000, Message: err.Error()}
		default:
			return nil, &RPCError{Code: -32603, Message: err.Error()}
		}
	}

	return result, nil
}

func (m *RPCMethods) GetRoleUsers(params json.RawMessage) (interface{}, error) {
	var p struct {
		RoleID string `json:"role_id"`
	}

	if err := json.Unmarshal(params, &p); err != nil {
		return nil, &RPCError{Code: -32602, Message: "Invalid params: " + err.Error()}
	}

	if p.RoleID == "" {
		return nil, &RPCError{Code: -32602, Message: "role_id is required"}
	}

	roleID, err := uuid.Parse(p.RoleID)
	if err != nil {
		return nil, &RPCError{Code: -32602, Message: "Invalid role_id format"}
	}

	result, err := m.roleService.GetRoleUsers(roleID)
	if err != nil {
		switch err.(type) {
		case *exceptions.RoleNotFoundError:
			return nil, &RPCError{Code: -32000, Message: err.Error()}
		default:
			return nil, &RPCError{Code: -32603, Message: err.Error()}
		}
	}

	return result, nil
}

func (m *RPCMethods) CheckUserAccess(params json.RawMessage) (interface{}, error) {
	var req dto.CheckUserAccessRequest

	if err := json.Unmarshal(params, &req); err != nil {
		return nil, &RPCError{Code: -32602, Message: "Invalid params: " + err.Error()}
	}

	if req.UserID == "" {
		return nil, &RPCError{Code: -32602, Message: "user_id is required"}
	}
	if req.ResourceType == "" {
		return nil, &RPCError{Code: -32602, Message: "resource_type is required"}
	}
	if req.Action == "" {
		return nil, &RPCError{Code: -32602, Message: "action is required"}
	}

	result, err := m.permissionService.CheckUserAccess(req)
	if err != nil {
		switch err.(type) {
		case *exceptions.UserNotFoundError:
			return nil, &RPCError{Code: -32000, Message: err.Error()}
		default:
			return nil, &RPCError{Code: -32603, Message: err.Error()}
		}
	}

	return result, nil
}
