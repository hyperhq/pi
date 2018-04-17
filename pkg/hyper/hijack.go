package hyper

import (
	"io"

	"github.com/hyperhq/hyper-api/types"
	"github.com/hyperhq/hypercli/pkg/stdcopy"

	"github.com/golang/glog"
)

func (cli *HyperCli) HoldHijackedConnection(tty bool, inputStream io.ReadCloser, outputStream, errorStream io.Writer, resp types.HijackedResponse) error {
	var err error
	receiveStdout := make(chan error, 1)
	if outputStream != nil || errorStream != nil {
		go func() {
			// When TTY is ON, use regular copy
			if tty && outputStream != nil {
				_, err = io.Copy(outputStream, resp.Reader)
			} else {
				_, err = stdcopy.StdCopy(outputStream, errorStream, resp.Reader)
			}
			glog.V(4).Infof("[hijack] End of stdout")
			receiveStdout <- err
		}()
	}

	stdinDone := make(chan struct{})
	go func() {
		if inputStream != nil {
			io.Copy(resp.Conn, inputStream)
			glog.V(4).Infof("[hijack] End of stdin")
		}

		if err := resp.CloseWrite(); err != nil {
			glog.V(4).Infof("Couldn't send EOF: %s", err)
		}
		close(stdinDone)
	}()

	select {
	case err := <-receiveStdout:
		if err != nil {
			glog.V(4).Infof("Error receiveStdout: %s", err)
			return err
		}
	case <-stdinDone:
		if outputStream != nil || errorStream != nil {
			if err := <-receiveStdout; err != nil {
				glog.V(4).Infof("Error receiveStdout: %s", err)
				return err
			}
		}
	}

	return nil
}
