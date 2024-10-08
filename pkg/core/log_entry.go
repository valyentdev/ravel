package core

type LogEntry struct {
	Timestamp  int64  `json:"timestamp,omitempty"`
	InstanceId string `json:"instance_id,omitempty"`
	Source     string `json:"source,omitempty"`
	Level      string `json:"level,omitempty"`
	Message    string `json:"message,omitempty"`
}
