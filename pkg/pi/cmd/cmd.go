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

	"github.com/hyperhq/client-go/tools/clientcmd"
	cmdconfig "github.com/hyperhq/pi/pkg/pi/cmd/config"
	"github.com/hyperhq/pi/pkg/pi/cmd/resource"
	"github.com/hyperhq/pi/pkg/pi/cmd/templates"
	cmdutil "github.com/hyperhq/pi/pkg/pi/cmd/util"
	"github.com/hyperhq/pi/pkg/pi/util/i18n"
	"k8s.io/apiserver/pkg/util/flag"

	"github.com/golang/glog"
	"github.com/spf13/cobra"
)

const (
	bashCompletionFunc = `# call pi get $1,
__pi_override_flag_list=(--user --context --server -n -s)
__pi_override_flags()
{
    local ${__pi_override_flag_list[*]##*-} two_word_of of var
    for w in "${words[@]}"; do
        if [ -n "${two_word_of}" ]; then
            eval "${two_word_of##*-}=\"${two_word_of}=\${w}\""
            two_word_of=
            continue
        fi
        for of in "${__pi_override_flag_list[@]}"; do
            case "${w}" in
                ${of}=*)
                    eval "${of##*-}=\"${w}\""
                    ;;
                ${of})
                    two_word_of="${of}"
                    ;;
            esac
        done
    done
    for var in "${__pi_override_flag_list[@]##*-}"; do
        if eval "test -n \"\$${var}\""; then
            eval "echo \${${var}}"
        fi
    done
}

__pi_config_get_contexts()
{
    __pi_parse_config "contexts"
}

__pi_config_get_clusters()
{
    __pi_parse_config "clusters"
}

__pi_config_get_users()
{
    __pi_parse_config "users"
}

# $1 has to be "contexts", "clusters" or "users"
__pi_parse_config()
{
    local template pi_out
    template="{{ range .$1  }}{{ .name }} {{ end }}"
    if pi_out=$(pi config $(__pi_override_flags) -o template --template="${template}" view 2>/dev/null); then
        COMPREPLY=( $( compgen -W "${pi_out[*]}" -- "$cur" ) )
    fi
}

__pi_parse_get()
{
    local template
    template="{{ range .items  }}{{ .metadata.name }} {{ end }}"
    local pi_out
    if pi_out=$(pi get $(__pi_override_flags) -o template --template="${template}" "$1" 2>/dev/null); then
        COMPREPLY=( $( compgen -W "${pi_out[*]}" -- "$cur" ) )
    fi
}

__pi_get_resource()
{
    if [[ ${#nouns[@]} -eq 0 ]]; then
        return 1
    fi
    __pi_parse_get "${nouns[${#nouns[@]} -1]}"
}

__pi_get_resource_namespace()
{
    __pi_parse_get "namespace"
}

__pi_get_resource_pod()
{
    __pi_parse_get "pod"
}

__pi_get_resource_rc()
{
    __pi_parse_get "rc"
}

__pi_get_resource_node()
{
    __pi_parse_get "node"
}

__pi_get_resource_clusterrole()
{
    __pi_parse_get "clusterrole"
}

# $1 is the name of the pod we want to get the list of containers inside
__pi_get_containers()
{
    local template
    template="{{ range .spec.containers  }}{{ .name }} {{ end }}"
    __debug "${FUNCNAME} nouns are ${nouns[*]}"

    local len="${#nouns[@]}"
    if [[ ${len} -ne 1 ]]; then
        return
    fi
    local last=${nouns[${len} -1]}
    local pi_out
    if pi_out=$(pi get $(__pi_override_flags) -o template --template="${template}" pods "${last}" 2>/dev/null); then
        COMPREPLY=( $( compgen -W "${pi_out[*]}" -- "$cur" ) )
    fi
}

# Require both a pod and a container to be specified
__pi_require_pod_and_container()
{
    if [[ ${#nouns[@]} -eq 0 ]]; then
        __pi_parse_get pods
        return 0
    fi;
    __pi_get_containers
    return 0
}

__custom_func() {
    case ${last_command} in
        pi_get | pi_describe | pi_delete)
            __pi_get_resource
            return
            ;;
        pi_exec)
            __pi_get_resource_pod
            return
            ;;
        *)
            ;;
    esac
}
`
)

