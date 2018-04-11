package hyper

import (
	"encoding/json"
	"log"
	"net/http"
)

type InfoCli struct {
	hyperCli *HyperConn
}

func NewInfoCli(client *HyperConn) *InfoCli {
	return &InfoCli{
		hyperCli: client,
	}
}

func (f *InfoCli) GetInfo() (int, map[string]string, error) {
	method := "GET"
	endpoint := "/info"

	result, httpStatus, err := f.hyperCli.SockRequest(method, endpoint, nil, "")
	if err != nil {
		log.Fatalf("send request error: %v", err)
	} else if httpStatus != http.StatusOK {
		log.Fatalf("response error: %v - %v", httpStatus, result)
	}
	var info map[string]string
	err = json.Unmarshal([]byte(result), &info)
	if err != nil {
		log.Fatalf("failed to convert result to info:%v", err)
	}
	return httpStatus, info, nil
}
