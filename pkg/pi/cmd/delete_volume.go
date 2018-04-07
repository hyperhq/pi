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

// NewCmdDeleteVolume groups subcommands to delete various zones of volumes
func NewCmdDeleteVolume(f cmdutil.Factory, cmdOut, errOut io.Writer) *cobra.Command {
	cmd := &cobra.Command{
		Use:     "volume NAME [--zone=string] [--size=string] [--dry-run]",
		Short:   i18n.T("Delete a volume"),
		Aliases: []string{"volumes"},
		Long:    delVolumeLong,
		Example: delVolumeExample,
		Run: func(cmd *cobra.Command, args []string) {
			err := DeleteVolumeGeneric(f, cmdOut, cmd, args)
			cmdutil.CheckErr(err)
		},
	}
	cmdutil.AddApplyAnnotationFlags(cmd)
	cmdutil.AddValidateFlags(cmd)
	cmdutil.AddPrinterFlags(cmd)
	cmdutil.AddGeneratorFlags(cmd, cmdutil.HyperVolumeV1GeneratorName)

	return cmd
}

var (
	delVolumeLong = templates.LongDesc(i18n.T(`Delete a volume.`))

	delVolumeExample = templates.Examples(i18n.T(`
	  # Delete a new volume named vol1 with default size and zone
	  pi delete volume vol1

	  # Delete a new volume named vol1 with specified size
	  pi delete volume vol1 --size=1

	  # Delete a new volume named vol1 with specified size and zone
	  pi delete volume vol1 --size=1 --zone=gcp-us-central1`))
)

// DeleteVolumeGeneric is the implementation of the delete volume generic command
func DeleteVolumeGeneric(f cmdutil.Factory, cmdOut io.Writer, cmd *cobra.Command, args []string) error {
	name, err := NameFromCommandArgs(cmd, args)
	if err != nil {
		return err
	}

	if cfg, err := f.ClientConfig(); err != nil {
		return err
	} else {
		hyperConn := hyper.NewHyperConn(cfg)
		volCli := hyper.NewVolumeCli(hyperConn)
		httpStatus, result := volCli.DeleteVolume(name, "")
		if httpStatus == http.StatusNoContent {
			fmt.Printf("volume %v deleted\n", name)
		} else {
			fmt.Println(result)
		}
	}
	return nil
}
