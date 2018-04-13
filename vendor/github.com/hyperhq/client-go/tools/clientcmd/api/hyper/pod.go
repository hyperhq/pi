package hyper

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"log"
	"net/http"
	"strings"

	"github.com/hyperhq/hyper-api/types"

	"github.com/ghodss/yaml"
	"github.com/golang/glog"
	"k8s.io/api/core/v1"
)

type PodCli struct {
	hyperCli *HyperConn
}

type PodExecResponse struct {
	Result string
}

type ContainerExecCreateResponse struct {
	ID string
}

func NewPodCli(client *HyperConn) *PodCli {
	return &PodCli{
		hyperCli: client,
	}
}

func (p *PodCli) CreatePod(podFile string, podName string) {
	if podFile == "" {
		log.Fatalf("Please specify --file")
	}
	//read yaml
	ymlData, err := ioutil.ReadFile(podFile)
	if err != nil {
		log.Fatalf("read yaml file '%v' error: %v", podFile, err)
	}

	//conver to json
	jsonData, err := yaml.YAMLToJSON(ymlData)
	if err != nil {
		log.Fatalf("convert yaml to json error: %v", err)
	}

	var newPod v1.Pod
	err = json.Unmarshal([]byte(jsonData), &newPod)
	if err != nil {
		log.Fatalf("convert json to pod error:%v", err)
	}
	if podName != "" {
		newPod.Name = podName
	}

	buf, err := json.MarshalIndent(newPod, "", "  ")
	if err != nil {
		log.Fatalf("convert pod to json error:%v", err)
	}

	//call api
	method := "POST"
	endpoint := "/api/v1/namespaces/default/pods"
	result, httpStatus, err := p.hyperCli.SockRequest(method, endpoint, strings.NewReader(string(buf)), "application/json")
	if err != nil {
		log.Fatalf("send request error: %v", err)
	} else if httpStatus != http.StatusCreated {
		log.Fatalf("response error: %v - %v", httpStatus, result)
	}

	var createdPod v1.Pod
	if err = json.Unmarshal([]byte(result), &createdPod); err != nil {
		log.Fatalf("failed to parse created pod")
	}
	fmt.Printf("pod %v created\n", createdPod.Name)
}

func (p *PodCli) ListPods() (*v1.PodList, error) {
	method := "GET"
	endpoint := "/api/v1/namespaces/default/pods"

	result, httpStatus, err := p.hyperCli.SockRequest(method, endpoint, nil, "")
	if err != nil {
		log.Fatalf("send request error: %v", err)
	} else if httpStatus != http.StatusOK {
		log.Fatalf("response error: %v - %v", httpStatus, result)
	}
	var podList v1.PodList
	json.Unmarshal([]byte(result), &podList)
	return &podList, nil
}

func (p *PodCli) GetPod(podName string) (*v1.Pod, error) {
	if podName == "" {
		log.Fatalf("Please specify pod name")
	}
	method := "GET"
	endpoint := fmt.Sprintf("/api/v1/namespaces/default/pods/%s", podName)

	result, httpStatus, err := p.hyperCli.SockRequest(method, endpoint, nil, "")
	if err != nil {
		log.Fatalf("error: %v", err)
	} else if httpStatus != http.StatusOK {
		log.Fatalf("failed: %v - %v", httpStatus, result)
	}
	var pod v1.Pod
	err = json.Unmarshal([]byte(result), &pod)
	if err != nil {
		log.Fatalf("failed to convert reesult to pod:%v", err)
	}
	return &pod, nil
}

//use k8s exec
func (p *PodCli) ExecPod(podName, containerName string, command []string) {
	if podName == "" {
		log.Fatalf("Please specify pod name")
	}
	if len(command) == 0 {
		log.Fatalf("Command can not be empty")
	}

	config := types.ExecConfig{
		User:         "",
		Privileged:   false,
		Tty:          true,
		AttachStdin:  true,
		AttachStderr: true,
		AttachStdout: true,
		Detach:       false,
		DetachKeys:   "",
		Cmd:          command,
	}

	jsonData, err := json.MarshalIndent(config, "", "  ")
	if err != nil {
		log.Fatalf("exec config error:%v", err)
	}

	//call exec
	method := "POST"
	endpoint := fmt.Sprintf("/api/v1/namespaces/default/pods/%v/exec2?container=%v", podName, containerName)
	result, httpStatus, err := p.hyperCli.SockRequest(method, endpoint, strings.NewReader(string(jsonData)), "application/json")
	if err != nil {
		log.Fatalf("send request error: %v", err)
	} else if httpStatus != http.StatusCreated {
		log.Fatalf("response error: %v - %v", httpStatus, result)
	}
	var resultData PodExecResponse
	if err = json.Unmarshal([]byte(result), &resultData); err == nil {
		fmt.Println(resultData.Result)
	} else {
		log.Fatalf("parse exec result error:%v", err)
	}
}

