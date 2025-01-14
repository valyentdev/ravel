package humautil

import (
	"errors"
	"net/http"

	"github.com/danielgtaylor/huma/v2"
	"github.com/valyentdev/ravel/api/errdefs"
)

func OverrideHumaErrorBuilder() {
	huma.NewError = func(status int, msg string, errs ...error) huma.StatusError {
		details := make([]*errdefs.ErrorDetail, 0, len(errs))
		for _, err := range errs {
			if err == nil {
				continue
			}
			var detail *huma.ErrorDetail
			if errors.As(err, &detail) {
				details = append(details, &errdefs.ErrorDetail{
					Message:  detail.Message,
					Location: detail.Location,
					Value:    detail.Value,
				})
			} else {
				details = append(details, &errdefs.ErrorDetail{
					Message: err.Error(),
				})
			}
		}

		if status == 500 {
			return &errdefs.RavelError{
				RavelCode: errdefs.CodeInternal,
				Status:    status,
				Title:     http.StatusText(status),
				Detail:    "An uknown error occurred. Please try again later or check logs.",
			}
		}

		return &errdefs.RavelError{
			Detail:    msg,
			RavelCode: getRavelCodeFromHTTPStatus(status),
			Title:     http.StatusText(status),
			Status:    status,
			Errors:    details,
		}
	}
}

func getRavelCodeFromHTTPStatus(status int) errdefs.Code {
	switch status {
	case http.StatusBadRequest:
		return errdefs.CodeInvalidArgument
	case http.StatusTooManyRequests:
		return errdefs.CodeResourcesExhausted
	case http.StatusInternalServerError:
		return errdefs.CodeInternal
	default:
		return errdefs.CodeUnknown
	}
}
