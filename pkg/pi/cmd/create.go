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
	//"net/url"
	"os"
	//"strings"

	"github.com/hyperhq/client-go/tools/clientcmd/api/hyper"
	"github.com/hyperhq/pi/pkg/pi"
	"github.com/hyperhq/pi/pkg/pi/cmd/templates"
	cmdutil "github.com/hyperhq/pi/pkg/pi/cmd/util"
	"github.com/hyperhq/pi/pkg/pi/resource"
	"github.com/hyperhq/pi/pkg/pi/util/i18n"

	"github.com/spf13/cobra"
	"k8s.io/apimachinery/pkg/api/meta"
	"k8s.io/apimachinery/pkg/runtime/schema"
)

type CreateOptions struct {
	FilenameOptions  resource.FilenameOptions
	Selector         string
	EditBeforeCreate bool
	Raw              string
}

var (
	createLong = templates.LongDesc(i18n.T(`
		Create a resource(pod, service, secret, volume, fip).

		JSON and YAML formats are accepted(pod, service, secret).`))

	createExample = templates.Examples(i18n.T(`
		# Create a pod using the data in yaml.
		pi create -f examples/pod/pod-nginx.yaml

		# Create multiple pods using the data in yaml.
		pi create -f pod-test1.yaml -f pod-test2.yaml

		# Create a service using the data in yaml.
		pi create -f examples/service/service-nginx.yaml

		# Create a secret using the data in yaml.
		pi create -f examples/secret/secret-dockerconfigjson.yaml`))
)

func NewCmdCreate(f cmdutil.Factory, out, errOut io.Writer) *cobra.Command {
	var options CreateOptions

	cmd := &cobra.Command{
		Use:     "create -f FILENAME",
		Short:   i18n.T("Create a resource(support pod, service, secret, volume, fip)"),
		Long:    createLong,
		Example: createExample,
		Run: func(cmd *cobra.Command, args []string) {
			if cmdutil.IsFilenameSliceEmpty(options.FilenameOptions.Filenames) {
				defaultRunFunc := cmdutil.DefaultSubCommandRun(errOut)
				defaultRunFunc(cmd, args)
				return
			}
			//cmdutil.CheckErr(options.ValidateArgs(cmd, args))
			cmdutil.CheckErr(RunCreate(f, cmd, out, errOut, &options))
		},
	}

	usage := "to use to create the resource"
	cmdutil.AddFilenameOptionFlags(cmd, &options.FilenameOptions, usage)
	cmd.MarkFlagRequired("filename")
	//cmdutil.AddValidateFlags(cmd)
	//cmdutil.AddPrinterFlags(cmd)
	//cmd.Flags().BoolVar(&options.EditBeforeCreate, "edit", false, "Edit the API resource before creating")
	//cmd.Flags().Bool("windows-line-endings", runtime.GOOS == "windows",
	//	"Only relevant if --edit=true. Defaults to the line ending native to your platform.")
	//cmdutil.AddApplyAnnotationFlags(cmd)
	//cmdutil.AddRecordFlag(cmd)
	//cmdutil.AddDryRunFlag(cmd)
	//cmd.Flags().StringVarP(&options.Selector, "selector", "l", "", "Selector (label query) to filter on, supports '=', '==', and '!='.(e.g. -l key1=value1,key2=value2)")
	//cmd.Flags().StringVar(&options.Raw, "raw", options.Raw, "Raw URI to POST to the server.  Uses the transport specified by the kubeconfig file.")

	// create subcommands
	//cmd.AddCommand(NewCmdCreateSecret(f, out, errOut))
	//cmd.AddCommand(NewCmdCreateService(f, out, errOut))

	// create volume, fip
	cmd.AddCommand(NewCmdCreateVolume(f, out, errOut))
	cmd.AddCommand(NewCmdCreateFip(f, out, errOut))
	return cmd
}

