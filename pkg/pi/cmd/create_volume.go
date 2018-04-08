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

	"github.com/hyperhq/pi/pkg/pi"
	"github.com/hyperhq/pi/pkg/pi/cmd/templates"
	cmdutil "github.com/hyperhq/pi/pkg/pi/cmd/util"
	"github.com/hyperhq/pi/pkg/pi/util/i18n"
)

// NewCmdCreateVolume groups subcommands to create various zones of volumes
func NewCmdCreateVolume(f cmdutil.Factory, cmdOut, errOut io.Writer) *cobra.Command {
	cmd := &cobra.Command{
		Use:     "volume NAME [--zone=string] [--size=int]",
		Short:   i18n.T("Create a volume"),
		Long:    volumeLong,
		Example: volumeExample,
		Run: func(cmd *cobra.Command, args []string) {
			err := CreateVolumeGeneric(f, cmdOut, cmd, args)
			cmdutil.CheckErr(err)
		},
	}
	cmdutil.AddApplyAnnotationFlags(cmd)
	cmdutil.AddValidateFlags(cmd)
	cmdutil.AddPrinterFlags(cmd)
	cmdutil.AddGeneratorFlags(cmd, cmdutil.HyperVolumeV1GeneratorName)
	cmd.Flags().String("size", "", "Specify the volume size, default 10(GB), min 1, max 1024")
	cmd.Flags().String("zone", "", i18n.T("The zone of volume to create"))
	return cmd
}

var (
	volumeLong = templates.LongDesc(i18n.T(`Create a volume.`))

	volumeExample = templates.Examples(i18n.T(`
	  # Create a new volume named vol1 with default size and zone
	  pi create volume vol1

	  # Create a new volume named vol1 with specified size
	  pi create volume vol1 --size=1

	  # Create a new volume named vol1 with specified size and zone
	  pi create volume vol1 --size=1 --zone=gcp-us-central1`))
)

// CreateVolumeGeneric is the implementation of the create volume generic command
func CreateVolumeGeneric(f cmdutil.Factory, cmdOut io.Writer, cmd *cobra.Command, args []string) error {
	name, err := NameFromCommandArgs(cmd, args)
	if err != nil {
		return err
	}
	var generator pi.StructuredGenerator
	switch generatorName := cmdutil.GetFlagString(cmd, "generator"); generatorName {
	case cmdutil.HyperVolumeV1GeneratorName:
		generator = &pi.VolumeGeneratorV1{
			Name: name,
			Size: cmdutil.GetFlagString(cmd, "size"),
			Zone: cmdutil.GetFlagString(cmd, "zone"),
		}
	default:
		return errUnsupportedGenerator(cmd, generatorName)
	}
	return RunCreateVolumeSubcommand(f, cmd, cmdOut, &CreateSubcommandOptions{
		Name:                name,
		StructuredGenerator: generator,
		DryRun:              cmdutil.GetDryRunFlag(cmd),
		OutputFormat:        cmdutil.GetFlagString(cmd, "output"),
	})
}
