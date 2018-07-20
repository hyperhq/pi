package hyper

import (
	"fmt"
	"log"
	"net/http"
)

type UserCli struct {
	hyperCli *HyperConn
}

func NewUserCli(client *HyperConn) *UserCli {
	return &UserCli{
		hyperCli: client,
	}
}

func (u *UserCli) UpdateDefaultZone(tenant, zone string) (int, error) {
	var (
		result     string
		httpStatus int
		err        error
	)
	method := "POST"
	endpoint := fmt.Sprintf("/api/v1/users/%v/defaultzone?zone=%v", tenant, zone)
	result, httpStatus, err = u.hyperCli.SockRequest(method, endpoint, nil, "")
	if err != nil {
		log.Fatalf("send request error: %v", err)
	} else if httpStatus != http.StatusOK {
		log.Fatalf("response error: %v - %v", httpStatus, result)
	}
	return httpStatus, nil
}
