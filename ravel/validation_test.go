package ravel

import (
	"strings"
	"testing"

	"github.com/alexisbouchez/ravel/api"
)

func TestValidateVolumes(t *testing.T) {
	tests := []struct {
		name    string
		volumes []api.VolumeMount
		wantErr bool
		errMsg  string
	}{
		{
			name:    "empty volumes",
			volumes: []api.VolumeMount{},
			wantErr: false,
		},
		{
			name: "valid single volume",
			volumes: []api.VolumeMount{
				{Name: "disk1", Path: "/data"},
			},
			wantErr: false,
		},
		{
			name: "valid multiple volumes",
			volumes: []api.VolumeMount{
				{Name: "disk1", Path: "/data"},
				{Name: "disk2", Path: "/logs"},
			},
			wantErr: false,
		},
		{
			name: "too many volumes",
			volumes: []api.VolumeMount{
				{Name: "disk1", Path: "/data1"},
				{Name: "disk2", Path: "/data2"},
				{Name: "disk3", Path: "/data3"},
				{Name: "disk4", Path: "/data4"},
				{Name: "disk5", Path: "/data5"},
				{Name: "disk6", Path: "/data6"},
				{Name: "disk7", Path: "/data7"},
				{Name: "disk8", Path: "/data8"},
				{Name: "disk9", Path: "/data9"},
				{Name: "disk10", Path: "/data10"},
				{Name: "disk11", Path: "/data11"},
			},
			wantErr: true,
			errMsg:  "Too many volumes",
		},
		{
			name: "empty name",
			volumes: []api.VolumeMount{
				{Name: "", Path: "/data"},
			},
			wantErr: true,
			errMsg:  "Volume name cannot be empty",
		},
		{
			name: "duplicate names",
			volumes: []api.VolumeMount{
				{Name: "disk1", Path: "/data"},
				{Name: "disk1", Path: "/logs"},
			},
			wantErr: true,
			errMsg:  "Duplicate volume name",
		},
		{
			name: "empty path",
			volumes: []api.VolumeMount{
				{Name: "disk1", Path: ""},
			},
			wantErr: true,
			errMsg:  "Volume path cannot be empty",
		},
		{
			name: "relative path",
			volumes: []api.VolumeMount{
				{Name: "disk1", Path: "data"},
			},
			wantErr: true,
			errMsg:  "Volume path must be absolute",
		},
		{
			name: "duplicate paths",
			volumes: []api.VolumeMount{
				{Name: "disk1", Path: "/data"},
				{Name: "disk2", Path: "/data"},
			},
			wantErr: true,
			errMsg:  "Duplicate volume path",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := validateVolumes(tt.volumes)
			if tt.wantErr {
				if err == nil {
					t.Errorf("validateVolumes() expected error containing %q, got nil", tt.errMsg)
					return
				}
				if tt.errMsg != "" && !strings.Contains(err.Error(), tt.errMsg) {
					t.Errorf("validateVolumes() error = %v, want error containing %q", err, tt.errMsg)
				}
			} else {
				if err != nil {
					t.Errorf("validateVolumes() unexpected error = %v", err)
				}
			}
		})
	}
}

func TestValidateMetadata(t *testing.T) {
	tests := []struct {
		name     string
		metadata *api.Metadata
		wantErr  bool
	}{
		{
			name:     "nil metadata",
			metadata: nil,
			wantErr:  false,
		},
		{
			name: "empty metadata",
			metadata: &api.Metadata{
				Labels:      map[string]string{},
				Annotations: map[string]string{},
			},
			wantErr: false,
		},
		{
			name: "valid labels",
			metadata: &api.Metadata{
				Labels: map[string]string{
					"env":  "prod",
					"team": "backend",
				},
			},
			wantErr: false,
		},
		{
			name: "too many labels",
			metadata: &api.Metadata{
				Labels: func() map[string]string {
					m := make(map[string]string)
					for i := 0; i < 65; i++ {
						m[string(rune('a'+i%26))+string(rune('0'+i/26))] = "value"
					}
					return m
				}(),
			},
			wantErr: true,
		},
		{
			name: "label key too long",
			metadata: &api.Metadata{
				Labels: map[string]string{
					string(make([]byte, 64)): "value",
				},
			},
			wantErr: true,
		},
		{
			name: "reserved prefix",
			metadata: &api.Metadata{
				Labels: map[string]string{
					"ravel.internal": "value",
				},
			},
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := ValidateMetadata(tt.metadata)
			if (err != nil) != tt.wantErr {
				t.Errorf("ValidateMetadata() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}
