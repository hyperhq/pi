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
	options := &DeleteOptions{}
	cmd := &cobra.Command{
		Use:     "volume NAME",
		Short:   i18n.T("Delete volume(s)"),
		Aliases: []string{"volumes"},
		Long:    delVolumeLong,
		Example: delVolumeExample,
		Run: func(cmd *cobra.Command, args []string) {
			err := options.DeleteVolumeGeneric(f, cmdOut, cmd, args)
			cmdutil.CheckErr(err)
		},
	}
	cmd.Flags().BoolVar(&options.DeleteAll, "all", false, "Delete all volumes")
	return cmd
}

var (
	delVolumeLong = templates.LongDesc(i18n.T(`Delete volume(s).`))

	delVolumeExample = templates.Examples(i18n.T(`
	  # Delete a volume named vol1
	  pi delete volumes vol1

	  # Delete multiple volumes
	  pi delete volumes vol1 vol2`))
)

// DeleteVolumeGeneric is the implementation of the delete volume generic command
func (o *DeleteOptions) DeleteVolumeGeneric(f cmdutil.Factory, cmdOut io.Writer, cmd *cobra.Command, args []string) error {
	if len(args) != 0 && o.DeleteAll {
		return fmt.Errorf("name cannot be provided when --all is specified")
	}
	if len(args) == 0 && !o.DeleteAll {
		return fmt.Errorf("resource(s) were provided, but no name or --all flag specified")
	}

	if cfg, err := f.ClientConfig(); err != nil {
		return err
	} else {
		hyperConn := hyper.NewHyperConn(cfg)
		volCli := hyper.NewVolumeCli(hyperConn)

		if o.DeleteAll {
			_, volList, err := volCli.ListVolumes("")
			if err != nil {
				return fmt.Errorf("failed to list all fips, error:%v", err)
			}
			for _, vol := range volList {
				args = append(args, vol.Name)
			}
		}

		for _, name := range args {
			httpStatus, result := volCli.DeleteVolume(name, "")
			if httpStatus == http.StatusNoContent {
				fmt.Printf("volume \"%v\" deleted\n", name)
			} else {
				fmt.Println(result)
			}
		}
	}
	return nil
}