var (
	bash_completion_flags = map[string]string{
	//"namespace": "__pi_get_resource_namespace",
	//"context":   "__pi_config_get_contexts",
	//"cluster":   "__pi_config_get_clusters",
	//"user":      "__pi_config_get_users",
	}
)

// NewPiCommand creates the `pi` command and its nested children.
func NewPiCommand(f cmdutil.Factory, in io.Reader, out, err io.Writer) *cobra.Command {
	// Parent command to which all subcommands are added.
	cmds := &cobra.Command{
		Use:   "pi",
		Short: i18n.T("pi controls the resources on Hyper GCP cluster"),
		Long: templates.LongDesc(`
      pi controls the resources on Hyper GCP cluster.

      Find more information at https://github.com/hyperhq/pi.`),
		Run: runHelp,
		BashCompletionFunction: bashCompletionFunc,
	}

	f.BindFlags(cmds.PersistentFlags())
	f.BindExternalFlags(cmds.PersistentFlags())

	// Sending in 'nil' for the getLanguageFn() results in using
	// the LANG environment variable.
	//
	// TODO: Consider adding a flag or file preference for setting
	// the language, instead of just loading from the LANG env. variable.
	i18n.LoadTranslations("pi", nil)

	// From this point and forward we get warnings on flags that contain "_" separators
	cmds.SetGlobalNormalizationFunc(flag.WarnWordSepNormalizeFunc)

	groups := templates.CommandGroups{
		{
			Message: "Basic Commands (Beginner):",
			Commands: []*cobra.Command{
				NewCmdCreate(f, out, err),
			},
		},
		{
			Message: "Basic Commands (Intermediate):",
			Commands: []*cobra.Command{
				resource.NewCmdGet(f, out, err),
				NewCmdDelete(f, out, err),
				NewCmdName(f, out, err),
			},
		},
		{
			Message: "Troubleshooting and Debugging Commands:",
			Commands: []*cobra.Command{
				//NewCmdDescribe(f, out, err),
				NewCmdExec(f, in, out, err),
			},
		},
	}
	groups.Add(cmds)

	filters := []string{"options"}

	templates.ActsAsRootCommand(cmds, filters, groups...)

	for name, completion := range bash_completion_flags {
		if cmds.Flag(name) != nil {
			if cmds.Flag(name).Annotations == nil {
				cmds.Flag(name).Annotations = map[string][]string{}
			}
			cmds.Flag(name).Annotations[cobra.BashCompCustom] = append(
				cmds.Flag(name).Annotations[cobra.BashCompCustom],
				completion,
			)
		}
	}

	cmds.AddCommand(NewCmdOptions(out))
	cmds.AddCommand(cmdconfig.NewCmdConfig(clientcmd.NewDefaultPathOptions(), out, err))
	return cmds
}

func runHelp(cmd *cobra.Command, args []string) {
	cmd.Help()
}

func printDeprecationWarning(command, alias string) {
	glog.Warningf("%s is DEPRECATED and will be removed in a future version. Use %s instead.", alias, command)
}

// deprecatedAlias is intended to be used to create a "wrapper" command around
// an existing command. The wrapper works the same but prints a deprecation
// message before running. This command is identical functionality.
func deprecatedAlias(deprecatedVersion string, cmd *cobra.Command) *cobra.Command {
	// Have to be careful here because Cobra automatically extracts the name
	// of the command from the .Use field.
	originalName := cmd.Name()

	cmd.Use = deprecatedVersion
	cmd.Deprecated = fmt.Sprintf("use %q instead", originalName)
	cmd.Short = fmt.Sprintf("%s. This command is deprecated, use %q instead", cmd.Short, originalName)
	cmd.Hidden = true
	return cmd
}

// deprecated is similar to deprecatedAlias, but it is used for deprecations
// that are not simple aliases; this command is actually a different
// (deprecated) codepath.
func deprecated(baseName, to string, parent, cmd *cobra.Command) string {
	cmd.Long = fmt.Sprintf("Deprecated: all functionality can be found in \"%s %s\"", baseName, to)
	cmd.Short = fmt.Sprintf("Deprecated: use %s", to)
	parent.AddCommand(cmd)
	return cmd.Name()
}