func (o *CreateOptions) ValidateArgs(cmd *cobra.Command, args []string) error {
	if len(args) != 0 {
		return cmdutil.UsageErrorf(cmd, "Unexpected args: %v", args)
	}
	//if len(o.Raw) > 0 {
	//	if o.EditBeforeCreate {
	//		return cmdutil.UsageErrorf(cmd, "--raw and --edit are mutually exclusive")
	//	}
	//	if len(o.FilenameOptions.Filenames) != 1 {
	//		return cmdutil.UsageErrorf(cmd, "--raw can only use a single local file or stdin")
	//	}
	//	if strings.HasPrefix(o.FilenameOptions.Filenames[0], "http") {
	//		return cmdutil.UsageErrorf(cmd, "--raw cannot read from a url")
	//	}
	//	if o.FilenameOptions.Recursive {
	//		return cmdutil.UsageErrorf(cmd, "--raw and --recursive are mutually exclusive")
	//	}
	//	if len(o.Selector) > 0 {
	//		return cmdutil.UsageErrorf(cmd, "--raw and --selector (-l) are mutually exclusive")
	//	}
	//	if len(cmdutil.GetFlagString(cmd, "output")) > 0 {
	//		return cmdutil.UsageErrorf(cmd, "--raw and --output are mutually exclusive")
	//	}
	//	if _, err := url.ParseRequestURI(o.Raw); err != nil {
	//		return cmdutil.UsageErrorf(cmd, "--raw must be a valid URL path: %v", err)
	//	}
	//}

	return nil
}

