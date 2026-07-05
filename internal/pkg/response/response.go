package response

import (
	"errors"
	"net/http"

	"chatgoo/internal/pkg/errcode"
)

// Response is the unified API response structure.
type Response struct {
	Code    errcode.ErrCode `json:"code"`
	Message string          `json:"message"`
	Data    interface{}     `json:"data,omitempty"`
}

// OK creates a success response.
func OK(data interface{}) *Response {
	return &Response{Code: errcode.Success, Message: "success", Data: data}
}

// APIError implements error and carries an HTTP status code.
type APIError struct {
	Code       errcode.ErrCode
	Msg        string
	HTTPStatus int
}

func (e *APIError) Error() string { return e.Msg }

// BadRequest returns a 400 error.
func BadRequest(msg string) error {
	return &APIError{Code: errcode.InvalidParams, Msg: msg, HTTPStatus: http.StatusBadRequest}
}

// Unauthorized returns a 401 error.
func Unauthorized(msg string) error {
	return &APIError{Code: errcode.Unauthorized, Msg: msg, HTTPStatus: http.StatusUnauthorized}
}

// NotFound returns a 404 error.
func NotFound(msg string) error {
	return &APIError{Code: errcode.NotFound, Msg: msg, HTTPStatus: http.StatusNotFound}
}

// Conflict returns a 409 error.
func Conflict(msg string) error {
	return &APIError{Code: errcode.Conflict, Msg: msg, HTTPStatus: http.StatusConflict}
}

// FromError converts a service-layer error into an APIError.
func FromError(err error) error {
	if err == nil {
		return nil
	}

	var apiErr *APIError
	if errors.As(err, &apiErr) {
		return apiErr
	}

	switch err.Error() {
	case "username already exists":
		return &APIError{Code: errcode.UsernameExists, Msg: err.Error(), HTTPStatus: http.StatusConflict}
	case "invalid username or password":
		return &APIError{Code: errcode.InvalidCredential, Msg: err.Error(), HTTPStatus: http.StatusUnauthorized}
	case "session not found":
		return &APIError{Code: errcode.SessionNotFound, Msg: err.Error(), HTTPStatus: http.StatusNotFound}
	case "user is not a participant of this session":
		return &APIError{Code: errcode.NotParticipant, Msg: err.Error(), HTTPStatus: http.StatusForbidden}
	default:
		return &APIError{Code: errcode.InternalError, Msg: "internal server error", HTTPStatus: http.StatusInternalServerError}
	}
}
