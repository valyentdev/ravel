package ravelerrors

import (
	"errors"
	"net/http"

	"github.com/danielgtaylor/huma/v2"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

type Code string

const (
	codeUnknown            Code = "UNKNOWN"
	codeInvalidArgument    Code = "INVALID_ARGUMENT"
	codeNotFound           Code = "NOT_FOUND"
	codeAlreadyExists      Code = "ALREADY_EXISTS"
	codeFailedPrecondition Code = "FAILED_PRECONDITION"
	codeUnavailable        Code = "UNAVAILABLE"
	codeInternal           Code = "INTERNAL"
	codeResourcesExhausted Code = "RESOURCES_EXHAUSTED"
	codeNotImplemented     Code = "NOT_IMPLEMENTED"
	codeDeadlineExceeded   Code = "DEADLINE_EXCEEDED"
)

func getHTTPStatus(code Code) int {
	switch code {
	case codeInvalidArgument:
		return http.StatusBadRequest
	case codeNotFound:
		return http.StatusNotFound
	case codeAlreadyExists:
		return http.StatusConflict
	case codeFailedPrecondition:
		return http.StatusConflict
	case codeDeadlineExceeded:
		return http.StatusRequestTimeout
	case codeUnavailable:
		return http.StatusServiceUnavailable
	case codeNotImplemented:
		return http.StatusNotImplemented
	case codeResourcesExhausted:
		return http.StatusInsufficientStorage
	default:
		return http.StatusInternalServerError
	}
}

func getGRPCCode(code Code) codes.Code {
	switch code {
	case codeInvalidArgument:
		return codes.InvalidArgument
	case codeNotFound:
		return codes.NotFound
	case codeAlreadyExists:
		return codes.AlreadyExists
	case codeFailedPrecondition:
		return codes.FailedPrecondition
	case codeUnavailable:
		return codes.Unavailable
	case codeNotImplemented:
		return codes.Unimplemented
	case codeDeadlineExceeded:
		return codes.DeadlineExceeded
	case codeResourcesExhausted:
		return codes.ResourceExhausted
	default:
		return codes.Unknown
	}
}

var _ error = (*ravelError)(nil)
var _ huma.StatusError = (*ravelError)(nil)
var _ interface{ GRPCStatus() *status.Status } = (*ravelError)(nil)

type ravelError struct {
	code Code
	msg  string
}

func (r *ravelError) Code() Code {
	return r.code
}

func (r *ravelError) Error() string {
	return r.msg
}

func (r *ravelError) GetStatus() int {
	return getHTTPStatus(r.code)
}

func (r *ravelError) GRPCStatus() *status.Status {
	return status.New(getGRPCCode(r.code), r.msg)
}

type RavelError interface {
	Code() Code                 // Ravel specific error code
	GRPCStatus() *status.Status // Associated grpc status
	GetStatus() int             // Associated http status
	Error() string              // Error message for humans
}

func new(code Code, msg string) RavelError {
	return &ravelError{
		code: code,
		msg:  msg,
	}
}

func New(err RavelError, msg string) RavelError {
	if err == nil {
		return nil
	}

	return new(err.Code(), msg)
}

func NewUnknown(msg string) RavelError {
	return new(codeUnknown, msg)
}

func NewInvalidArgument(msg string) RavelError {
	return new(codeInvalidArgument, msg)
}

func NewNotFound(msg string) RavelError {
	return new(codeNotFound, msg)
}

func NewAlreadyExists(msg string) RavelError {
	return new(codeAlreadyExists, msg)
}

func NewFailedPrecondition(msg string) RavelError {
	return new(codeFailedPrecondition, msg)
}

func NewUnavailable(msg string) RavelError {
	return new(codeUnavailable, msg)
}

func NewNotImplemented(msg string) RavelError {
	return new(codeNotImplemented, msg)
}

func NewDeadlineExceeded(msg string) RavelError {
	return new(codeDeadlineExceeded, msg)
}

func NewResourcesExhausted(msg string) RavelError {
	return new(codeResourcesExhausted, msg)
}

func FromGRPCErr(err error) RavelError {
	if err == nil {
		return nil
	}

	st, ok := status.FromError(err)
	if !ok {
		return NewUnknown(err.Error())
	}

	return new(Code(st.Code().String()), st.Message())
}

func IsNotFound(err error) bool {
	var rErr RavelError
	if !errors.As(err, &rErr) {
		return false
	}

	return rErr.Code() == codeNotFound
}

func IsAlreadyExists(err error) bool {
	var rErr RavelError
	if !errors.As(err, &rErr) {
		return false
	}

	return rErr.Code() == codeAlreadyExists
}

func IsFailedPrecondition(err error) bool {
	var rErr RavelError
	if !errors.As(err, &rErr) {
		return false
	}

	return rErr.Code() == codeFailedPrecondition
}

func IsResourcesExhausted(err error) bool {
	var rErr RavelError
	if !errors.As(err, &rErr) {
		return false
	}

	return rErr.Code() == codeResourcesExhausted
}

func IsInvalidArgument(err error) bool {
	var rErr RavelError
	if !errors.As(err, &rErr) {
		return false
	}

	return rErr.Code() == codeInvalidArgument
}
