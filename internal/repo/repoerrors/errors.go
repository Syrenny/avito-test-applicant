package repoerrors

import "errors"

var (
	ErrNotFound      = errors.New("not found")
	ErrAlreadyExists = errors.New("already exists")

	ErrNotEnoughBalance    = errors.New("not enough balance")
	ErrUsernameTakenInTeam = errors.New("username already taken in team")
)
