package hyper

import (
	"bufio"
	"bytes"
	"crypto/tls"
	"encoding/json"
	"fmt"
	"io"
	"io/ioutil"
	"net"
	"net/http"
	"net/http/httputil"
	"net/url"
	"strings"
	"time"

	restclient "github.com/hyperhq/client-go/rest"
	"github.com/hyperhq/hyper-api/signature"

	"github.com/docker/go-connections/tlsconfig"
	"github.com/golang/glog"
)

type HyperConn struct {
	Host      string
	Region    string
	AccessKey string
	SecretKey string
}

func NewHyperConn(config *restclient.Config) *HyperConn {
	return &HyperConn{
		Host:      config.Host,
		Region:    config.Region,
		AccessKey: config.AccessKey,
		SecretKey: config.SecretKey,
	}
}

func (u *HyperConn) SockRequest(method, endpoint string, data io.Reader, contentType string) (string, int, error) {
	var postData = ""
	if data != nil {
		buf, err := ioutil.ReadAll(data)

		glog.V(4).Infof("raw post data:%v", string(buf))

		//string -> struct
		var postMap map[string]interface{}
		if err := json.Unmarshal(buf, &postMap); err != nil {
			return "", 0, fmt.Errorf("generateCURL Unmarshal error:%v", err)
		}
		//struct -> string(format json)
		buf, err = json.Marshal(postMap)
		if err != nil {
			return "", 0, fmt.Errorf("generateCURL Marshal error:%v", err)
		}
		//save post data string
		postData = string(buf)
		//post data string to io.Reader
		data = bytes.NewReader(buf)
	}

	//prepare req
	req, err := u.prepareRequest(method, endpoint, data, contentType)
	if err != nil {
		return "", 0, fmt.Errorf("prepare request error: %v", err)
	}

	if glog.V(7) {
		//generate curl command line
		curlStr := generateCURL(req, method, postData)
		glog.Infof("========================================\n%s\n----------------------------------------\n",
			strings.Join(curlStr, "\\\n"))
	}

	//call http request
	result, statusCode, err := sendRequest(req)
	if err != nil {
		return "", statusCode, err
	}
	return result, statusCode, nil
}

func (u *HyperConn) sockRawRequest(method, endpoint string, data io.Reader, contentType string) (*http.Response, error) {
	var postData = ""
	if data != nil {
		buf, err := ioutil.ReadAll(data)

		//string -> struct
		var postMap map[string]interface{}
		if err := json.Unmarshal(buf, &postMap); err != nil {
			return nil, fmt.Errorf("generateCURL Unmarshal error:%v", err)
		}
		//struct -> string(format json)
		buf, err = json.Marshal(postMap)
		if err != nil {
			return nil, fmt.Errorf("generateCURL Marshal error:%v", err)
		}
		//save post data string
		postData = string(buf)
		//post data string to io.Reader
		data = bytes.NewReader(buf)
	}

	//prepare req
	req, err := u.prepareRequest(method, endpoint, data, contentType)
	req.Header.Add("Upgrade", "")
	if err != nil {
		return nil, fmt.Errorf("prepare request error: %v", err)
	}

	if glog.V(7) {
		//generate curl command line
		curlStr := generateCURL(req, method, postData)
		glog.Infof("========================================\n%s\n----------------------------------------\n",
			strings.Join(curlStr, "\\\n"))
	}

	//call http request
	return sendRawRequest(req)
}

func (u *HyperConn) prepareRequest(method string, endpoint string, data io.Reader, contentType string) (*http.Request, error) {
	hostURL, err := url.Parse(u.Host)
	if err != nil {
		return nil, fmt.Errorf("host url format error: %v", err)
	}

	//replace default domain
	if strings.Contains(hostURL.Host, restclient.DefaultDomain) {
		hostURL.Host = strings.Replace(hostURL.Host, "*", u.Region, 1)
		glog.V(4).Infof("replace default domain to %v", hostURL.Host)
	}

	req, err := http.NewRequest(method, endpoint, data)
	if err != nil {
		return nil, fmt.Errorf("could not create new request: %v", err)
	}

	req.URL.Scheme = "tcp"
	req.URL.Host = hostURL.Host

	if contentType != "" {
		req.Header.Set("Content-Type", contentType)
	}
	//calculate sign4 for apirouter
	req = signature.Sign4(u.AccessKey, u.SecretKey, req, u.Region)
	return req, nil
}

func generateCURL(req *http.Request, method string, postData string) []string {
	var curlStr []string
	curlStr = append(curlStr, fmt.Sprint("[REQUEST]: \ncurl -v -k "))
	curlStr = append(curlStr, fmt.Sprintf("  -X %v ", method))
	for k, v := range req.Header {
		curlStr = append(curlStr, fmt.Sprintf("  -H \"%v: %v\" ", k, v[0]))
	}
	if req.Body != nil {
		curlStr = append(curlStr, fmt.Sprintf("  -d '%v' ", postData))
	}
	curlStr = append(curlStr, fmt.Sprintf("  https://%v%v", req.URL.Host, req.URL.RequestURI()))
	return curlStr
}

