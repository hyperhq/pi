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

// NewCmdCreateFip groups subcommands to create various zones of fips
func NewCmdCreateFip(f cmdutil.Factory, cmdOut, errOut io.Writer) *cobra.Command {
	cmd := &cobra.Command{
		Use:     "fip [--count=int]",
		Short:   i18n.T("Create one or more fip(s) using specified subcommand"),
		Long:    fipLong,
		Aliases: []string{"fips"},
		Example: fipExample,
		Run: func(cmd *cobra.Command, args []string) {
			err := CreateFipGeneric(f, cmdOut, cmd, args)
			cmdutil.CheckErr(err)
		},
	}
	//cmdutil.AddGeneratorFlags(cmd, cmdutil.HyperFipV1GeneratorName)
	cmd.Flags().Int("count", 1, "Specify the count of fip to allocate, default is 1")
	return cmd
}

var (
	fipLong = templates.LongDesc(i18n.T(`Create one or more fip(s) using specified subcommand`))

	fipExample = templates.Examples(i18n.T(`
	  # Create one new fip
	  pi create fip

	  # Create two new fips
	  pi create fip --count=2`))
)

// CreateFipGeneric is the implementation of the create fip generic command
func CreateFipGeneric(f cmdutil.Factory, cmdOut io.Writer, cmd *cobra.Command, args []string) error {
	var generator pi.StructuredGenerator
	switch generatorName := cmdutil.HyperFipV1GeneratorName; generatorName {
	case cmdutil.HyperFipV1GeneratorName:
		generator = &pi.FipGeneratorV1{
			Count: cmdutil.GetFlagInt(cmd, "count"),
		}
	default:
		return errUnsupportedGenerator(cmd, generatorName)
	}
	return RunCreateFipSubcommand(f, cmd, cmdOut, &CreateSubcommandOptions{
		StructuredGenerator: generator,
	})
}
