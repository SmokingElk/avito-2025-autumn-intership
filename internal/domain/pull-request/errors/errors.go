package errors

import "errors"

var (
	ErrTeamOrUserNotFound = errors.New("team or user not found")
	ErrAlreadyExists      = errors.New("pr already exists")
	ErrNotFound           = errors.New("pr not found")
	ErrAlreadyMerged      = errors.New("pr already merged and can not be reassigned")
)
