package resource

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

import (
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"
	"os"

	"github.com/hyperhq/client-go/tools/clientcmd/api/hyper"
	"github.com/hyperhq/pi/pkg/pi/cmd/templates"
	cmdutil "github.com/hyperhq/pi/pkg/pi/cmd/util"
	"github.com/hyperhq/pi/pkg/pi/util/i18n"

	"github.com/olekukonko/tablewriter"
	"github.com/spf13/cobra"
)

// NewCmdGetVolume groups subcommands to get various zones of volumes
func NewCmdGetVolume(f cmdutil.Factory, cmdOut, errOut io.Writer) *cobra.Command {
	cmd := &cobra.Command{
		Use:     "volume NAME [--zone=string]",
		Short:   i18n.T("list volumes or get a volume"),
		Long:    volumeLong,
		Example: volumeExample,
		Aliases: []string{"volumes"},
		Run: func(cmd *cobra.Command, args []string) {
			err := GetVolumeGeneric(f, cmdOut, cmd, args)
			cmdutil.CheckErr(err)
		},
	}
	cmd.Flags().StringP("output", "o", "", "Output format. One of: json|name")
	cmd.Flags().String("zone", "", i18n.T("The zone of volume to get"))
	return cmd
}

var (
	volumeLong = templates.LongDesc(i18n.T(`List volumes or get a volume.`))

	volumeExample = templates.Examples(i18n.T(`
	  # List volumes
	  pi get volumes	  

	  # Get a volume named vol1 with default size and zone
	  pi get volume vol1`))
)

// GetVolumeGeneric is the implementation of the get volume generic command
func GetVolumeGeneric(f cmdutil.Factory, cmdOut io.Writer, cmd *cobra.Command, args []string) error {
	name := VolNameFromCommandArgs(cmd, args)
	zone := cmdutil.GetFlagString(cmd, "zone")
	output := cmdutil.GetFlagString(cmd, "output")

	if cfg, err := f.ClientConfig(); err != nil {
		return err
	} else {
		hyperConn := hyper.NewHyperConn(cfg)
		volCli := hyper.NewVolumeCli(hyperConn)
		if name == "" {
			if _, volList, err := volCli.ListVolumes(zone); err != nil {
				return err
			} else {
				if len(volList) == 0 {
					fmt.Println("No resources found.")
				} else {
					return PrintVolumeResult(output, volList)
				}
			}
		} else {
			if httpStatus, vol, err := volCli.GetVolume(name, zone); err != nil {
				return err
			} else {
				if httpStatus == http.StatusNotFound {
					fmt.Println("No resources found.")
				} else {
					return PrintVolumeResult(output, []hyper.VolumeResponse{*vol})
				}
			}
		}
	}
	return nil
}

func PrintVolumeResult(output string, result []hyper.VolumeResponse) error {
	if output == "" {
		data := [][]string{}
		for _, vol := range result {
			item := []string{vol.Name, vol.Zone, fmt.Sprint(vol.Size), vol.CreatedAt.Format("2006-01-02T15:04:05-07:00"), vol.Pod}
			data = append(data, item)
		}
		table := tablewriter.NewWriter(os.Stdout)
		table.SetHeader([]string{"Name", "Zone", "Size(GB)", "CreatedAt", "Pod"})

		//set table style
		table.SetBorder(false)
		table.SetHeaderLine(false)
		table.SetRowLine(false)
		table.SetColumnSeparator("")
		table.SetHeaderAlignment(tablewriter.ALIGN_LEFT)
		table.SetAlignment(tablewriter.ALIGN_LEFT)
		table.SetColumnAlignment([]int{tablewriter.ALIGN_LEFT})

		for _, vol := range data {
			table.Append(vol)
		}
		table.Render()
	} else if output == "json" {
		if buf, err := json.MarshalIndent(result, "", "  "); err != nil {
			log.Fatal(err)
		} else {
			fmt.Println(string(buf))
		}
	} else if output == "name" {
		for _, vol := range result {
			fmt.Println(vol.Name)
		}
	} else {
		err := fmt.Errorf("error: output format \"%v\" not recognized", output)
		return err
	}
	return nil
}

func VolNameFromCommandArgs(cmd *cobra.Command, args []string) string {
	if len(args) == 0 {
		return ""
	}
	return args[0]
}
