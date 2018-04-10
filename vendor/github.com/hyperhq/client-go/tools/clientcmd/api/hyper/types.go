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
type FipAllocateRequest struct {
	Count int `json:"count"`
}

func (fa *FipAllocateRequest) GetObjectKind() schema.ObjectKind {
	return nil
}
func (fa *FipAllocateRequest) DeepCopyObject() runtime.Object {
	return nil
}

type FipRenameRequest struct {
	Name string `json:"name"`
}

func (fn *FipRenameRequest) GetObjectKind() schema.ObjectKind {
	return nil
}
func (fn *FipRenameRequest) DeepCopyObject() runtime.Object {
	return nil
}

type FipResponse struct {
	Fip       string    `json:"fip"`
	Name      string    `json:"name"`
	CreatedAt time.Time `json:"createdAt"`
	Services  []string  `json:"services"`
}
