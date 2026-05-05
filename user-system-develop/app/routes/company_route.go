package routes

import (
	"user-system/app/controller"

	"github.com/gin-gonic/gin"
)

func RegisterCompanyRoutes(router *gin.RouterGroup, companyController *controller.CompanyController) {
	company := router.Group("/companies")
	{
		company.GET("", companyController.GetAllCompanies)
		company.GET("/:id", companyController.GetCompanyByID)
		company.PUT("/:id", companyController.UpdateCompany)
		company.DELETE("/:id", companyController.DeleteCompany)
	}
}
