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
	"os"

	"github.com/hyperhq/client-go/tools/clientcmd/api/hyper"
	"github.com/hyperhq/pi/pkg/pi/cmd/templates"
	cmdutil "github.com/hyperhq/pi/pkg/pi/cmd/util"
	"github.com/hyperhq/pi/pkg/pi/util/i18n"

	"github.com/golang/glog"
	"github.com/olekukonko/tablewriter"
	"github.com/spf13/cobra"
)

// NewCmdGetVolume groups subcommands to get various zones of volumes
func NewCmdGetVolume(f cmdutil.Factory, cmdOut, errOut io.Writer) *cobra.Command {
	cmd := &cobra.Command{
		Use:     "volumes NAME [--zone=string]",
		Short:   i18n.T("list volumes or get a volume"),
		Long:    volumeLong,
		Example: volumeExample,
		Aliases: []string{"volume"},
		Run: func(cmd *cobra.Command, args []string) {
			err := GetVolumeGeneric(f, cmdOut, cmd, args)
			cmdutil.CheckErr(err)
		},
	}
	cmdutil.AddApplyAnnotationFlags(cmd)
	cmdutil.AddValidateFlags(cmd)
	cmdutil.AddPrinterFlags(cmd)
	cmdutil.AddGeneratorFlags(cmd, cmdutil.HyperVolumeV1GeneratorName)
	cmd.Flags().String("zone", "", i18n.T("The zone of volume to get"))
	return cmd
}

var (
	volumeLong = templates.LongDesc(i18n.T(`Get a volume.`))

	volumeExample = templates.Examples(i18n.T(`
	  # Get a new volume named vol1 with default size and zone
	  pi get volume vol1

	  # Get a new volume named vol1 with specified size
	  pi get volume vol1 --size=1

	  # Get a new volume named vol1 with specified size and zone
	  pi get volume vol1 --size=1 --zone=gcp-us-central1`))
)

// GetVolumeGeneric is the implementation of the get volume generic command
func GetVolumeGeneric(f cmdutil.Factory, cmdOut io.Writer, cmd *cobra.Command, args []string) error {
	name := NameFromCommandArgs(cmd, args)
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
				PrintResult(output, volList)
			}
		} else {
			if _, vol, err := volCli.GetVolume(name, zone); err != nil {
				return err
			} else {
				PrintResult(output, []hyper.VolumeResponse{*vol})
			}
		}
	}
	return nil
}

func PrintResult(output string, result []hyper.VolumeResponse) {
	if output == "table" || output == "" {
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
		table.SetHeaderAlignment(tablewriter.ALIGN_LEFT)
		table.SetRowLine(false)
		table.SetColumnSeparator("")
		table.SetAlignment(tablewriter.ALIGN_LEFT)

		for _, vol := range data {
			table.Append(vol)
		}
		table.Render()
	} else if output == "json" {
		if buf, err := json.MarshalIndent(result, "", "  "); err != nil {
			log.Fatal(err)
		} else {
			fmt.Print(string(buf))
		}
	} else {
		glog.Warningf("--output support table,json")
	}
}

func NameFromCommandArgs(cmd *cobra.Command, args []string) string {
	if len(args) == 0 {
		return ""
	}
	return args[0]
}
