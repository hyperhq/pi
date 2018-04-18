/*
Copyright 2014 The Kubernetes Authors.

Licensed under the Apache License, Version 2.0 (the "License");
you may not use this file except in compliance with the License.
You may obtain a copy of the License at

    http://www.apache.org/licenses/LICENSE-2.0

Unless required by applicable law or agreed to in writing, software
distributed under the License is distributed on an "AS IS" BASIS,
WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
See the License for the specific language governing permissions and
limitations under the License.
*/

package cmd

import (
	"fmt"
	"io"
	"net/url"

	restclient "github.com/hyperhq/client-go/rest"
	"github.com/hyperhq/client-go/tools/remotecommand"
	"github.com/hyperhq/hyper-api/types"
	"github.com/hyperhq/hypercli/pkg/promise"
	"github.com/hyperhq/pi/pkg/hyper"
	"github.com/hyperhq/pi/pkg/pi/cmd/templates"
	"github.com/hyperhq/pi/pkg/pi/cmd/util"
	cmdutil "github.com/hyperhq/pi/pkg/pi/cmd/util"
	"github.com/hyperhq/pi/pkg/pi/util/i18n"
	"github.com/hyperhq/pi/pkg/pi/util/term"
	"k8s.io/kubernetes/pkg/util/interrupt"

	dockerterm "github.com/docker/docker/pkg/term"
	"github.com/golang/glog"
	"github.com/spf13/cobra"
	"golang.org/x/net/context"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/kubernetes/pkg/api/legacyscheme"
	api "k8s.io/kubernetes/pkg/apis/core"
	coreclient "k8s.io/kubernetes/pkg/client/clientset_generated/internalclientset/typed/core/internalversion"
)

var (
	exec_example = templates.Examples(i18n.T(`
		# Get output from running 'date' from pod 123456-7890, using the first container by default
		pi exec 123456-7890 date

		# Get output from running 'date' in ruby-container from pod 123456-7890
		pi exec 123456-7890 -c ruby-container date

		# Switch to raw terminal mode, sends stdin to 'bash' in ruby-container from pod 123456-7890
		# and sends stdout/stderr from 'bash' back to the client
		pi exec 123456-7890 -c ruby-container -i -t -- bash -il

		# List contents of /usr from the first container of pod 123456-7890 and sort by modification time.
		# If the command you want to execute in the pod has any flags in common (e.g. -i),
		# you must use two dashes (--) to separate your command's flags/arguments.
		# Also note, do not surround your command and its flags/arguments with quotes
		# unless that is how you would execute it normally (i.e., do ls -t /usr, not "ls -t /usr").
		pi exec 123456-7890 -i -t -- ls -t /usr
		`))
)

const (
	execUsageStr = "expected 'exec POD_NAME COMMAND [ARG1] [ARG2] ... [ARGN]'.\nPOD_NAME and COMMAND are required arguments for the exec command"
)

func NewCmdExec(f cmdutil.Factory, cmdIn io.Reader, cmdOut, cmdErr io.Writer) *cobra.Command {
	options := &ExecOptions{
		StreamOptions: StreamOptions{
			In:  cmdIn,
			Out: cmdOut,
			Err: cmdErr,
		},

		Executor: &DefaultRemoteExecutor{},
	}
	cmd := &cobra.Command{
		Use:     "exec POD [-c CONTAINER] -- COMMAND [args...]",
		Short:   i18n.T("Execute a command in a container"),
		Long:    "Execute a command in a container.",
		Example: exec_example,
		Run: func(cmd *cobra.Command, args []string) {
			argsLenAtDash := cmd.ArgsLenAtDash()
			cmdutil.CheckErr(options.Complete(f, cmd, args, argsLenAtDash))
			cmdutil.CheckErr(options.Validate())
			//cmdutil.CheckErr(options.Run())
			cmdutil.CheckErr(options.RunHyper(f))
		},
	}
	//cmd.Flags().StringVarP(&options.PodName, "pod", "p", "", "Pod name")
	// TODO support UID
	cmd.Flags().StringVarP(&options.ContainerName, "container", "c", "", "Container name. If omitted, the first container in the pod will be chosen")
	cmd.Flags().BoolVarP(&options.Stdin, "stdin", "i", false, "Pass stdin to the container")
	cmd.Flags().BoolVarP(&options.TTY, "tty", "t", false, "Stdin is a TTY")
	return cmd
}

