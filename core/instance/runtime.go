package instance

type InstanceOptions struct {
	Id       string           `json:"id"`
	Metadata InstanceMetadata `json:"metadata"`
	Config   InstanceConfig   `json:"config"`
}
