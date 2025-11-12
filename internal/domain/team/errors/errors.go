package errors

import "errors"

var (
	ErrTeamNotFound      = errors.New("team not found")
	ErrTeamExists        = errors.New("team already exists")
	ErrMemberOfOtherTeam = errors.New("user is already member of other team")
)
