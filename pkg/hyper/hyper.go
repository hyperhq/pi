package hyper

import (
	"fmt"
	"io"
	"net/http"
	"runtime"
	"strings"

	"github.com/hyperhq/client-go/rest"
	hyperClient "github.com/hyperhq/hyper-api/client"
	"github.com/hyperhq/hypercli/pkg/term"
	"github.com/hyperhq/pi"

	"github.com/docker/go-connections/sockets"
	"github.com/docker/go-connections/tlsconfig"
	"github.com/golang/glog"
)

type HyperCli struct {
	// initializing closure
	init func() error

	// in holds the input stream and closer (io.ReadCloser) for the client.
	In io.ReadCloser
	// out holds the output stream (io.Writer) for the client.
	Out io.Writer
	// err holds the error stream (io.Writer) for the client.
	Err io.Writer
	// keyFile holds the key file as a string.
	keyFile string
	// inFd holds the file descriptor of the client's STDIN (if valid).
	inFd uintptr
	// outFd holds file descriptor of the client's STDOUT (if valid).
	outFd uintptr
	// isTerminalIn indicates whether the client's STDIN is a TTY
	IsTerminalIn bool
	// isTerminalOut indicates whether the client's STDOUT is a TTY
	IsTerminalOut bool
	// client is the http client that performs all API operations
	Client hyperClient.APIClient
	// state holds the terminal state
	state *term.State

	region string
	host   string
}

// An StatusError reports an unsuccessful exit by a command.
type StatusError struct {
	Status     string
	StatusCode int
}

func (e StatusError) Error() string {
	return fmt.Sprintf("Status: %s, Code: %d", e.Status, e.StatusCode)
}

func NewHyperCli(host string, config *rest.Config, stdin io.ReadCloser, stdout io.Writer, stderr io.Writer) (*HyperCli, error) {
	//replace default domain
	if strings.Contains(host, rest.DefaultDomain) {
		host = strings.Replace(host, "*", config.Region, 1)
		glog.V(4).Infof("NewHyperCli: replace default domain to %v", host)
	}

	tlsOptions := &tlsconfig.Options{
		InsecureSkipVerify: true,
	}
	httpClient, err := newHTTPClient(host, tlsOptions)
	if err != nil {
		return nil, err
	}
	customHeaders := map[string]string{}
	customHeaders["User-Agent"] = "Pi/" + pi.Version + " (" + runtime.GOOS + ")"
	apiVersion := "api/v1"

	client, err := hyperClient.NewClient(host, apiVersion, httpClient, customHeaders, config.AccessKey, config.SecretKey, config.Region)
	if err != nil {
		return nil, err
	}

	glog.V(7).Infof("host:%v\n ver:%v\n customHeaders:%v\n", host, apiVersion, customHeaders)

	cli := &HyperCli{
		Client: client,
		host:   host,
		In:     stdin,
		Out:    stdout,
		Err:    stderr,
	}
	if stdin != nil {
		cli.inFd, cli.IsTerminalIn = term.GetFdInfo(stdin)
	}
	if cli.Out != nil {
		cli.outFd, cli.IsTerminalOut = term.GetFdInfo(stdout)
	}
	return cli, nil
}

func newHTTPClient(host string, tlsOptions *tlsconfig.Options) (*http.Client, error) {
	if tlsOptions == nil {
		// let the api client configure the default transport.
		return nil, nil
	}

	config, err := tlsconfig.Client(*tlsOptions)
	if err != nil {
		return nil, err
	}
	tr := &http.Transport{
		TLSClientConfig: config,
	}
	proto, addr, _, err := hyperClient.ParseHost(host)
	if err != nil {
		return nil, err
	}

	sockets.ConfigureTransport(tr, proto, addr)

	return &http.Client{
		Transport: tr,
	}, nil
}
