package hyper

import (
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"net/url"
	"strings"
)

type FipCli struct {
	hyperCli *HyperConn
}

func NewFipCli(client *HyperConn) *FipCli {
	return &FipCli{
		hyperCli: client,
	}
}

func (f *FipCli) AllocateFip(protocols []string, count int) (int, []FipListResponse, error) {
	var (
		result     string
		httpStatus int
		err        error
	)
	method := "POST"
	endpoint := fmt.Sprintf("/api/v1/hyper/fips?count=%v", count)
	if len(protocols) == 0 {
		result, httpStatus, err = f.hyperCli.SockRequest(method, endpoint, nil, "")
	} else {
		data := fmt.Sprintf(`{"protocols":"%v"}`, strings.Join(protocols, ","))
		result, httpStatus, err = f.hyperCli.SockRequest(method, endpoint, strings.NewReader(data), "application/json")
	}
	if err != nil {
		log.Fatalf("send request error: %v", err)
	} else if httpStatus != http.StatusCreated {
		log.Fatalf("response error: %v - %v", httpStatus, result)
	}
	var fipListAllocated []FipListResponse
	if err = json.Unmarshal([]byte(result), &fipListAllocated); err != nil {
		log.Fatalf("failed to parse allocated fip list")
	}
	return httpStatus, fipListAllocated, nil
}

func (f *FipCli) ListFips() (int, []FipListResponse, error) {
	method := "GET"
	endpoint := "/api/v1/hyper/fips"

	result, httpStatus, err := f.hyperCli.SockRequest(method, endpoint, nil, "")
	if err != nil {
		log.Fatalf("send request error: %v", err)
	} else if httpStatus != http.StatusOK {
		log.Fatalf("response error: %v - %v", httpStatus, result)
	}
	var fipList []FipListResponse
	json.Unmarshal([]byte(result), &fipList)
	if len(fipList) == 0 {
		log.Println(result)
	}
	return httpStatus, fipList, nil
}

func (f *FipCli) GetFip(ip string) (int, *FipGetResponse, error) {
	if ip == "" {
		log.Fatal("Please specify ip")
	}

	method := "GET"
	endpoint := fmt.Sprintf("/api/v1/hyper/fips/%v", url.QueryEscape(ip))

	result, httpStatus, err := f.hyperCli.SockRequest(method, endpoint, nil, "")
	if err != nil {
		log.Fatalf("send request error: %v", err)
	} else if httpStatus != http.StatusOK {
		log.Fatalf("response error: %v - %v", httpStatus, result)
	}
	var fip FipGetResponse
	err = json.Unmarshal([]byte(result), &fip)
	if err != nil {
		log.Fatalf("failed to convert result to fip:%v", err)
	}
	return httpStatus, &fip, nil
}

func (f *FipCli) NameFip(ip, name string) (int, string, error) {
	if ip == "" {
		log.Fatal("Please specify ip")
	}
	if name == "" {
		log.Fatal("Please specify --name")
	}
	method := "POST"
	endpoint := fmt.Sprintf("/api/v1/hyper/fips/%v", ip)
	data := fmt.Sprintf(`{"name":"%v"}`, name)
	result, httpStatus, err := f.hyperCli.SockRequest(method, endpoint, strings.NewReader(data), "application/json")
	if err != nil {
		log.Fatalf("send request error: %v", err)
	} else if httpStatus != http.StatusNoContent {
		log.Fatalf("response error: %v - %v", httpStatus, result)
	}
	return httpStatus, result, nil
}

func (f *FipCli) ReleaseFip(ip string) (int, string) {
	if ip == "" {
		log.Fatal("Please specify ip")
	}

	method := "DELETE"
	endpoint := fmt.Sprintf("/api/v1/hyper/fips/%v", ip)

	result, httpStatus, err := f.hyperCli.SockRequest(method, endpoint, nil, "")
	if err != nil {
		log.Fatalf("send request error: %v", err)
	} else if httpStatus != http.StatusNoContent {
		log.Fatalf("response error: %v - %v", httpStatus, result)
	}
	return httpStatus, result
}

func (f *FipCli) ReleaseAllFips() {
	method := "GET"
	endpoint := "/api/v1/hyper/fips"

	result, httpStatus, err := f.hyperCli.SockRequest(method, endpoint, nil, "")
	if err != nil {
		log.Fatalf("send request error: %v", err)
	} else if httpStatus != http.StatusOK {
		log.Fatalf("response error: %v - %v", httpStatus, result)
	}

	var fipList []FipListResponse
	err = json.Unmarshal([]byte(result), &fipList)
	if err != nil {
		log.Fatalf("failed to parse fip list:%v", err)
	}
	for _, i := range fipList {
		log.Printf("delete fip:%v", i.Fip)
		f.ReleaseFip(i.Fip)
	}
}
