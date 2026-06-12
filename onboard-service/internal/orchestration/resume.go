package orchestration

import "github.com/NguyenHoangNguyen22120236/card-onboarding-services/onboard-service/internal/entity"

func isCompleted(status entity.RequestStatus) bool {
	return status.OverallStatus == entity.StatusSucceeded
}

func shouldRegisterCustomer(status entity.RequestStatus, newStatus bool) bool {
	return newStatus && status.CustomerRegistrationStatus == entity.StatusInProgress
}

func shouldFetchInterestDetails(status entity.RequestStatus) bool {
	return status.CustomerRegistrationStatus == entity.StatusSucceeded &&
		(status.InterestDetailsStatus == "" ||
			status.InterestDetailsStatus == entity.StatusInProgress ||
			status.InterestDetailsStatus == entity.StatusFailed)
}

func shouldOnboardAccount(status entity.RequestStatus) bool {
	return status.InterestDetailsStatus == entity.StatusSucceeded &&
		(status.AccountOnboardingStatus == "" ||
			status.AccountOnboardingStatus == entity.StatusInProgress ||
			status.AccountOnboardingStatus == entity.StatusFailed)
}
