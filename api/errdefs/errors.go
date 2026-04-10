package errdefs

import (
	"encoding/json"
	"errors"
	"net/http"
)

type Code string

const (
	CodeInvalidArgument    Code = "INVALID_ARGUMENT"
	CodeNotFound           Code = "NOT_FOUND"
	CodeAlreadyExists      Code = "ALREADY_EXISTS"
	CodeFailedPrecondition Code = "FAILED_PRECONDITION"
	CodeInternal           Code = "INTERNAL"
	CodeNotImplemented     Code = "NOT_IMPLEMENTED"
	CodeDeadlineExceeded   Code = "DEADLINE_EXCEEDED"
	CodeResourcesExhausted Code = "RESOURCES_EXHAUSTED"
	CodeInternalError      Code = "INTERNAL"
	CodeUnknown            Code = "UNKNOWN"
)

type ErrorDetail struct {
	Message  string `json:"message,omitempty" doc:"Error message text"`
	Location string `json:"location,omitempty" doc:"Where the error occurred, e.g. 'body.items[3].tags' or 'path.thing-id'"`
	Value    any    `json:"value,omitempty" doc:"The value at the given location"`
}

type RavelError struct {
	RavelCode Code           `json:"Code"`
	Status    int            `json:"status"`
	Title     string         `json:"title"`
	Detail    string         `json:"detail"`
	Errors    []*ErrorDetail `json:"errors,omitempty"`
}

var _ interface {
	GetStatus() int
	Error() string
	Code() Code
} = (*RavelError)(nil)

func (r *RavelError) Code() Code {
	return r.RavelCode
}

func (r *RavelError) Error() string {
	return r.Detail
}

func (r *RavelError) GetStatus() int {
	return r.Status
}

func new(Code Code, msg string, details ...*ErrorDetail) *RavelError {
	return &RavelError{
		RavelCode: Code,
		Status:    getHTTPStatus(Code),
		Title:     string(Code),
		Detail:    msg,
		Errors:    details,
	}
}

func FromHTTPResponse(resp *http.Response) *RavelError {
	model := RavelError{}

	err := json.NewDecoder(resp.Body).Decode(&model)
	if err != nil {
		return &RavelError{
			RavelCode: CodeUnknown,
			Status:    resp.StatusCode,
			Detail:    http.StatusText(resp.StatusCode),
			Title:     http.StatusText(resp.StatusCode),
		}
	}

	model.Status = resp.StatusCode

	return &model
}

func is(err error, Code Code) bool {
	var rErr *RavelError
	if !errors.As(err, &rErr) {
		return false
	}

	return rErr.Code() == Code
}

func IsNotFound(err error) bool {
	return is(err, CodeNotFound)
}

func IsAlreadyExists(err error) bool {
	return is(err, CodeAlreadyExists)
}

func IsFailedPrecondition(err error) bool {
	return is(err, CodeFailedPrecondition)
}

func IsResourcesExhausted(err error) bool {
	return is(err, CodeResourcesExhausted)
}

func IsInvalidArgument(err error) bool {
	return is(err, CodeInvalidArgument)
}

func IsInternal(err error) bool {
	return is(err, CodeInternal)
}

func IsUnknown(err error) bool {
	return is(err, CodeUnknown)
}

func IsNotImplemented(err error) bool {
	return is(err, CodeNotImplemented)
}

func getHTTPStatus(Code Code) int {
	switch Code {
	case CodeInvalidArgument:
		return http.StatusBadRequest
	case CodeFailedPrecondition:
		return http.StatusBadRequest
	case CodeNotFound:
		return http.StatusNotFound
	case CodeAlreadyExists:
		return http.StatusConflict
	case CodeDeadlineExceeded:
		return http.StatusRequestTimeout
	case CodeNotImplemented:
		return http.StatusNotImplemented
	case CodeResourcesExhausted:
		return http.StatusTooManyRequests
	default:
		return http.StatusInternalServerError
	}
}

func NewUnknown(msg string) error {
	return new(CodeUnknown, msg)
}

func NewInvalidArgument(msg string, details ...*ErrorDetail) error {
	return new(CodeInvalidArgument, msg, details...)
}

func NewNotFound(msg string) error {
	return new(CodeNotFound, msg)
}

func NewAlreadyExists(msg string) error {
	return new(CodeAlreadyExists, msg)
}

func NewFailedPrecondition(msg string) error {
	return new(CodeFailedPrecondition, msg)
}

func NewNotImplemented(msg string) error {
	return new(CodeNotImplemented, msg)
}

func NewDeadlineExceeded(msg string) error {
	return new(CodeDeadlineExceeded, msg)
}

func NewResourcesExhausted(msg string) error {
	return new(CodeResourcesExhausted, msg)
}
