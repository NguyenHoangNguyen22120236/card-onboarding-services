package main

import customer "github.com/NguyenHoangNguyen22120236/card-onboarding-services/customer-management-service/internal/customer"

func main() {
	router := customer.NewRouter(customer.NewService())
	if err := router.Run(":8081"); err != nil {
		panic(err)
	}
}
