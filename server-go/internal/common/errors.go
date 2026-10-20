package common

import "net/http"

/** ApiError mirrors Java's ApiException: a business error with an HTTP status. */
type ApiError struct {
	Status  int
	Message string
}

func (e *ApiError) Error() string { return e.Message }

func BadRequest(msg string) *ApiError   { return &ApiError{http.StatusBadRequest, msg} }
func Unauthorized(msg string) *ApiError { return &ApiError{http.StatusUnauthorized, msg} }
func NotFound(msg string) *ApiError     { return &ApiError{http.StatusNotFound, msg} }
func Conflict(msg string) *ApiError     { return &ApiError{http.StatusConflict, msg} }
func BadGateway(msg string) *ApiError   { return &ApiError{http.StatusBadGateway, msg} }
func Internal(msg string) *ApiError     { return &ApiError{http.StatusInternalServerError, msg} }
