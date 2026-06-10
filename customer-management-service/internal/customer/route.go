package customer

import (
	"github.com/gin-gonic/gin"

	api "github.com/NguyenHoangNguyen22120236/card-onboarding-services/customer-management-service/pkg/customer"
)

func NewRouter(service CustomerService) *gin.Engine {
	router := gin.Default()
	api.RegisterHandlers(router, NewHandler(service))
	return router
}
