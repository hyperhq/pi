package hyper

import (
	"time"

	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/runtime/schema"
)

// volume
type VolumeCreateRequest struct {
	Name string `json:"name"`
	Zone string `json:"zone"`
	Size int    `json:"size"`
}

func (v *VolumeCreateRequest) GetObjectKind() schema.ObjectKind {
	return nil
}

func (v *VolumeCreateRequest) DeepCopyObject() runtime.Object {
	return nil
}

type VolumeData struct {
	ID     string
	Tenant string
	GCP    GCPDiskInfo
	Disk   VolumeResponse
}

type GCPDiskInfo struct {
	Name string
	ID   string
	Zone string
}

type VolumeResponse struct {
	Name      string    `json:"name"`
	Size      int       `json:"size"`
	Zone      string    `json:"zone"`
	Pod       string    `json:"pod"`
	CreatedAt time.Time `json:"createdAt"`
}

//fip
type FipRenameRequest struct {
	Name string `json:"name"`
}

type FipListResponse struct {
	Fip       string    `json:"fip"`
	Name      string    `json:"name"`
	Protocols []string  `json:"protocols"`
	CreatedAt time.Time `json:"createdAt"`
}

type FipGetResponse struct {
	Fip       string    `json:"fip"`
	Name      string    `json:"name"`
	Protocols []string  `json:"protocols"`
	CreatedAt time.Time `json:"createdAt"`
	Pods      []string  `json:"pods"`
}