// RemoteExecutor defines the interface accepted by the Exec command - provided for test stubbing
type RemoteExecutor interface {
	Execute(method string, url *url.URL, config *restclient.Config, stdin io.Reader, stdout, stderr io.Writer, tty bool, terminalSizeQueue remotecommand.TerminalSizeQueue) error
}

// DefaultRemoteExecutor is the standard implementation of remote command execution
type DefaultRemoteExecutor struct{}

func (*DefaultRemoteExecutor) Execute(method string, url *url.URL, config *restclient.Config, stdin io.Reader, stdout, stderr io.Writer, tty bool, terminalSizeQueue remotecommand.TerminalSizeQueue) error {
	exec, err := remotecommand.NewSPDYExecutor(config, method, url)
	if err != nil {
		return err
	}
	return exec.Stream(remotecommand.StreamOptions{
		Stdin:             stdin,
		Stdout:            stdout,
		Stderr:            stderr,
		Tty:               tty,
		TerminalSizeQueue: terminalSizeQueue,
	})
}

type StreamOptions struct {
	Namespace     string
	PodName       string
	ContainerName string
	Stdin         bool
	TTY           bool
	// minimize unnecessary output
	Quiet bool
	// InterruptParent, if set, is used to handle interrupts while attached
	InterruptParent *interrupt.Handler
	In              io.Reader
	Out             io.Writer
	Err             io.Writer

	// for testing
	overrideStreams func() (io.ReadCloser, io.Writer, io.Writer)
	isTerminalIn    func(t term.TTY) bool
}

// ExecOptions declare the arguments accepted by the Exec command
type ExecOptions struct {
	StreamOptions

	Command []string

	FullCmdName       string
	SuggestedCmdUsage string

	Executor  RemoteExecutor
	PodClient coreclient.PodsGetter
	Config    *restclient.Config
}

// Complete verifies command line arguments and loads data from the command environment
func (p *ExecOptions) Complete(f cmdutil.Factory, cmd *cobra.Command, argsIn []string, argsLenAtDash int) error {
	// Let pi exec follow rules for `--`, see #13004 issue
	if len(p.PodName) == 0 && (len(argsIn) == 0 || argsLenAtDash == 0) {
		return cmdutil.UsageErrorf(cmd, execUsageStr)
	}
	if len(p.PodName) != 0 {
		printDeprecationWarning("exec POD_NAME", "-p POD_NAME")
		if len(argsIn) < 1 {
			return cmdutil.UsageErrorf(cmd, execUsageStr)
		}
		p.Command = argsIn
	} else {
		p.PodName = argsIn[0]
		p.Command = argsIn[1:]
		if len(p.Command) < 1 {
			return cmdutil.UsageErrorf(cmd, execUsageStr)
		}
	}

	namespace, _, err := f.DefaultNamespace()
	if err != nil {
		return err
	}
	p.Namespace = namespace

	cmdParent := cmd.Parent()
	if cmdParent != nil {
		p.FullCmdName = cmdParent.CommandPath()
	}
	if len(p.FullCmdName) > 0 && cmdutil.IsSiblingCommandExists(cmd, "describe") {
		p.SuggestedCmdUsage = fmt.Sprintf("Use '%s describe pod/%s -n %s' to see all of the containers in this pod.", p.FullCmdName, p.PodName, p.Namespace)
	}

	config, err := f.ClientConfig()
	if err != nil {
		return err
	}
	p.Config = config

	clientset, err := f.ClientSet()
	if err != nil {
		return err
	}
	p.PodClient = clientset.Core()

	return nil
}

// Validate checks that the provided exec options are specified.
func (p *ExecOptions) Validate() error {
	if len(p.PodName) == 0 {
		return fmt.Errorf("pod name must be specified")
	}
	if len(p.Command) == 0 {
		return fmt.Errorf("you must specify at least one command for the container")
	}
	if p.Out == nil || p.Err == nil {
		return fmt.Errorf("both output and error output must be provided")
	}
	if p.Executor == nil || p.PodClient == nil || p.Config == nil {
		return fmt.Errorf("client, client config, and executor must be provided")
	}
	return nil
}

