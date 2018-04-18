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
	"fmt"
	"io"
	"log"

	"github.com/hyperhq/client-go/tools/clientcmd/api/hyper"
	"github.com/hyperhq/pi/pkg/pi/cmd/templates"
	cmdutil "github.com/hyperhq/pi/pkg/pi/cmd/util"
	"github.com/hyperhq/pi/pkg/pi/util/i18n"

	"github.com/spf13/cobra"
)

// NewCmdNameFip groups subcommands to name various zones of ips
func NewCmdNameFip(f cmdutil.Factory, cmdOut, errOut io.Writer) *cobra.Command {
	cmd := &cobra.Command{
		Use:     "fip IP [--name=string]",
		Short:   i18n.T("Name a fip"),
		Long:    ipLong,
		Aliases: []string{"fips"},
		Example: ipExample,
		Run: func(cmd *cobra.Command, args []string) {
			err := NameFipGeneric(f, cmdOut, cmd, args)
			cmdutil.CheckErr(err)
		},
	}
	//cmdutil.AddGeneratorFlags(cmd, cmdutil.HyperFipV1RenameName)
	cmd.Flags().String("name", "", "Specify the name of ip to rename")
	return cmd
}

var (
	ipLong = templates.LongDesc(i18n.T(`Name a ip.`))

	ipExample = templates.Examples(i18n.T(`
	  # Name ip
	  pi name ip x.x.x.x --name=test`))
)

// GetFipGeneric is the implementation of the get fip generic command
func NameFipGeneric(f cmdutil.Factory, cmdOut io.Writer, cmd *cobra.Command, args []string) error {
	ip, err := IPFromCommandArgs(cmd, args)
	if err != nil {
		return err
	}

	name := cmdutil.GetFlagString(cmd, "name")
	if name == "" {
		return fmt.Errorf("--name can not be empty")
	}

	if cfg, err := f.ClientConfig(); err != nil {
		return err
	} else {
		hyperConn := hyper.NewHyperConn(cfg)
		fipCli := hyper.NewFipCli(hyperConn)
		if _, result, err := fipCli.NameFip(ip, name); err != nil {
			log.Fatal(result)
		} else {
			fmt.Printf("fip '%v' named to '%v'\n", ip, name)
		}
	}
	return nil
}
