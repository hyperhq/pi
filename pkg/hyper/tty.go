package hyper

import (
	"context"
	"errors"
	"io"
	"os"
	gosignal "os/signal"
	"runtime"
	"time"

	"github.com/hyperhq/hyper-api/client"
	"github.com/hyperhq/hyper-api/types"
	"github.com/hyperhq/hypercli/pkg/signal"
	"github.com/hyperhq/hypercli/pkg/term"

	"github.com/golang/glog"
)

// CheckTtyInput checks if we are trying to attach to a container tty
// from a non-tty client input stream, and if so, returns an error.
func (cli *HyperCli) CheckTtyInput(attachStdin, ttyMode bool) error {
	// In order to attach to a container tty, input stream for the client must
	// be a tty itself: redirecting or piping the client standard input is
	// incompatible with `docker run -t`, `docker exec -t` or `docker attach`.
	if ttyMode && attachStdin && !cli.IsTerminalIn {
		return errors.New("cannot enable tty mode on non tty input")
	}
	return nil
}

func (cli *HyperCli) SetRawTerminal() error {
	if cli.IsTerminalIn && os.Getenv("NORAW") == "" {
		state, err := term.SetRawTerminal(cli.inFd)
		if err != nil {
			return err
		}
		cli.state = state
	}
	return nil
}

func (cli *HyperCli) RestoreTerminal(in io.Closer) error {
	if cli.state != nil {
		term.RestoreTerminal(cli.inFd, cli.state)
	}
	// WARNING: DO NOT REMOVE THE OS CHECK !!!
	// For some reason this Close call blocks on darwin..
	// As the client exists right after, simply discard the close
	// until we find a better solution.
	if in != nil && runtime.GOOS != "darwin" {
		return in.Close()
	}
	return nil
}

func (cli *HyperCli) resizeTty(ctx context.Context, id string, isExec bool) {
	height, width := cli.getTtySize()
	if height == 0 && width == 0 {
		return
	}

	options := types.ResizeOptions{
		Height: height,
		Width:  width,
	}

	var err error
	if isExec {
		err = cli.Client.PodExecResize(ctx, id, options)
	}

	if err != nil {
		glog.V(4).Infof("Error resize: %s", err)
	}
}

// getExitCode perform an inspect on the container. It returns
// the running state and the exit code.
func getExitCode(ctx context.Context, cli *HyperCli, containerID string) (bool, int, error) {
	c, err := cli.Client.ContainerInspect(ctx, containerID)
	if err != nil {
		// If we can't connect, then the daemon probably died.
		if err != client.ErrConnectionFailed {
			return false, -1, err
		}
		return false, -1, nil
	}

	return c.State.Running, c.State.ExitCode, nil
}

// getExecExitCode perform an inspect on the exec command. It returns
// the running state and the exit code.
func (cli *HyperCli) GetExecExitCode(ctx context.Context, execID string) (bool, int, error) {
	resp, err := cli.Client.PodExecInspect(ctx, execID)
	if err != nil {
		// If we can't connect, then the daemon probably died.
		if err != client.ErrConnectionFailed {
			return false, -1, err
		}
		return false, -1, nil
	}

	return resp.Running, resp.ExitCode, nil
}

func (cli *HyperCli) MonitorTtySize(ctx context.Context, id string, isExec bool) error {
	cli.resizeTty(ctx, id, isExec)

	if runtime.GOOS == "windows" {
		go func() {
			prevH, prevW := cli.getTtySize()
			for {
				time.Sleep(time.Millisecond * 250)
				h, w := cli.getTtySize()

				if prevW != w || prevH != h {
					cli.resizeTty(ctx, id, isExec)
				}
				prevH = h
				prevW = w
			}
		}()
	} else {
		sigchan := make(chan os.Signal, 1)
		gosignal.Notify(sigchan, signal.SIGWINCH)
		go func() {
			for range sigchan {
				cli.resizeTty(ctx, id, isExec)
			}
		}()
	}
	return nil
}

func (cli *HyperCli) getTtySize() (int, int) {
	if !cli.IsTerminalOut {
		return 0, 0
	}
	ws, err := term.GetWinsize(cli.outFd)
	if err != nil {
		glog.V(4).Infof("Error getting size: %s", err)
		if ws == nil {
			return 0, 0
		}
	}
	return int(ws.Height), int(ws.Width)
}
