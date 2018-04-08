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

	"github.com/hyperhq/client-go/tools/clientcmd/api/hyper"
	"github.com/hyperhq/pi/pkg/pi/cmd/templates"
	cmdutil "github.com/hyperhq/pi/pkg/pi/cmd/util"
	"github.com/hyperhq/pi/pkg/pi/util/i18n"

	"github.com/spf13/cobra"
	"net/http"
)

// NewCmdDeleteFip groups subcommands to delete various zones of fips
func NewCmdDeleteFip(f cmdutil.Factory, cmdOut, errOut io.Writer) *cobra.Command {
	cmd := &cobra.Command{
		Use:     "fip IP",
		Short:   i18n.T("Delete a fip"),
		Aliases: []string{"fips"},
		Long:    delFipLong,
		Example: delFipExample,
		Run: func(cmd *cobra.Command, args []string) {
			err := DeleteFipGeneric(f, cmdOut, cmd, args)
			cmdutil.CheckErr(err)
		},
	}
	cmdutil.AddApplyAnnotationFlags(cmd)
	cmdutil.AddValidateFlags(cmd)
	cmdutil.AddPrinterFlags(cmd)
	cmdutil.AddGeneratorFlags(cmd, cmdutil.HyperFipV1GeneratorName)

	return cmd
}

var (
	delFipLong = templates.LongDesc(i18n.T(`Delete a fip.`))

	delFipExample = templates.Examples(i18n.T(`
	  # Delete a fip
	  pi delete fip x.x.x.x`))
)

// DeleteFipGeneric is the implementation of the delete fip generic command
func DeleteFipGeneric(f cmdutil.Factory, cmdOut io.Writer, cmd *cobra.Command, args []string) error {
	ip, err := IPFromCommandArgs(cmd, args)
	if err != nil {
		return err
	}

	if cfg, err := f.ClientConfig(); err != nil {
		return err
	} else {
		hyperConn := hyper.NewHyperConn(cfg)
		fipCli := hyper.NewFipCli(hyperConn)
		httpStatus, result := fipCli.ReleaseFip(ip)
		if httpStatus == http.StatusNoContent {
			fmt.Printf("fip \"%v\" deleted\n", ip)
		} else {
			fmt.Println(result)
		}
	}
	return nil
}

func IPFromCommandArgs(cmd *cobra.Command, args []string) (string, error) {
	if len(args) == 0 {
		return "", cmdutil.UsageErrorf(cmd, "IP is required")
	}
	return args[0], nil
}
