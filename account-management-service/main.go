package main

import account "github.com/NguyenHoangNguyen22120236/card-onboarding-services/account-management-service/internal/account"

func main() {
	router := account.NewRouter(account.NewService())
	if err := router.Run(":8082"); err != nil {
		panic(err)
	}
}
