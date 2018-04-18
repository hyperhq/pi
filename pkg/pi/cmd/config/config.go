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

package config

import (
	"fmt"
	"io"
	"path"
	"strconv"

	"github.com/spf13/cobra"

	"github.com/hyperhq/client-go/tools/clientcmd"
	"github.com/hyperhq/pi/pkg/pi/cmd/templates"
	cmdutil "github.com/hyperhq/pi/pkg/pi/cmd/util"
	"github.com/hyperhq/pi/pkg/pi/util/i18n"
)

// NewCmdConfig creates a command object for the "config" action, and adds all child commands to it.
func NewCmdConfig(pathOptions *clientcmd.PathOptions, out, errOut io.Writer) *cobra.Command {
	if len(pathOptions.ExplicitFileFlag) == 0 {
		pathOptions.ExplicitFileFlag = clientcmd.RecommendedConfigPathFlag
	}

	cmd := &cobra.Command{
		Use:   "config SUBCOMMAND",
		Short: i18n.T("Modify pi config file"),
		Long:  templates.LongDesc(`Modify pi config file ` + path.Join("${HOME}", pathOptions.GlobalFileSubpath)),
		Example: templates.Examples(`
			# Set credential for user (default region is gcp-us-central1)  
			pi config set-credentials user1 --access-key="xxx" --secret-key="xxxxxx"

			# Set credential for user with region
			pi config set-credentials user1 --region=gcp-us-central1 --access-key="xxx" --secret-key="xxxxxx"

			# Print credentials for current user
			pi config view --minify=true

			# Switch current credential of specified user
			pi config set-context default --user=user1

			# Delete specified credentials
			pi config delete-credentials user1`),
		Run: cmdutil.DefaultSubCommandRun(errOut),
	}

	// file paths are common to all sub commands
	cmd.PersistentFlags().StringVar(&pathOptions.LoadingRules.ExplicitPath, pathOptions.ExplicitFileFlag, pathOptions.LoadingRules.ExplicitPath, "use a particular pi config file")

	cmd.AddCommand(NewCmdConfigView(out, errOut, pathOptions))
	//cmd.AddCommand(NewCmdConfigSetCluster(out, pathOptions))
	cmd.AddCommand(NewCmdConfigSetAuthInfo(out, pathOptions))
	cmd.AddCommand(NewCmdConfigSetContext(out, pathOptions))
	//cmd.AddCommand(NewCmdConfigSet(out, pathOptions))
	//cmd.AddCommand(NewCmdConfigUnset(out, pathOptions))
	cmd.AddCommand(NewCmdConfigCurrentContext(out, pathOptions))
	//cmd.AddCommand(NewCmdConfigUseContext(out, pathOptions))
	cmd.AddCommand(NewCmdConfigGetContexts(out, pathOptions))
	//cmd.AddCommand(NewCmdConfigGetClusters(out, pathOptions))
	//cmd.AddCommand(NewCmdConfigDeleteCluster(out, pathOptions))
	//cmd.AddCommand(NewCmdConfigDeleteContext(out, errOut, pathOptions))
	//cmd.AddCommand(NewCmdConfigRenameContext(out, pathOptions))
	cmd.AddCommand(NewCmdConfigDeleteAuthInfo(out, errOut, pathOptions))
	return cmd
}

func toBool(propertyValue string) (bool, error) {
	boolValue := false
	if len(propertyValue) != 0 {
		var err error
		boolValue, err = strconv.ParseBool(propertyValue)
		if err != nil {
			return false, err
		}
	}

	return boolValue, nil
}

func helpErrorf(cmd *cobra.Command, format string, args ...interface{}) error {
	cmd.Help()
	msg := fmt.Sprintf(format, args...)
	return fmt.Errorf("%s\n", msg)
}
