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
	"strings"

	"github.com/hyperhq/pi/pkg/pi"
	"github.com/hyperhq/pi/pkg/pi/cmd/templates"
	cmdutil "github.com/hyperhq/pi/pkg/pi/cmd/util"
	"github.com/hyperhq/pi/pkg/pi/resource"
	"github.com/hyperhq/pi/pkg/pi/util/i18n"
	"github.com/hyperhq/pi/pkg/printers"
	printersinternal "github.com/hyperhq/pi/pkg/printers/internalversion"

	"github.com/spf13/cobra"
	apierrors "k8s.io/apimachinery/pkg/api/errors"
	utilerrors "k8s.io/apimachinery/pkg/util/errors"
	"k8s.io/apimachinery/pkg/util/sets"
)

var (
	describeLong = templates.LongDesc(`
		Show details of a specific resource or group of resources

		Print a detailed description of the selected resources, including related resources such
		as events or controllers. You may select a single object by name, all objects of that 
		type, provide a name prefix, or label selector. For example:

		    $ pi describe TYPE NAME_PREFIX

		will first check for an exact match on TYPE and NAME_PREFIX. If no such resource
		exists, it will output details for every resource that has a name prefixed with NAME_PREFIX.`)

	describeExample = templates.Examples(i18n.T(`
		# Describe a pod
		pi describe pods/nginx

		# Describe all pods
		pi describe pods

		# Describe pods by label name=myLabel
		pi describe po -l name=myLabel

		# Describe a service
		pi describe service my-service

		# Describe a secret
		pi describe secret my-secret`))
)

func NewCmdDescribe(f cmdutil.Factory, out, cmdErr io.Writer) *cobra.Command {
	options := &resource.FilenameOptions{}
	describerSettings := &printers.DescriberSettings{}

	// TODO: this should come from the factory, and may need to be loaded from the server, and so is probably
	//   going to have to be removed
	validArgs := printersinternal.DescribableResources()
	argAliases := pi.ResourceAliases(validArgs)

	cmd := &cobra.Command{
		Use:     "describe (TYPE [NAME_PREFIX | -l label] | TYPE/NAME)",
		Short:   i18n.T("Show details of a specific resource or group of resources"),
		Long:    describeLong + "\n\n" + cmdutil.ValidDescribeResourceTypeList(f) + "\n",
		Example: describeExample,
		Run: func(cmd *cobra.Command, args []string) {
			err := RunDescribe(f, out, cmdErr, cmd, args, options, describerSettings)
			cmdutil.CheckErr(err)
		},
		ValidArgs:  validArgs,
		ArgAliases: argAliases,
	}
	//usage := "containing the resource to describe"
	//cmdutil.AddFilenameOptionFlags(cmd, options, usage)
	cmd.Flags().StringP("selector", "l", "", "Selector (label query) to filter on, supports '=', '==', and '!='.(e.g. -l key1=value1,key2=value2)")
	//cmd.Flags().Bool("all-namespaces", false, "If present, list the requested object(s) across all namespaces. Namespace in current context is ignored even if specified with --namespace.")
	cmd.Flags().BoolVar(&describerSettings.ShowEvents, "show-events", true, "If true, display events related to the described object.")
	cmdutil.AddIncludeUninitializedFlag(cmd)
	return cmd
}

func RunDescribe(f cmdutil.Factory, out, cmdErr io.Writer, cmd *cobra.Command, args []string, options *resource.FilenameOptions, describerSettings *printers.DescriberSettings) error {
	selector := cmdutil.GetFlagString(cmd, "selector")
	//allNamespaces := cmdutil.GetFlagBool(cmd, "all-namespaces")
	cmdNamespace, _, err := f.DefaultNamespace()
	if err != nil {
		return err
	}
	//if allNamespaces {
	//	enforceNamespace = false
	//}
	if len(args) == 0 && cmdutil.IsFilenameSliceEmpty(options.Filenames) {
		fmt.Fprint(cmdErr, "You must specify the type of resource to describe. ", cmdutil.ValidDescribeResourceTypeList(f), "\n")
		return cmdutil.UsageErrorf(cmd, "Required resource not specified.")
	}

	// include the uninitialized objects by default
	// unless user explicitly set --include-uninitialized=false
	includeUninitialized := cmdutil.ShouldIncludeUninitialized(cmd, true)
	r := f.NewBuilder().
		Unstructured().
		ContinueOnError().
		NamespaceParam(cmdNamespace).DefaultNamespace(). //AllNamespaces(allNamespaces).
		//FilenameParam(enforceNamespace, options).
		LabelSelectorParam(selector).
		IncludeUninitialized(includeUninitialized).
		ResourceTypeOrNameArgs(true, args...).
		Flatten().
		Do()
	err = r.Err()
	if err != nil {
		return err
	}

	allErrs := []error{}
	infos, err := r.Infos()
	if err != nil {
		if apierrors.IsNotFound(err) && len(args) == 2 {
			return DescribeMatchingResources(f, cmdNamespace, args[0], args[1], describerSettings, out, err)
		}
		allErrs = append(allErrs, err)
	}

	errs := sets.NewString()
	first := true
	for _, info := range infos {
		mapping := info.ResourceMapping()
		describer, err := f.Describer(mapping)
		if err != nil {
			if errs.Has(err.Error()) {
				continue
			}
			allErrs = append(allErrs, err)
			errs.Insert(err.Error())
			continue
		}
		s, err := describer.Describe(info.Namespace, info.Name, *describerSettings)
		if err != nil {
			if errs.Has(err.Error()) {
				continue
			}
			allErrs = append(allErrs, err)
			errs.Insert(err.Error())
			continue
		}
		if first {
			first = false
			fmt.Fprint(out, s)
		} else {
			fmt.Fprintf(out, "\n\n%s", s)
		}
	}

	return utilerrors.NewAggregate(allErrs)
}

func DescribeMatchingResources(f cmdutil.Factory, namespace, rsrc, prefix string, describerSettings *printers.DescriberSettings, out io.Writer, originalError error) error {
	r := f.NewBuilder().
		Unstructured().
		NamespaceParam(namespace).DefaultNamespace().
		ResourceTypeOrNameArgs(true, rsrc).
		SingleResourceType().
		Flatten().
		Do()
	mapping, err := r.ResourceMapping()
	if err != nil {
		return err
	}
	describer, err := f.Describer(mapping)
	if err != nil {
		return err
	}
	infos, err := r.Infos()
	if err != nil {
		return err
	}
	isFound := false
	for ix := range infos {
		info := infos[ix]
		if strings.HasPrefix(info.Name, prefix) {
			isFound = true
			s, err := describer.Describe(info.Namespace, info.Name, *describerSettings)
			if err != nil {
				return err
			}
			fmt.Fprintf(out, "%s\n", s)
		}
	}
	if !isFound {
		return originalError
	}
	return nil
}
