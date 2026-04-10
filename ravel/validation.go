package ravel

import (
	"fmt"
	"strings"

	"github.com/alexisbouchez/ravel/api"
	"github.com/alexisbouchez/ravel/api/errdefs"
)

func validateObjectName(name string) error {
	if name == "" {
		return fmt.Errorf("name cannot be empty")
	}

	if len(name) > 63 {
		return fmt.Errorf("name cannot be longer than 63 characters")
	}

	for _, c := range name {
		if !(c == '-' || (c >= 'a' && c <= 'z') || (c >= 'A' && c <= 'Z') || (c >= '0' && c <= '9')) {
			return fmt.Errorf("name can only contain alphanumeric characters and dashes")
		}
	}

	if name[0] == '-' || name[len(name)-1] == '-' {
		return fmt.Errorf("name cannot start or end with a dash")
	}

	return nil
}

// ValidateMetadata validates metadata labels and annotations
func ValidateMetadata(metadata *api.Metadata) error {
	if metadata == nil {
		return nil
	}

	if err := validateLabels(metadata.Labels); err != nil {
		return err
	}

	if err := validateAnnotations(metadata.Annotations); err != nil {
		return err
	}

	return nil
}

// validateLabels validates label key-value pairs
func validateLabels(labels map[string]string) error {
	if len(labels) > api.MaxLabels {
		return errdefs.NewInvalidArgument(
			fmt.Sprintf("too many labels: maximum %d allowed, got %d", api.MaxLabels, len(labels)),
			&errdefs.ErrorDetail{
				Message:  fmt.Sprintf("maximum %d labels allowed", api.MaxLabels),
				Location: "metadata.labels",
				Value:    fmt.Sprintf("%d labels", len(labels)),
			},
		)
	}

	for key, value := range labels {
		if err := validateLabelKey(key); err != nil {
			return err
		}
		if err := validateLabelValue(value); err != nil {
			return err
		}
	}

	return nil
}

// validateAnnotations validates annotation key-value pairs
func validateAnnotations(annotations map[string]string) error {
	if len(annotations) > api.MaxAnnotations {
		return errdefs.NewInvalidArgument(
			fmt.Sprintf("too many annotations: maximum %d allowed, got %d", api.MaxAnnotations, len(annotations)),
			&errdefs.ErrorDetail{
				Message:  fmt.Sprintf("maximum %d annotations allowed", api.MaxAnnotations),
				Location: "metadata.annotations",
				Value:    fmt.Sprintf("%d annotations", len(annotations)),
			},
		)
	}

	for key, value := range annotations {
		if err := validateAnnotationKey(key); err != nil {
			return err
		}
		if err := validateAnnotationValue(value); err != nil {
			return err
		}
	}

	return nil
}

// validateLabelKey validates a label key
func validateLabelKey(key string) error {
	if key == "" {
		return errdefs.NewInvalidArgument(
			"label key cannot be empty",
			&errdefs.ErrorDetail{
				Message:  "label key cannot be empty",
				Location: "metadata.labels",
			},
		)
	}

	if len(key) > api.MaxLabelKeyLength {
		return errdefs.NewInvalidArgument(
			fmt.Sprintf("label key exceeds maximum length: %d characters allowed, got %d", api.MaxLabelKeyLength, len(key)),
			&errdefs.ErrorDetail{
				Message:  fmt.Sprintf("label key exceeds maximum length of %d characters", api.MaxLabelKeyLength),
				Location: "metadata.labels",
				Value:    key,
			},
		)
	}

	if strings.HasPrefix(key, api.ReservedMetadataPrefix) {
		return errdefs.NewInvalidArgument(
			fmt.Sprintf("label key uses reserved prefix '%s'", api.ReservedMetadataPrefix),
			&errdefs.ErrorDetail{
				Message:  fmt.Sprintf("label keys cannot start with reserved prefix '%s'", api.ReservedMetadataPrefix),
				Location: "metadata.labels",
				Value:    key,
			},
		)
	}

	// Validate key contains only allowed characters: alphanumeric, hyphens, underscores, dots, forward slashes
	for _, c := range key {
		if !((c >= 'a' && c <= 'z') || (c >= 'A' && c <= 'Z') || (c >= '0' && c <= '9') || c == '-' || c == '_' || c == '.' || c == '/') {
			return errdefs.NewInvalidArgument(
				"label key contains invalid characters",
				&errdefs.ErrorDetail{
					Message:  "label keys can only contain alphanumeric characters, hyphens, underscores, dots, and forward slashes",
					Location: "metadata.labels",
					Value:    key,
				},
			)
		}
	}

	return nil
}

// validateLabelValue validates a label value
func validateLabelValue(value string) error {
	if len(value) > api.MaxLabelValueLength {
		return errdefs.NewInvalidArgument(
			fmt.Sprintf("label value exceeds maximum length: %d characters allowed, got %d", api.MaxLabelValueLength, len(value)),
			&errdefs.ErrorDetail{
				Message:  fmt.Sprintf("label value exceeds maximum length of %d characters", api.MaxLabelValueLength),
				Location: "metadata.labels",
			},
		)
	}
	return nil
}

// validateAnnotationKey validates an annotation key
func validateAnnotationKey(key string) error {
	if key == "" {
		return errdefs.NewInvalidArgument(
			"annotation key cannot be empty",
			&errdefs.ErrorDetail{
				Message:  "annotation key cannot be empty",
				Location: "metadata.annotations",
			},
		)
	}

	if len(key) > api.MaxAnnotationKeyLength {
		return errdefs.NewInvalidArgument(
			fmt.Sprintf("annotation key exceeds maximum length: %d characters allowed, got %d", api.MaxAnnotationKeyLength, len(key)),
			&errdefs.ErrorDetail{
				Message:  fmt.Sprintf("annotation key exceeds maximum length of %d characters", api.MaxAnnotationKeyLength),
				Location: "metadata.annotations",
				Value:    key,
			},
		)
	}

	if strings.HasPrefix(key, api.ReservedMetadataPrefix) {
		return errdefs.NewInvalidArgument(
			fmt.Sprintf("annotation key uses reserved prefix '%s'", api.ReservedMetadataPrefix),
			&errdefs.ErrorDetail{
				Message:  fmt.Sprintf("annotation keys cannot start with reserved prefix '%s'", api.ReservedMetadataPrefix),
				Location: "metadata.annotations",
				Value:    key,
			},
		)
	}

	// Validate key contains only allowed characters: alphanumeric, hyphens, underscores, dots, forward slashes
	for _, c := range key {
		if !((c >= 'a' && c <= 'z') || (c >= 'A' && c <= 'Z') || (c >= '0' && c <= '9') || c == '-' || c == '_' || c == '.' || c == '/') {
			return errdefs.NewInvalidArgument(
				"annotation key contains invalid characters",
				&errdefs.ErrorDetail{
					Message:  "annotation keys can only contain alphanumeric characters, hyphens, underscores, dots, and forward slashes",
					Location: "metadata.annotations",
					Value:    key,
				},
			)
		}
	}

	return nil
}

// validateAnnotationValue validates an annotation value
func validateAnnotationValue(value string) error {
	if len(value) > api.MaxAnnotationValueLength {
		return errdefs.NewInvalidArgument(
			fmt.Sprintf("annotation value exceeds maximum length: %d characters allowed, got %d", api.MaxAnnotationValueLength, len(value)),
			&errdefs.ErrorDetail{
				Message:  fmt.Sprintf("annotation value exceeds maximum length of %d characters", api.MaxAnnotationValueLength),
				Location: "metadata.annotations",
			},
		)
	}
	return nil
}
