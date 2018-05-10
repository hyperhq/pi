/*
Copyright 2016 The Kubernetes Authors.

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

	"github.com/hyperhq/client-go/tools/clientcmd"
	"github.com/hyperhq/pi/pkg/pi/cmd/templates"
	cmdutil "github.com/hyperhq/pi/pkg/pi/cmd/util"
	"github.com/hyperhq/pi/pkg/pi/util/i18n"
	"github.com/spf13/cobra"
)

var (
	delete_authinfo_example = templates.Examples(`
		# Delete the credentials
		pi config delete-credentials user1`)
)

func NewCmdConfigDeleteAuthInfo(out, errOut io.Writer, configAccess clientcmd.ConfigAccess) *cobra.Command {
	cmd := &cobra.Command{
		Use:     "delete-credentials NAME",
		Short:   i18n.T("Delete the specified credentials from the pi config"),
		Long:    "Delete the specified credentials from the pi config",
		Example: delete_authinfo_example,
		Run: func(cmd *cobra.Command, args []string) {
			err := runDeleteAuthInfo(out, errOut, configAccess, cmd)
			cmdutil.CheckErr(err)
		},
	}

	return cmd
}

func runDeleteAuthInfo(out, errOut io.Writer, configAccess clientcmd.ConfigAccess, cmd *cobra.Command) error {
	config, err := configAccess.GetStartingConfig()
	if err != nil {
		return err
	}

	args := cmd.Flags().Args()
	if len(args) != 1 {
		cmd.Help()
		return nil
	}

	configFile := configAccess.GetDefaultFilename()
	if configAccess.IsExplicitFile() {
		configFile = configAccess.GetExplicitFile()
	}

	name := args[0]
	_, exists := config.AuthInfos[name]
	if !exists {
		return fmt.Errorf("credentials %v not found", name)
	}

	delete(config.AuthInfos, name)

	if err := clientcmd.ModifyConfig(configAccess, *config, true); err != nil {
		return err
	}

	fmt.Fprintf(out, "deleted credentials %s from %s\n", name, configFile)

	return nil
}
