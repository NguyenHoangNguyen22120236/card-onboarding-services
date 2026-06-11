package entity

type Status string

const (
	StatusInProgress Status = "IN_PROGRESS"
	StatusSucceeded  Status = "SUCCEEDED"
	StatusFailed     Status = "FAILED"
)
