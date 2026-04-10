package api

import "time"

// BuildStatus represents the current state of a build
type BuildStatus string

const (
	BuildStatusPending   BuildStatus = "pending"
	BuildStatusBuilding  BuildStatus = "building"
	BuildStatusPushing   BuildStatus = "pushing"
	BuildStatusCompleted BuildStatus = "completed"
	BuildStatusFailed    BuildStatus = "failed"
)

// CreateBuildPayload is the request payload for creating a new build
type CreateBuildPayload struct {
	ImageName  string            `json:"image_name"`
	Tag        string            `json:"tag,omitempty"`
	Registry   string            `json:"registry"`
	Dockerfile string            `json:"dockerfile,omitempty"`
	BuildArgs  map[string]string `json:"build_args,omitempty"`
	Target     string            `json:"target,omitempty"`
	NoCache    bool              `json:"no_cache,omitempty"`
}

// Build represents a container image build
type Build struct {
	Id          string      `json:"id"`
	Namespace   string      `json:"namespace"`
	NodeId      string      `json:"node_id,omitempty"`
	ImageName   string      `json:"image_name"`
	Tag         string      `json:"tag"`
	Registry    string      `json:"registry"`
	FullImage   string      `json:"full_image"`
	Status      BuildStatus `json:"status"`
	Digest      string      `json:"digest,omitempty"`
	Error       string      `json:"error,omitempty"`
	DurationMs  int64       `json:"duration_ms,omitempty"`
	CreatedAt   time.Time   `json:"created_at"`
	CompletedAt *time.Time  `json:"completed_at,omitempty"`
}

// BuildLog represents a single log entry from a build
type BuildLog struct {
	Timestamp time.Time `json:"timestamp"`
	Stream    string    `json:"stream"`
	Message   string    `json:"message"`
}

// AgentBuildRequest is sent to the agent to start a build
type AgentBuildRequest struct {
	Id         string            `json:"id"`
	Namespace  string            `json:"namespace"`
	ImageName  string            `json:"image_name"`
	Tag        string            `json:"tag"`
	Registry   string            `json:"registry"`
	Dockerfile string            `json:"dockerfile"`
	BuildArgs  map[string]string `json:"build_args"`
	Target     string            `json:"target"`
	NoCache    bool              `json:"no_cache"`
}

// AgentBuildResponse is returned by the agent after starting a build
type AgentBuildResponse struct {
	Id     string      `json:"id"`
	Status BuildStatus `json:"status"`
}

// AgentBuildStatusResponse is returned by the agent for build status queries
type AgentBuildStatusResponse struct {
	Id          string      `json:"id"`
	Status      BuildStatus `json:"status"`
	Digest      string      `json:"digest,omitempty"`
	Error       string      `json:"error,omitempty"`
	DurationMs  int64       `json:"duration_ms,omitempty"`
	CompletedAt *time.Time  `json:"completed_at,omitempty"`
}
