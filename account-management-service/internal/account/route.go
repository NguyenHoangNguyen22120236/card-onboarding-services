package account

import (
	"github.com/gin-gonic/gin"

	api "github.com/NguyenHoangNguyen22120236/card-onboarding-services/account-management-service/pkg/account"
)

func NewRouter(service AccountService) *gin.Engine {
	router := gin.Default()
	api.RegisterHandlers(router, NewHandler(service))
	return router
}
