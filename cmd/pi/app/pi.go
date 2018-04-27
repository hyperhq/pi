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

package app

import (
	"flag"
	"os"

	"github.com/hyperhq/pi/pkg/pi/cmd"
	cmdutil "github.com/hyperhq/pi/pkg/pi/cmd/util"
	"github.com/hyperhq/pi/pkg/pi/util/logs"

	_ "github.com/hyperhq/client-go/plugin/pkg/client/auth" // pi auth providers.
	_ "k8s.io/kubernetes/pkg/client/metrics/prometheus"     // for client metric registration
	_ "k8s.io/kubernetes/pkg/version/prometheus"            // for version metric registration
)

/*
WARNING: this logic is duplicated, with minor changes, in cmd/hyperkube/pi.go
Any salient changes here will need to be manually reflected in that file.
*/

// Run runs the pi program (creates and executes a new cobra command).
func Run() error {
	logs.InitLogs()
	defer logs.FlushLogs()

	//fix: logging before flag.Parse
	flag.CommandLine.Parse([]string{})

	cmd := cmd.NewPiCommand(cmdutil.NewFactory(nil), os.Stdin, os.Stdout, os.Stderr)
	return cmd.Execute()
}
