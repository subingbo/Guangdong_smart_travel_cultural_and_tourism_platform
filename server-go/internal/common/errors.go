package common

import "net/http"

// ApiError is a business error that carries the HTTP status to return.
type ApiError struct {
	Status  int
	Message string
}

func (e *ApiError) Error() string { return e.Message }

func BadRequest(msg string) *ApiError      { return &ApiError{http.StatusBadRequest, msg} }
func Unauthorized(msg string) *ApiError    { return &ApiError{http.StatusUnauthorized, msg} }
func Forbidden(msg string) *ApiError       { return &ApiError{http.StatusForbidden, msg} }
func NotFound(msg string) *ApiError        { return &ApiError{http.StatusNotFound, msg} }
func Conflict(msg string) *ApiError        { return &ApiError{http.StatusConflict, msg} }
func TooManyRequests(msg string) *ApiError { return &ApiError{http.StatusTooManyRequests, msg} }
func BadGateway(msg string) *ApiError      { return &ApiError{http.StatusBadGateway, msg} }
func ServiceUnavailable(msg string) *ApiError {
	return &ApiError{http.StatusServiceUnavailable, msg}
}
func Internal(msg string) *ApiError { return &ApiError{http.StatusInternalServerError, msg} }
