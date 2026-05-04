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

type CompanyController struct {
	service service.CompanyService
}

func NewCompanyController(service service.CompanyService) *CompanyController {
	return &CompanyController{service: service}
}

func (c *CompanyController) GetAllCompanies(ctx *gin.Context) {
	companies, err := c.service.GetAllCompanies()
	if err != nil {
		ctx.JSON(http.StatusInternalServerError, dto.ErrorResponse{Error: "failed to get companies"})
		return
	}

	responseDTOs := mapper.CompaniesToDTOs(companies)

	ctx.JSON(http.StatusOK, dto.SuccessResponse{
		Message: "companies retrieved successfully",
		Data:    responseDTOs,
	})
}

func (c *CompanyController) GetCompanyByID(ctx *gin.Context) {
	id, err := uuid.Parse(ctx.Param("id"))
	if err != nil {
		ctx.JSON(http.StatusBadRequest, dto.ErrorResponse{Error: "invalid company ID"})
		return
	}

	company, err := c.service.GetCompanyByID(id)
	if err != nil {
		if _, ok := err.(*exceptions.CompanyNotFoundError); ok {
			ctx.JSON(http.StatusNotFound, dto.ErrorResponse{Error: err.Error()})
		} else {
			ctx.JSON(http.StatusInternalServerError, dto.ErrorResponse{Error: "internal server error"})
		}
		return
	}

	responseDTO := mapper.CompanyToDTO(company)

	ctx.JSON(http.StatusOK, dto.SuccessResponse{
		Message: "company retrieved successfully",
		Data:    responseDTO,
	})
}

func (c *CompanyController) UpdateCompany(ctx *gin.Context) {
	id, err := uuid.Parse(ctx.Param("id"))
	if err != nil {
		ctx.JSON(http.StatusBadRequest, dto.ErrorResponse{Error: "invalid company ID"})
		return
	}

	var companyDTO dto.UpdateCompanyRequest
	if err := ctx.ShouldBindJSON(&companyDTO); err != nil {
		ctx.JSON(http.StatusBadRequest, dto.ErrorResponse{
			Error:   "invalid request data",
			Message: err.Error(),
		})
		return
	}

	existingCompany, err := c.service.GetCompanyByID(id)
	if err != nil {
		if _, ok := err.(*exceptions.CompanyNotFoundError); ok {
			ctx.JSON(http.StatusNotFound, dto.ErrorResponse{Error: err.Error()})
		} else {
			ctx.JSON(http.StatusInternalServerError, dto.ErrorResponse{Error: "internal server error"})
		}
		return
	}

	mapper.UpdateCompanyRequestToModel(existingCompany, companyDTO)

	if err := c.service.UpdateCompany(existingCompany); err != nil {
		switch e := err.(type) {
		case *exceptions.CompanyNotFoundError:
			ctx.JSON(http.StatusNotFound, dto.ErrorResponse{Error: e.Error()})
		case *exceptions.ValidationError:
			ctx.JSON(http.StatusBadRequest, dto.ErrorResponse{Error: e.Error()})
		default:
			ctx.JSON(http.StatusInternalServerError, dto.ErrorResponse{Error: "failed to update company"})
		}
		return
	}

	responseDTO := mapper.CompanyToDTO(existingCompany)

	ctx.JSON(http.StatusOK, dto.SuccessResponse{
		Message: "company updated successfully",
		Data:    responseDTO,
	})
}

func (c *CompanyController) DeleteCompany(ctx *gin.Context) {
	id, err := uuid.Parse(ctx.Param("id"))
	if err != nil {
		ctx.JSON(http.StatusBadRequest, dto.ErrorResponse{Error: "invalid company ID"})
		return
	}

	if err := c.service.DeleteCompany(id); err != nil {
		if _, ok := err.(*exceptions.CompanyNotFoundError); ok {
			ctx.JSON(http.StatusNotFound, dto.ErrorResponse{Error: err.Error()})
		} else {
			ctx.JSON(http.StatusInternalServerError, dto.ErrorResponse{Error: "failed to delete company"})
		}
		return
	}

	ctx.JSON(http.StatusNoContent, nil)
}

func (c *CompanyController) CreateCompany(ctx *gin.Context) {
	var req dto.CreateCompanyPublicRequest
	if err := ctx.ShouldBindJSON(&req); err != nil {
		ctx.JSON(http.StatusBadRequest, dto.ErrorResponse{
			Error:   "invalid request data",
			Message: err.Error(),
		})
		return
	}

	if req.Domain == "" {
		req.Domain = "company-" + uuid.New().String()
	}

	company := mapper.CreatePublicRequestToModel(req)

	if err := c.service.CreateCompany(&company); err != nil {
		switch e := err.(type) {
		case *exceptions.ValidationError:
			ctx.JSON(http.StatusBadRequest, dto.ErrorResponse{Error: e.Error()})
		default:
			ctx.JSON(http.StatusInternalServerError, dto.ErrorResponse{Error: "failed to create company"})
		}
		return
	}

	responseDTO := mapper.CompanyToDTO(&company)

	ctx.JSON(http.StatusCreated, dto.SuccessResponse{
		Message: "company created successfully",
		Data:    responseDTO,
	})
}