func RunCreate(f cmdutil.Factory, cmd *cobra.Command, out, errOut io.Writer, options *CreateOptions) error {
	// raw only makes sense for a single file resource multiple objects aren't likely to do what you want.
	// the validator enforces this, so
	if len(options.Raw) > 0 {
		restClient, err := f.RESTClient()
		if err != nil {
			return err
		}

		var data io.ReadCloser
		if options.FilenameOptions.Filenames[0] == "-" {
			data = os.Stdin
		} else {
			data, err = os.Open(options.FilenameOptions.Filenames[0])
			if err != nil {
				return err
			}
		}
		// TODO post content with stream.  Right now it ignores body content
		bytes, err := restClient.Post().RequestURI(options.Raw).Body(data).DoRaw()
		if err != nil {
			return err
		}

		fmt.Fprintf(out, "%v", string(bytes))
		return nil
	}

	//schema, err := f.Validator(cmdutil.GetFlagBool(cmd, "validate"))
	//if err != nil {
	//	return err
	//}

	cmdNamespace, enforceNamespace, err := f.DefaultNamespace()
	if err != nil {
		return err
	}

	r := f.NewBuilder().
		Unstructured().
		//Schema(schema).
		ContinueOnError().
		NamespaceParam(cmdNamespace).DefaultNamespace().
		FilenameParam(enforceNamespace, &options.FilenameOptions).
		LabelSelectorParam(options.Selector).
		Flatten().
		Do()
	err = r.Err()
	if err != nil {
		return err
	}

	//dryRun := cmdutil.GetFlagBool(cmd, "dry-run")
	//output := cmdutil.GetFlagString(cmd, "output")
	dryRun := false
	output := "name"

	mapper := r.Mapper().RESTMapper

	count := 0
	err = r.Visit(func(info *resource.Info, err error) error {
		if err != nil {
			return err
		}

		//if cmdutil.ShouldRecord(cmd, info) {
		//	if err := cmdutil.RecordChangeCause(info.Object, f.Command(cmd, false)); err != nil {
		//		return cmdutil.AddSourceToErr("creating", info.Source, err)
		//	}
		//}

		if !dryRun {
			if err := createAndRefresh(info); err != nil {
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

		f.PrintSuccess(mapper, shortOutput, out, info.Mapping.Resource, info.Name, dryRun, "created")
		return nil
	})
	if err != nil {
		return err
	}
	if count == 0 {
		return fmt.Errorf("no objects passed to create")
	}
	return nil
}

// createAndRefresh creates an object from input info and refreshes info with that object
func createAndRefresh(info *resource.Info) error {
	obj, err := resource.NewHelper(info.Client, info.Mapping).Create(info.Namespace, true, info.Object)
	if err != nil {
		return err
	}
	info.Refresh(obj, true)
	return nil
}

// NameFromCommandArgs is a utility function for commands that assume the first argument is a resource name
func NameFromCommandArgs(cmd *cobra.Command, args []string) (string, error) {
	if len(args) == 0 {
		return "", cmdutil.UsageErrorf(cmd, "NAME is required")
	}
	return args[0], nil
}

// CreateSubcommandOptions is an options struct to support create subcommands
type CreateSubcommandOptions struct {
	// Name of resource being created
	Name string
	// StructuredGenerator is the resource generator for the object being created
	StructuredGenerator pi.StructuredGenerator
}

// RunCreateSubcommand executes a create subcommand using the specified options
func RunCreateSubcommand(f cmdutil.Factory, cmd *cobra.Command, out io.Writer, options *CreateSubcommandOptions) error {
	namespace, nsOverriden, err := f.DefaultNamespace()
	if err != nil {
		return err
	}
	obj, err := options.StructuredGenerator.StructuredGenerate()
	if err != nil {
		return err
	}
	mapper, typer := f.Object()
	gvks, _, err := typer.ObjectKinds(obj)
	if err != nil {
		return err
	}
	gvk := gvks[0]
	mapping, err := mapper.RESTMapping(schema.GroupKind{Group: gvk.Group, Kind: gvk.Kind}, gvk.Version)
	if err != nil {
		return err
	}
	client, err := f.ClientForMapping(mapping)
	if err != nil {
		return err
	}
	resourceMapper := &resource.Mapper{
		ObjectTyper:  typer,
		RESTMapper:   mapper,
		ClientMapper: resource.ClientMapperFunc(f.ClientForMapping),
	}
	info, err := resourceMapper.InfoForObject(obj, nil)
	if err != nil {
		return err
	}
	obj = info.Object

	dryRun := false
	outputFormat := "name"

	if !dryRun {
		obj, err = resource.NewHelper(client, mapping).Create(namespace, false, info.Object)
		if err != nil {
			return err
		}
	} else {
		if meta, err := meta.Accessor(obj); err == nil && nsOverriden {
			meta.SetNamespace(namespace)
		}
	}

	if useShortOutput := outputFormat == "name"; useShortOutput || len(outputFormat) == 0 {
		f.PrintSuccess(mapper, useShortOutput, out, mapping.Resource, info.Name, dryRun, "created")
		return nil
	}

	return f.PrintObject(cmd, false, mapper, obj, out)
}

//////////////////////////////////////////////////////
// create volume subcommand
func RunCreateVolumeSubcommand(f cmdutil.Factory, cmd *cobra.Command, out io.Writer, options *CreateSubcommandOptions) error {
	obj, err := options.StructuredGenerator.StructuredGenerate()
	if err != nil {
		return err
	}
	opts := obj.(*hyper.VolumeCreateRequest)
	if opts.Size < 1 {
		return fmt.Errorf("volume size should be >=1 (GB)")
	}
	if cfg, err := f.ClientConfig(); err != nil {
		return err
	} else {
		hyperConn := hyper.NewHyperConn(cfg)
		volCli := hyper.NewVolumeCli(hyperConn)
		if _, volCreated, err := volCli.CreateVolume(opts.Name, opts.Zone, fmt.Sprintf("%v", opts.Size)); err != nil {
			return err
		} else {
			fmt.Printf("volume/%v\n", volCreated.Name)
		}
	}
	return nil
}

// create Fip subcommand
func RunCreateFipSubcommand(f cmdutil.Factory, cmd *cobra.Command, out io.Writer, options *CreateSubcommandOptions) error {
	obj, err := options.StructuredGenerator.StructuredGenerate()
	if err != nil {
		return err
	}
	opts := obj.(*hyper.FipAllocateRequest)
	if cfg, err := f.ClientConfig(); err != nil {
		return err
	} else {
		hyperConn := hyper.NewHyperConn(cfg)
		fipCli := hyper.NewFipCli(hyperConn)
		if _, fipList, err := fipCli.AllocateFip(opts.Count); err != nil {
			return err
		} else {
			for _, fip := range fipList {
				fmt.Printf("fip/%v\n", fip.Fip)
			}
		}
	}
	return nil
}
