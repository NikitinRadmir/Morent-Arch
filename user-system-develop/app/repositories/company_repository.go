package repositories

import (
	"user-system/app/models"
	"errors"

	"github.com/google/uuid"
	"gorm.io/gorm"
)

type CompanyRepository interface {
	GetAll() ([]models.Company, error)
	GetByID(id uuid.UUID) (*models.Company, error)
	GetByName(name string) (*models.Company, error)
	Create(company *models.Company) error
	Update(company *models.Company) error
	Delete(id uuid.UUID) error
	FindByID(id uuid.UUID) (*models.Company, error)
	GetCompanyWithUsers(id uuid.UUID) ([]models.User, error)
}

type companyRepository struct {
	db *gorm.DB
}

func NewCompanyRepository(db *gorm.DB) CompanyRepository {
	return &companyRepository{db: db}
}

func (r *companyRepository) GetAll() ([]models.Company, error) {
	var companies []models.Company
	err := r.db.Find(&companies).Error
	return companies, err
}

func (r *companyRepository) GetByName(name string) (*models.Company, error) {
	var company models.Company
	err := r.db.Where("name = ?", name).First(&company).Error
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, nil
		}
		return nil, err
	}
	return &company, nil
}

func (r *companyRepository) GetByID(id uuid.UUID) (*models.Company, error) {
	var company models.Company
	err := r.db.First(&company, "id = ?", id).Error
	if err != nil {
		return nil, err
	}
	return &company, nil
}

func (r *companyRepository) Create(company *models.Company) error {
	return r.db.Create(company).Error
}

func (r *companyRepository) Update(company *models.Company) error {
	return r.db.Save(company).Error
}

func (r *companyRepository) Delete(id uuid.UUID) error {
	return r.db.Delete(&models.Company{}, "id = ?", id).Error
}

func (r *companyRepository) FindByID(id uuid.UUID) (*models.Company, error) {
	var company models.Company
	err := r.db.Where("id = ?", id).First(&company).Error
	if err != nil {
		return nil, err
	}
	return &company, nil
}

func (r *companyRepository) GetCompanyWithUsers(id uuid.UUID) ([]models.User, error) {
	var users []models.User
	err := r.db.Where("company_id = ?", id).Find(&users).Error
	return users, err
}
