package card

import (
	"github.com/gin-gonic/gin"

	api "github.com/NguyenHoangNguyen22120236/card-onboarding-services/onboard-service/pkg/onboard"
)

func NewRouter(service CardService) *gin.Engine {
	router := gin.Default()
	api.RegisterHandlers(router, NewHandler(service))
	return router
}
