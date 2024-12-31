package response

import (
	"fmt"
	"github.com/go-playground/validator/v10"
	"strings"
)

type Response struct {
	Status string `json:"status"`
	Error  string `json:"error"`
}

const (
	StatusOK    = "OK"
	StatusError = "Error"
)

func Error(msg string) Response {
	return Response{
		Status: StatusError,
		Error:  msg,
	}
}

func OK() Response {
	return Response{
		Status: StatusOK,
	}
}
func ValidateError(errs validator.ValidationErrors) Response {
	var ErrMsgs []string

	for _, err := range errs {
		switch err.ActualTag() {
		case "required":
			ErrMsgs = append(ErrMsgs, fmt.Sprintf("field %s is required", err.Field()))
		case "url":
			ErrMsgs = append(ErrMsgs, fmt.Sprintf("field %s is not valid URL"), err.Field())
		default:
			ErrMsgs = append(ErrMsgs, fmt.Sprintf("field %s is not valid", err.Field()))
		}
	}
	return Response{
		Status: StatusError,
		Error:  strings.Join(ErrMsgs, ","),
	}

}