func (o *StreamOptions) setupTTY() term.TTY {
	t := term.TTY{
		Parent: o.InterruptParent,
		Out:    o.Out,
	}

	if !o.Stdin {
		// need to nil out o.In to make sure we don't create a stream for stdin
		o.In = nil
		o.TTY = false
		return t
	}

	t.In = o.In
	if !o.TTY {
		return t
	}

	if o.isTerminalIn == nil {
		o.isTerminalIn = func(tty term.TTY) bool {
			return tty.IsTerminalIn()
		}
	}
	if !o.isTerminalIn(t) {
		o.TTY = false

		if o.Err != nil {
			fmt.Fprintln(o.Err, "Unable to use a TTY - input is not a terminal or the right kind of file")
		}

		return t
	}

	// if we get to here, the user wants to attach stdin, wants a TTY, and o.In is a terminal, so we
	// can safely set t.Raw to true
	t.Raw = true

	if o.overrideStreams == nil {
		// use dockerterm.StdStreams() to get the right I/O handles on Windows
		o.overrideStreams = dockerterm.StdStreams
	}
	stdin, stdout, _ := o.overrideStreams()
	o.In = stdin
	t.In = stdin
	if o.Out != nil {
		o.Out = stdout
		t.Out = stdout
	}

	return t
}

// Run executes a validated remote execution against a pod.
func (p *ExecOptions) Run() error {
	pod, err := p.PodClient.Pods(p.Namespace).Get(p.PodName, metav1.GetOptions{})
	if err != nil {
		return err
	}

	if pod.Status.Phase == api.PodSucceeded || pod.Status.Phase == api.PodFailed {
		return fmt.Errorf("cannot exec into a container in a completed pod; current phase is %s", pod.Status.Phase)
	}

	containerName := p.ContainerName
	if len(containerName) == 0 {
		if len(pod.Spec.Containers) > 1 {
			usageString := fmt.Sprintf("Defaulting container name to %s.", pod.Spec.Containers[0].Name)
			if len(p.SuggestedCmdUsage) > 0 {
				usageString = fmt.Sprintf("%s\n%s", usageString, p.SuggestedCmdUsage)
			}
			fmt.Fprintf(p.Err, "%s\n", usageString)
		}
		containerName = pod.Spec.Containers[0].Name
	}

	// ensure we can recover the terminal while attached
	t := p.setupTTY()

	var sizeQueue remotecommand.TerminalSizeQueue
	if t.Raw {
		// this call spawns a goroutine to monitor/update the terminal size
		sizeQueue = t.MonitorSize(t.GetSize())

		// unset p.Err if it was previously set because both stdout and stderr go over p.Out when tty is
		// true
		p.Err = nil
	}

	fn := func() error {
		restClient, err := restclient.RESTClientFor(p.Config)
		if err != nil {
			return err
		}

		// TODO: consider abstracting into a client invocation or client helper
		req := restClient.Post().
			Resource("pods").
			Name(pod.Name).
			Namespace(pod.Namespace).
			SubResource("exec").
			Param("container", containerName)
		req.VersionedParams(&api.PodExecOptions{
			Container: containerName,
			Command:   p.Command,
			Stdin:     p.Stdin,
			Stdout:    p.Out != nil,
			Stderr:    p.Err != nil,
			TTY:       t.Raw,
		}, legacyscheme.ParameterCodec)

		return p.Executor.Execute("POST", req.URL(), p.Config, p.In, p.Out, p.Err, t.Raw, sizeQueue)
	}

	if err := t.Safe(fn); err != nil {
		return err
	}

	return nil
}

