package hyper

import (
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"strings"
	"time"
)

type Volume struct {
	Name       string                 // Name is the Name of the volume
	Driver     string                 // Driver is the Driver Name used to create the volume
	Mountpoint string                 // Mountpoint is the location on disk of the volume
	Status     map[string]interface{} `json:",omitempty"` // Status provides low-level status information about the volume
	Labels     map[string]string      // Labels is metadata specific to the volume
	Scope      string                 // Scope describes the level at which the volume exists (e.g. `global` for cluster-wide or `local` for machine level)

	CreatedAt time.Time
}

type VolumeCli struct {
	hyperCli *HyperConn
}

func NewVolumeCli(client *HyperConn) *VolumeCli {
	return &VolumeCli{
		hyperCli: client,
	}
}

func (v *VolumeCli) CreateVolume(volName, zone, size string) (int, *VolumeResponse, error) {
	method := "POST"
	endpoint := "/api/v1/hyper/volumes"
	data := ""
	if size == "" {
		data = fmt.Sprintf(`{"name":"%v","zone":"%v"}`, volName, zone)
	} else {
		data = fmt.Sprintf(`{"name":"%v","zone":"%v", "size": %v}`, volName, zone, size)
	}
	result, httpStatus, err := v.hyperCli.SockRequest(method, endpoint, strings.NewReader(data), "application/json")
	if err != nil {
		log.Fatalf("send request error: %v", err)
	} else if httpStatus != http.StatusCreated {
		log.Fatalf("response error: %v - %v", httpStatus, result)
	}

	var createdVolume VolumeResponse
	if err = json.Unmarshal([]byte(result), &createdVolume); err != nil {
		log.Fatalf("failded to parse created volume")
	}
	return httpStatus, &createdVolume, nil
}

func (v *VolumeCli) ListVolumes(zone string) (int, []VolumeResponse, error) {
	method := "GET"
	endpoint := fmt.Sprintf("/api/v1/hyper/volumes?zone=%v", zone)

	result, httpStatus, err := v.hyperCli.SockRequest(method, endpoint, nil, "")
	if err != nil {
		log.Fatalf("send request error: %v", err)
	} else if httpStatus != http.StatusOK {
		log.Fatalf("response error: %v - %v", httpStatus, result)
	}

	var volumeList []VolumeResponse
	err = json.Unmarshal([]byte(result), &volumeList)
	return httpStatus, volumeList, err
}

func (v *VolumeCli) GetVolume(volName, zone string) (int, *VolumeResponse, error) {
	if volName == "" {
		log.Fatal("Please specify volume name")
	}

	method := "GET"
	endpoint := fmt.Sprintf("/api/v1/hyper/volumes/%v?zone=%v", volName, zone)

	result, httpStatus, err := v.hyperCli.SockRequest(method, endpoint, nil, "")
	if err != nil {
		log.Fatalf("send request error: %v", err)
	} else if httpStatus != http.StatusOK {
		log.Fatalf("response error: %v - %v", httpStatus, result)
	}
	var vol VolumeResponse
	err = json.Unmarshal([]byte(result), &vol)
	return httpStatus, &vol, err
}

func (v *VolumeCli) DeleteVolume(volName, zone string) (int, string) {
	if volName == "" {
		log.Fatal("Please specify volume name")
	}
	method := "DELETE"
	endpoint := fmt.Sprintf("/api/v1/hyper/volumes/%v?zone=%v", volName, zone)

	result, httpStatus, err := v.hyperCli.SockRequest(method, endpoint, nil, "")
	if err != nil {
		log.Fatalf("send request error: %v", err)
	} else if httpStatus != http.StatusNoContent {
		log.Fatalf("response error: %v - %v", httpStatus, result)
	}
	return httpStatus, result
}

func (v *VolumeCli) DeleteAllVolumes(zone string) {
	method := "GET"
	endpoint := fmt.Sprintf("/api/v1/hyper/volumes?zone=%v", zone)

	result, httpStatus, err := v.hyperCli.SockRequest(method, endpoint, nil, "")
	if err != nil {
		log.Fatalf("send request error: %v", err)
	} else if httpStatus != http.StatusOK {
		log.Fatalf("response error: %v - %v", httpStatus, result)
	}
	var volumeList []VolumeResponse
	err = json.Unmarshal([]byte(result), &volumeList)
	if err != nil {
		log.Fatalf("failed to parse volume list in zone %v, error:%v", zone, err)
	}
	for _, vol := range volumeList {
		log.Printf("delete volume %v in zone %v", vol.Name, vol.Zone)
		v.DeleteVolume(vol.Name, vol.Zone)
	}
}
