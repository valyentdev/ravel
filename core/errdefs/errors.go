package errdefs

import (
	"encoding/json"
	"errors"
	"net/http"

	"github.com/danielgtaylor/huma/v2"
	"google.golang.org/grpc/codes"
)

type Code string

const (
	codeInvalidArgument    Code = "INVALID_ARGUMENT"
	codeNotFound           Code = "NOT_FOUND"
	codeAlreadyExists      Code = "ALREADY_EXISTS"
	codeFailedPrecondition Code = "FAILED_PRECONDITION"
	codeInternal           Code = "INTERNAL"
	codeNotImplemented     Code = "NOT_IMPLEMENTED"
	codeDeadlineExceeded   Code = "DEADLINE_EXCEEDED"
	codeResourcesExhausted Code = "RESOURCES_EXHAUSTED"
	codeInternalError      Code = "INTERNAL"
	codeUnknown            Code = "UNKNOWN"
)

type ErrorDetail = huma.ErrorDetail

type RavelError struct {
	grpcCode  codes.Code
	RavelCode Code           `json:"code"`
	Status    int            `json:"status"`
	Title     string         `json:"title"`
	Detail    string         `json:"detail"`
	Errors    []*ErrorDetail `json:"errors,omitempty"`
}

var _ interface {
	error
	huma.StatusError
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

func new(code Code, msg string, details ...*ErrorDetail) *RavelError {
	return &RavelError{
		RavelCode: code,
		Status:    getHTTPStatus(code),
		Title:     string(code),
		Detail:    msg,
		Errors:    details,
	}
}

func FromHTTPResponse(resp *http.Response) *RavelError {
	model := RavelError{}

	err := json.NewDecoder(resp.Body).Decode(&model)
	if err != nil {
		return &RavelError{
			grpcCode:  codes.Unknown,
			RavelCode: codeUnknown,
			Status:    resp.StatusCode,
			Detail:    http.StatusText(resp.StatusCode),
			Title:     http.StatusText(resp.StatusCode),
		}
	}

	model.Status = resp.StatusCode

	return &model
}

func is(err error, code Code) bool {
	var rErr *RavelError
	if !errors.As(err, &rErr) {
		return false
	}

	return rErr.Code() == code
}

func IsNotFound(err error) bool {
	return is(err, codeNotFound)
}

func IsAlreadyExists(err error) bool {
	return is(err, codeAlreadyExists)
}

func IsFailedPrecondition(err error) bool {
	return is(err, codeFailedPrecondition)
}

func IsResourcesExhausted(err error) bool {
	return is(err, codeResourcesExhausted)
}

func IsInvalidArgument(err error) bool {
	return is(err, codeInvalidArgument)
}

func IsInternal(err error) bool {
	return is(err, codeInternal)
}

func IsUnknown(err error) bool {
	return is(err, codeUnknown)
}

func IsNotImplemented(err error) bool {
	return is(err, codeNotImplemented)
}

func getHTTPStatus(code Code) int {
	switch code {
	case codeInvalidArgument:
		return http.StatusBadRequest
	case codeFailedPrecondition:
		return http.StatusBadRequest
	case codeNotFound:
		return http.StatusNotFound
	case codeAlreadyExists:
		return http.StatusConflict
	case codeDeadlineExceeded:
		return http.StatusRequestTimeout
	case codeNotImplemented:
		return http.StatusNotImplemented
	case codeResourcesExhausted:
		return http.StatusTooManyRequests
	default:
		return http.StatusInternalServerError
	}
}

func getRavelCodeFromHTTPStatus(status int) Code {
	switch status {
	case http.StatusBadRequest:
		return codeInvalidArgument
	case http.StatusTooManyRequests:
		return codeResourcesExhausted
	case http.StatusInternalServerError:
		return codeInternal
	default:
		return codeUnknown
	}
}

func OverrideHumaErrorBuilder() {
	huma.NewError = func(status int, msg string, errs ...error) huma.StatusError {
		details := make([]*huma.ErrorDetail, 0, len(errs))
		for _, err := range errs {
			if err == nil {
				continue
			}
			var detail *huma.ErrorDetail
			if errors.As(err, &detail) {
				details = append(details, detail)
			} else {
				details = append(details, &huma.ErrorDetail{
					Message: err.Error(),
				})
			}
		}

		if status == 500 {
			return &RavelError{
				Detail:    "An uknown error occurred. Please try again later or check logs.",
				RavelCode: codeInternal,
				Title:     http.StatusText(status),
				Status:    status,
			}
		}

		return &RavelError{
			Detail:    msg,
			RavelCode: getRavelCodeFromHTTPStatus(status),
			Title:     http.StatusText(status),
			Status:    status,
			Errors:    details,
		}
	}
}

func NewUnknown(msg string) error {
	return new(codeUnknown, msg)
}

func NewInvalidArgument(msg string, details ...*ErrorDetail) error {
	return new(codeInvalidArgument, msg, details...)
}

func NewNotFound(msg string) error {
	return new(codeNotFound, msg)
}

func NewAlreadyExists(msg string) error {
	return new(codeAlreadyExists, msg)
}

func NewFailedPrecondition(msg string) error {
	return new(codeFailedPrecondition, msg)
}

func NewNotImplemented(msg string) error {
	return new(codeNotImplemented, msg)
}

func NewDeadlineExceeded(msg string) error {
	return new(codeDeadlineExceeded, msg)
}

func NewResourcesExhausted(msg string) error {
	return new(codeResourcesExhausted, msg)
}