func (p *ExecOptions) RunHyper(f util.Factory) error {
	pod, err := p.PodClient.Pods(p.Namespace).Get(p.PodName, metav1.GetOptions{})
	if err != nil {
		return err
	}

	if pod.Status.Phase == api.PodSucceeded || pod.Status.Phase == api.PodFailed {
		return fmt.Errorf("cannot exec into a container in a completed pod; current phase is %s", pod.Status.Phase)
	}

	containerName := p.ContainerName
	if len(containerName) == 0 {
		if len(pod.Spec.Containers) > 1 {
			usageString := fmt.Sprintf("Defaulting container name to %s.", pod.Spec.Containers[0].Name)
			if len(p.SuggestedCmdUsage) > 0 {
				usageString = fmt.Sprintf("%s\n%s", usageString, p.SuggestedCmdUsage)
			}
			fmt.Fprintf(p.Err, "%s\n", usageString)
		}
		containerName = pod.Spec.Containers[0].Name
	}

	if cfg, err := f.ClientConfig(); err != nil {
		return err
	} else {
		// ensure we can recover the terminal while attached
		t := p.setupTTY()

		// Set terminal emulation based on platform as required.
		stdin, stdout, stderr := dockerterm.StdStreams()
		cli, err := hyper.NewHyperCli(cfg.Host, cfg, stdin, stdout, stderr)
		if err != nil {
			return err
		}

		execConfig := &types.ExecConfig{
			Cmd:          p.Command,
			DetachKeys:   "",
			Detach:       false,
			Tty:          t.Raw,
			AttachStdin:  p.Stdin,
			AttachStdout: true,
			AttachStderr: true,
		}

		ctx := context.Background()
		response, err := cli.Client.PodExecCreate(ctx, pod.Name, containerName, *execConfig)
		if err != nil {
			return err
		}

		execID := response.ID
		if execID == "" {
			fmt.Printf("exec ID empty")
			return nil
		}

		//Temp struct for execStart so that we don't need to transfer all the execConfig
		if !execConfig.Detach {
			if err := cli.CheckTtyInput(execConfig.AttachStdin, execConfig.Tty); err != nil {
				return err
			}
		} else {
			execStartCheck := types.ExecStartCheck{
				Detach: execConfig.Detach,
				Tty:    execConfig.Tty,
			}

			if err := cli.Client.PodExecStart(ctx, execID, execStartCheck); err != nil {
				return err
			}
			// For now don't print this - wait for when we support exec wait()
			// fmt.Fprintf(cli.out, "%s\n", execID)
			return nil
		}

		glog.V(7).Infof("contID:%v execID:%v\n execConfig.DetachKeys:%v execConfig.Detach:%v execConfig.AttachStdin:%v execConfig.Tty:%v\n", containerName, execID, execConfig.DetachKeys, execConfig.Detach, execConfig.AttachStdin, execConfig.Tty)

		// Interactive exec requested.
		var (
			sOut  io.Writer
			sIn   io.ReadCloser
			sErr  io.Writer
			errCh chan error
		)

		if execConfig.AttachStdin {
			sIn = cli.In
		}
		if execConfig.AttachStdout {
			sOut = cli.Out
		}
		if execConfig.AttachStderr {
			if execConfig.Tty {
				sErr = cli.Out
			} else {
				sErr = cli.Err
			}
		}

		glog.V(7).Infof("PodExecAttach: execID:%v\n", execID)
		resp, err := cli.Client.PodExecAttach(ctx, execID, *execConfig)
		if err != nil {
			return err
		}
		defer resp.Close()
		glog.V(7).Infof("sIn:%v, execConfig.Tty:%v\n", sIn, execConfig.Tty)
		if sIn != nil && execConfig.Tty {
			if err := cli.SetRawTerminal(); err != nil {
				return err
			}
			defer cli.RestoreTerminal(sIn)
		}
		errCh = promise.Go(func() error {
			glog.V(7).Infof("HoldHijackedConnection\n")
			return cli.HoldHijackedConnection(execConfig.Tty, sIn, sOut, sErr, resp)
		})

		glog.V(7).Infof("MonitorTtySize: execConfig.Tty:%v cli.IsTerminalIn:%v\n", execConfig.Tty, cli.IsTerminalIn)
		if execConfig.Tty && cli.IsTerminalIn {
			if err := cli.MonitorTtySize(ctx, execID, true); err != nil {
				fmt.Fprintf(cli.Err, "Error monitoring TTY size: %s\n", err)
			}
		}

		if err := <-errCh; err != nil {
			glog.Errorf("Error hijack: %s", err)
			return err
		}

		var status int
		if _, status, err = cli.GetExecExitCode(ctx, execID); err != nil {
			return err
		}

		if status != 0 {
			return hyper.StatusError{StatusCode: status}
		}

		return nil
	}
}
