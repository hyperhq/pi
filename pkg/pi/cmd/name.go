/*
Copyright 2014 The Kubernetes Authors.

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

	"github.com/hyperhq/pi/pkg/pi"
	"github.com/hyperhq/pi/pkg/pi/cmd/templates"
	cmdutil "github.com/hyperhq/pi/pkg/pi/cmd/util"
	"github.com/hyperhq/pi/pkg/pi/resource"
	"github.com/hyperhq/pi/pkg/pi/util/i18n"

	"github.com/spf13/cobra"
)

type NameOptions struct {
}

var (
	nameLong = templates.LongDesc(i18n.T(`
		Name a resource.`))

	nameExample = templates.Examples(i18n.T(`
		# Name a resource.
		pi name fip x.x.x.x --name=test`))
)

func NewCmdName(f cmdutil.Factory, out, errOut io.Writer) *cobra.Command {
	var options NameOptions

	cmd := &cobra.Command{
		Use:     "name",
		Short:   i18n.T("Name a resource"),
		Long:    nameLong,
		Example: nameExample,
		Run: func(cmd *cobra.Command, args []string) {
			if err := options.Complete(f, out, errOut, args, cmd); err != nil {
				cmdutil.CheckErr(err)
			}
			cmdutil.CheckErr(RunName(f, cmd, out, errOut, &options))
		},
	}

	// name fip
	cmd.AddCommand(NewCmdNameFip(f, out, errOut))
	return cmd
}

func (o *NameOptions) Complete(f cmdutil.Factory, out, errOut io.Writer, args []string, cmd *cobra.Command) error {
	if len(args) == 0 {
		usageString := fmt.Sprint("You must specify the type of resource to rename. ",
			cmdutil.ValidNameResourceTypeList(f), "\nerror: Required resource not specified.")
		return cmdutil.UsageErrorf(cmd, usageString)
	}
	return nil
}

func RunName(f cmdutil.Factory, cmd *cobra.Command, out, errOut io.Writer, options *NameOptions) error {
	// raw only makes sense for a single file resource multiple objects aren't likely to do what you want.
	// the validator enforces this, so

	cmdNamespace, _, err := f.DefaultNamespace()
	if err != nil {
		return err
	}

	r := f.NewBuilder().
		Unstructured().
		ContinueOnError().
		NamespaceParam(cmdNamespace).DefaultNamespace().
		Flatten().
		Do()
	err = r.Err()
	if err != nil {
		return err
	}

	dryRun := false
	output := "name"

	mapper := r.Mapper().RESTMapper

	count := 0
	err = r.Visit(func(info *resource.Info, err error) error {
		if err != nil {
			return err
		}

		if cmdutil.ShouldRecord(cmd, info) {
			if err := cmdutil.RecordChangeCause(info.Object, f.Command(cmd, false)); err != nil {
				return cmdutil.AddSourceToErr("creating", info.Source, err)
			}
		}

		count++

		shortOutput := output == "name"
		if len(output) > 0 && !shortOutput {
			return f.PrintResourceInfoForCommand(cmd, info, out)
		}
		if !shortOutput {
			f.PrintObjectSpecificMessage(info.Object, out)
		}

		f.PrintSuccess(mapper, shortOutput, out, info.Mapping.Resource, info.Name, dryRun, "named")
		return nil
	})
	if err != nil {
		return err
	}
	if count == 0 {
		return fmt.Errorf("no objects passed to name")
	}
	return nil
}

// NameSubcommandOptions is an options struct to support name subcommands
type NameSubcommandOptions struct {
	// Name of resource being named
	Name string
	// StructuredGenerator is the resource generator for the object being named
	StructuredGenerator pi.StructuredGenerator
	// DryRun is true if the command should be simulated but not run against the server
	DryRun       bool
	OutputFormat string
}
