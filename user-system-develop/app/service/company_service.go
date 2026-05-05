package service

import (
	"user-system/app/dto"
	"user-system/app/exceptions"
	"user-system/app/models"
	"user-system/app/repositories"
	"time"

	"github.com/google/uuid"
)

type CompanyService interface {
	GetAllCompanies() ([]models.Company, error)
	GetCompanyByID(id uuid.UUID) (*models.Company, error)
	CreateCompany(company *models.Company) error
	UpdateCompany(company *models.Company) error
	DeleteCompany(id uuid.UUID) error
	GetCompanyHierarchy(companyID uuid.UUID) (*dto.CompanyHierarchyResponse, error)
}

type companyService struct {
	companyRepo repositories.CompanyRepository
}

func NewCompanyService(companyRepo repositories.CompanyRepository) CompanyService {
	return &companyService{
		companyRepo: companyRepo,
	}
}

func (s *companyService) GetAllCompanies() ([]models.Company, error) {
	return s.companyRepo.GetAll()
}

func (s *companyService) GetCompanyByID(id uuid.UUID) (*models.Company, error) {
	if id == uuid.Nil {
		return nil, &exceptions.CompanyNotFoundError{CompanyID: "invalid"}
	}

	company, err := s.companyRepo.GetByID(id)
	if err != nil {
		return nil, &exceptions.CompanyNotFoundError{CompanyID: id.String()}
	}
	return company, nil
}

func (s *companyService) CreateCompany(company *models.Company) error {

	if company.Name == "" {
		return &exceptions.ValidationError{Field: "name", Reason: "company name is required"}
	}

	if company.ID == uuid.Nil {
		company.ID = uuid.New()
	}

	if company.CreatedAt.IsZero() {
		company.CreatedAt = time.Now()
	}
	if company.UpdatedAt.IsZero() {
		company.UpdatedAt = time.Now()
	}

	if !company.IsActive {
		company.IsActive = true
	}
	return s.companyRepo.Create(company)
}

func (s *companyService) UpdateCompany(company *models.Company) error {
	if company.ID == uuid.Nil {
		return &exceptions.CompanyNotFoundError{CompanyID: "invalid"}
	}
	if company.Name == "" {
		return &exceptions.ValidationError{Field: "name", Reason: "company name is required"}
	}

	_, err := s.companyRepo.GetByID(company.ID)
	if err != nil {
		return &exceptions.CompanyNotFoundError{CompanyID: company.ID.String()}
	}

	return s.companyRepo.Update(company)
}

func (s *companyService) DeleteCompany(id uuid.UUID) error {
	if id == uuid.Nil {
		return &exceptions.CompanyNotFoundError{CompanyID: "invalid"}
	}

	_, err := s.companyRepo.GetByID(id)
	if err != nil {
		return &exceptions.CompanyNotFoundError{CompanyID: id.String()}
	}

	return s.companyRepo.Delete(id)
}

func (s *companyService) GetCompanyHierarchy(companyID uuid.UUID) (*dto.CompanyHierarchyResponse, error) {
	company, err := s.companyRepo.FindByID(companyID)
	if err != nil {
		return nil, &exceptions.CompanyNotFoundError{CompanyID: companyID.String()}
	}

	users, err := s.companyRepo.GetCompanyWithUsers(companyID)
	if err != nil {
		return nil, err
	}

	departmentMap := make(map[string][]dto.UserBriefInfo)
	for _, user := range users {
		userBrief := dto.UserBriefInfo{
			ID:        user.ID,
			Email:     user.Email,
			FirstName: user.FirstName,
			LastName:  user.LastName,
			Position:  user.Position,
			IsActive:  user.IsActive,
		}

		dept := user.Department
		if dept == "" {
			dept = "Без отдела"
		}
		departmentMap[dept] = append(departmentMap[dept], userBrief)
	}

	var departments []dto.DepartmentInfo
	for name, users := range departmentMap {
		departments = append(departments, dto.DepartmentInfo{
			Name:  name,
			Users: users,
			Count: len(users),
		})
	}

	return &dto.CompanyHierarchyResponse{
		CompanyID:   company.ID,
		CompanyName: company.Name,
		Departments: departments,
		TotalUsers:  len(users),
	}, nil
}
