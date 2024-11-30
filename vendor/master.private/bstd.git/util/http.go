package util

import "fmt"

type HttpError interface {
	error
	StatusCode() int
	Message() string
}

type httpError struct {
	code    int
	message string
}

func NewHttpError(statusCode int, message string) HttpError {
	return &httpError{
		code:    statusCode,
		message: message,
	}
}

func (h *httpError) Error() string {
	return fmt.Sprintf(
		"http error:\n\tstatuscode: %d\n\tmessage: %s",
		h.code, h.message,
	)
}

func (h *httpError) StatusCode() int {
	return h.code
}

func (h *httpError) Message() string {
	return h.message
}