func sendRequest(req *http.Request) (string, int, error) {
	tlsConfig, err := tlsconfig.Client(tlsconfig.Options{
		InsecureSkipVerify: true,
	})
	if err != nil {
		return "", 0, fmt.Errorf("create TLS configuration error: %v", err)
	}

	req.URL.Scheme = "tcp"

	dialer := &net.Dialer{Timeout: time.Duration(10 * time.Second)}
	conn, err := tls.DialWithDialer(dialer, req.URL.Scheme, req.URL.Host, tlsConfig)
	if err != nil {
		return "", 0, fmt.Errorf("dial with dialer error: %v", err)
	}
	client := httputil.NewClientConn(conn, nil)
	resp, err := client.Do(req)
	if err != nil {
		return "", 0, fmt.Errorf("http request error: %v", err)
	}
	defer resp.Body.Close()
	body, err := readBody(resp.Body)
	if err != nil {
		return "", 0, fmt.Errorf("read body error: %v", err)
	}
	return string(body), resp.StatusCode, nil
}

func sendRawRequest(req *http.Request) (*http.Response, error) {
	tlsConfig, err := tlsconfig.Client(tlsconfig.Options{
		InsecureSkipVerify: true,
	})
	if err != nil {
		return nil, fmt.Errorf("create TLS configuration error: %v", err)
	}

	req.URL.Scheme = "tcp"

	dialer := &net.Dialer{Timeout: time.Duration(10 * time.Second)}
	conn, err := tls.DialWithDialer(dialer, req.URL.Scheme, req.URL.Host, tlsConfig)
	if err != nil {
		return nil, fmt.Errorf("dial with dialer error: %v", err)
	}
	client := httputil.NewClientConn(conn, nil)
	return client.Do(req)
}

func readBody(b io.ReadCloser) ([]byte, error) {
	defer b.Close()
	return ioutil.ReadAll(b)
}

////////////////////////////////////////////////
// for hijack connection
////////////////////////////////////////////////
func (u *HyperConn) SockRequestHijack(method, endpoint string, data io.Reader, ct string) (net.Conn, *bufio.Reader, error) {
	req, client, err := u.newRequestHyperConn(method, endpoint, data, ct)
	if err != nil {
		return nil, nil, err
	}

	client.Do(req)
	conn, br := client.Hijack()
	return conn, br, nil
}

func (u *HyperConn) newRequestHyperConn(method, endpoint string, data io.Reader, ct string) (*http.Request, *httputil.ClientConn, error) {
	//replace default domain
	if strings.Contains(u.Host, restclient.DefaultDomain) {
		u.Host = strings.Replace(u.Host, "*", u.Region, 1)
		glog.V(4).Infof("replace default domain to %v", u.Host)
	}

	c, err := u.sockConn(time.Duration(10 * time.Second))
	if err != nil {
		return nil, nil, fmt.Errorf("could not dial docker daemon: %v", err)
	}

	client := httputil.NewClientConn(c, nil)

	//save to postData
	postData := fmt.Sprintf("%v", data)

	req, err := http.NewRequest(method, endpoint, data)
	if err != nil {
		client.Close()
		return nil, nil, fmt.Errorf("could not create new request: %v", err)
	}

	//init
	req.URL.Scheme = "tcp"
	req.URL.Host = strings.Split(u.Host, "://")[1]

	if ct != "" {
		req.Header.Set("Content-Type", ct)
	}

	//calculate sign4 for apirouter
	req = signature.Sign4(u.AccessKey, u.SecretKey, req, u.Region)

	if glog.V(7) {
		//output curl command line
		s := fmt.Sprintln("\n--------------------------------------------------------------------------------------------")
		s += fmt.Sprintln("curl -v -k \\")
		for k, v := range req.Header {
			s += fmt.Sprintf("  --region \"%v: %v\" \\\n", k, v[0])
		}
		s += fmt.Sprintf("  -X %v \\\n", method)
		if req.Body != nil {
			s += fmt.Sprintf("  -d '%v' \\\n", postData)
		}

		s += fmt.Sprintf("  https://%v%v\n", req.URL.Host, req.URL.RequestURI())
		s += fmt.Sprintln("--------------------------------------------------------------------------------------------")
		fmt.Print(s)
	}
	return req, client, nil
}

func (u *HyperConn) sockConn(timeout time.Duration) (net.Conn, error) {
	daemon := u.Host
	daemonURL, err := url.Parse(daemon)
	if err != nil {
		return nil, fmt.Errorf("could not parse url %q: %v", daemon, err)
	}

	daemonURL.Scheme = "tcp"

	tlsConfig, err := tlsconfig.Client(tlsconfig.Options{
		InsecureSkipVerify: true,
	})
	if err != nil {
		return nil, err
	}
	dialer := &net.Dialer{Timeout: timeout}
	return tls.DialWithDialer(dialer, daemonURL.Scheme, daemonURL.Host, tlsConfig)
}