//use hyper exec: support hijack connection
func (p *PodCli) HyperExecPod(podName, containerName string, command []string) {
	if podName == "" {
		log.Fatalf("Please specify pod name")
	}
	if len(command) == 0 {
		log.Fatalf("Command can not be empty")
	}

	config := types.ExecConfig{
		User:         "",
		Privileged:   false,
		Tty:          true,
		AttachStdin:  true,
		AttachStderr: true,
		AttachStdout: true,
		Detach:       false,
		DetachKeys:   "",
		Cmd:          command,
	}

	jsonData, err := json.MarshalIndent(config, "", "  ")
	if err != nil {
		log.Fatalf("exec config error:%v", err)
	}

	//call exec create
	method := "POST"
	endpoint := fmt.Sprintf("/api/v1/namespaces/default/pods/%v/exec?container=%v", podName, containerName)
	result, httpStatus, err := p.hyperCli.SockRequest(method, endpoint, strings.NewReader(string(jsonData)), "application/json")
	if err != nil {
		log.Fatalf("send request error: %v", err)
	} else if httpStatus != http.StatusCreated {
		log.Fatalf("response error: %v - %v", httpStatus, result)
	}
	var resultData ContainerExecCreateResponse
	if err = json.Unmarshal([]byte(result), &resultData); err != nil {
		log.Fatalf("parse exec result error:%v", err)
	}

	execStartCheck := &types.ExecStartCheck{
		Detach: true,
		Tty:    true,
	}

	jsonData, err = json.MarshalIndent(execStartCheck, "", "  ")
	if err != nil {
		log.Fatalf("exec start error:%v", err)
	}

	payload := bytes.NewBufferString(`{"Tty":true}`)
	conn, _, err := p.hyperCli.SockRequestHijack("POST", fmt.Sprintf("/api/v1/exec/%s/start", resultData.ID), payload, "application/json")
	if err != nil {
		log.Fatalf("sockRequestHijack error:%v", err)
	}
	defer conn.Close()
	//output result realtime
	buf := make([]byte, 255)
	for {
		n, err := conn.Read(buf)
		if err != nil {
			glog.V(4).Infof("conn closed")
			return
		}
		fmt.Printf("%v", string(buf[0:n]))
	}
}

func (p *PodCli) DeletePod(podName string, gracePeriodSeconds int) {
	var (
		result     string
		httpStatus int
		err        error
	)
	method := "DELETE"
	endpoint := fmt.Sprintf("/api/v1/namespaces/default/pods/%s", podName)

	if gracePeriodSeconds != -1 {
		data := fmt.Sprintf(`{"kind":"DeleteOptions","apiVersion":"v1","gracePeriodSeconds":%v}`, gracePeriodSeconds)
		result, httpStatus, err = p.hyperCli.SockRequest(method, endpoint, strings.NewReader(data), "application/json")
	} else {
		result, httpStatus, err = p.hyperCli.SockRequest(method, endpoint, nil, "")
	}

	if err != nil {
		log.Fatalf("error: %v", err)
	} else if httpStatus != http.StatusNoContent {
		log.Fatalf("failed: %v - %v", httpStatus, result)
	}
	fmt.Printf("%v", result)
}

func (p *PodCli) DeleteAllPods(filter string, gracePeriodSeconds int) {
	method := "GET"
	endpoint := "/api/v1/namespaces/default/pods"

	result, httpStatus, err := p.hyperCli.SockRequest(method, endpoint, nil, "")
	if err != nil {
		log.Fatalf("send request error: %v", err)
	} else if httpStatus != http.StatusOK {
		log.Fatalf("response error: %v - %v", httpStatus, result)
	}
	var podList v1.PodList
	err = json.Unmarshal([]byte(result), &podList)
	if err != nil {
		log.Fatalf("parse pod list failed:%v", err)
	}
	for _, pod := range podList.Items {
		if strings.Contains(pod.Name, filter) {
			log.Printf("delete pod:%v", pod.Name)
			p.DeletePod(pod.Name, gracePeriodSeconds)
		}
	}
}
