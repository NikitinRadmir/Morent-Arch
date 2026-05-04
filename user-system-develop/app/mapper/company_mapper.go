package mapper

import (
	"user-system/app/dto"
	"user-system/app/models"
)

func CompanyToDTO(company *models.Company) dto.CompanyResponse {
	return dto.CompanyResponse{
		ID:          company.ID,
		Name:        company.Name,
		Description: company.Description,
		Domain:      company.Domain,
		CreatedAt:   company.CreatedAt,
		UpdatedAt:   company.UpdatedAt,
		IsActive:    company.IsActive,
	}
}

func CompaniesToDTOs(companies []models.Company) []dto.CompanyResponse {
	dtos := make([]dto.CompanyResponse, 0, len(companies))
	for _, company := range companies {
		dtos = append(dtos, CompanyToDTO(&company))
	}
	return dtos
}

func CreateCompanyRequestToModel(req dto.CreateCompanyRequest) models.Company {
	return models.Company{
		Name:        req.Name,
		Description: req.Description,
		Domain:      req.Domain,
	}
}

func CreatePublicRequestToModel(req dto.CreateCompanyPublicRequest) models.Company {
	return models.Company{
		Name:        req.Name,
		Description: req.Description,
		Domain:      req.Domain,
	}
}

func UpdateCompanyRequestToModel(existingCompany *models.Company, req dto.UpdateCompanyRequest) {
	if req.Name != "" {
		existingCompany.Name = req.Name
	}
	if req.Description != "" {
		existingCompany.Description = req.Description
	}
	if req.Domain != "" {
		existingCompany.Domain = req.Domain
	}
	if req.IsActive != nil {
		existingCompany.IsActive = *req.IsActive
	}
}
