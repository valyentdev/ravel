package api

// Metadata represents labels and annotations for resources (fleets and machines).
// Labels are key-value pairs used for organization and filtering.
// Annotations are key-value pairs for storing larger metadata like descriptions or JSON data.
type Metadata struct {
	Labels      map[string]string `json:"labels,omitempty"`
	Annotations map[string]string `json:"annotations,omitempty"`
}

// Metadata validation constants
const (
	// MaxLabels is the maximum number of labels allowed per resource
	MaxLabels = 64
	// MaxAnnotations is the maximum number of annotations allowed per resource
	MaxAnnotations = 32
	// MaxLabelKeyLength is the maximum length for label keys
	MaxLabelKeyLength = 63
	// MaxLabelValueLength is the maximum length for label values
	MaxLabelValueLength = 255
	// MaxAnnotationKeyLength is the maximum length for annotation keys
	MaxAnnotationKeyLength = 253
	// MaxAnnotationValueLength is the maximum length for annotation values (64KB)
	MaxAnnotationValueLength = 65536
	// ReservedMetadataPrefix is the reserved prefix for system metadata
	ReservedMetadataPrefix = "ravel."
)

// UpdateFleetMetadataPayload is the request payload for updating fleet metadata
type UpdateFleetMetadataPayload struct {
	Metadata Metadata `json:"metadata"`
}

// UpdateMachineMetadataPayload is the request payload for updating machine metadata
type UpdateMachineMetadataPayload struct {
	Metadata Metadata `json:"metadata"`
}
