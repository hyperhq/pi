/*
Copyright 2015 The Kubernetes Authors.

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
	"io"

	"github.com/spf13/cobra"

	"github.com/hyperhq/pi/pkg/pi/cmd/templates"
	cmdutil "github.com/hyperhq/pi/pkg/pi/cmd/util"
	"github.com/hyperhq/pi/pkg/pi/util/i18n"
)

var (
	createPodLong = templates.LongDesc(i18n.T(`Create and run a pod with particular image.`))

	createPodExample = templates.Examples(i18n.T(`
		# Start a single instance of nginx.
		pi create pod nginx --image=nginx

		# Start a single instance of nginx and set environment variables "DNS_DOMAIN=cluster" and "POD_NAMESPACE=default" in the container.
		pi create pod nginx --image=nginx --env="DNS_DOMAIN=cluster" --env="POD_NAMESPACE=default"

		# Start a single instance of nginx and set labels "app=nginx" and "env=prod" in the container.
		pi create pod nginx --image=nginx --labels="app=nginx,env=prod"

		# Start a pod of busybox and keep it in the foreground, don't restart it if it exits.
		pi create pod -it busybox --image=busybox --restart=Never -- sh

		# Start the nginx container using a specified command and custom arguments.
		pi create pod nginx --image=nginx -- <cmd> <arg1> ... <argN>

		# Start the nginx container using a specified command and custom arguments.
		pi create pod nginx --rm --image=nginx -- echo hello world`))
)

// NewCmdCreatePod groups subcommands to create pod
func NewCmdCreatePod(f cmdutil.Factory, cmdIn io.Reader, cmdOut, errOut io.Writer) *cobra.Command {
	cmd := &cobra.Command{
		Use:     "pod NAME --image=image [options] [COMMAND]",
		Short:   i18n.T("Create and run a pod with particular image."),
		Long:    createPodLong,
		Example: createPodExample,
		Run: func(cmd *cobra.Command, args []string) {
			argsLenAtDash := cmd.ArgsLenAtDash()
			err := RunRun(f, cmdIn, cmdOut, errOut, cmd, args, argsLenAtDash)
			// RunRun(f, cmdOut, cmdErr, cmd, args, argsLenAtDash)
			cmdutil.CheckErr(err)
		},
	}

	AddRunFlags(cmd)

	return cmd
}
