package providers

import "errors"

var (
	ErrLoginFailed          = errors.New("login failed")
	ErrLanguageNotSupported = errors.New("language not supported")
	ErrSubmissionNotFound   = errors.New("submit id not found")
	ErrStatusNotFound       = errors.New("status not found")
)

type account struct {
	username string
	password string
}
