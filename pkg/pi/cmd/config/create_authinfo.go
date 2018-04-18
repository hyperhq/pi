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
	"errors"
	"fmt"
	"io"
	"io/ioutil"
	"path/filepath"
	"strings"

	"github.com/hyperhq/client-go/tools/clientcmd"
	clientcmdapi "github.com/hyperhq/client-go/tools/clientcmd/api"
	"github.com/hyperhq/pi/pkg/pi/cmd/templates"
	cmdutil "github.com/hyperhq/pi/pkg/pi/cmd/util"
	"github.com/hyperhq/pi/pkg/pi/util/i18n"

	"github.com/golang/glog"
	"github.com/spf13/cobra"
	"k8s.io/apiserver/pkg/util/flag"
)

type createAuthInfoOptions struct {
	configAccess      clientcmd.ConfigAccess
	name              string
	authPath          flag.StringFlag
	clientCertificate flag.StringFlag
	clientKey         flag.StringFlag
	token             flag.StringFlag
	username          flag.StringFlag
	password          flag.StringFlag
	embedCertData     flag.Tristate
	authProvider      flag.StringFlag

	authProviderArgs         map[string]string
	authProviderArgsToRemove []string

	region    flag.StringFlag
	accessKey flag.StringFlag
	secretKey flag.StringFlag
}

var (
	create_authinfo_long = fmt.Sprintf(templates.LongDesc(`
		Sets a user entry in pi config

		Specifying a name that already exists will merge new fields on top of existing values.

			Region flags:
			  --%v=region

			Credentials flags:
			  --%v=access_key --%v=secret_key

		Bearer token and basic auth are mutually exclusive.`), clientcmd.FlagRegion, clientcmd.FlagAccessKey, clientcmd.FlagSecretKey)

	create_authinfo_example = templates.Examples(`
		# Set credentials for hyper user
		pi config set-credentials user1 --region=gcp-us-central1 --access-key=xxxx --secret-key=xxxxxxxxx`)
)

func NewCmdConfigSetAuthInfo(out io.Writer, configAccess clientcmd.ConfigAccess) *cobra.Command {
	options := &createAuthInfoOptions{configAccess: configAccess}
	return newCmdConfigSetAuthInfo(out, options)
}

func newCmdConfigSetAuthInfo(out io.Writer, options *createAuthInfoOptions) *cobra.Command {
	cmd := &cobra.Command{
		Use:     fmt.Sprintf("set-credentials NAME [--%v=access_key] [--%v=secret_key] [--%v=region] ", clientcmd.FlagAccessKey, clientcmd.FlagSecretKey, clientcmd.FlagRegion),
		Short:   i18n.T("Sets a user entry in pi config"),
		Long:    create_authinfo_long,
		Example: create_authinfo_example,
		Run: func(cmd *cobra.Command, args []string) {
			err := options.complete(cmd, out)
			if err != nil {
				cmd.Help()
				cmdutil.CheckErr(err)
			}
			cmdutil.CheckErr(options.run())
			fmt.Fprintf(out, "User %q set.\n", options.name)
		},
	}

	cmd.Flags().Var(&options.accessKey, clientcmd.FlagAccessKey, clientcmd.FlagAccessKey+" for the user entry in pi config")
	cmd.Flags().Var(&options.secretKey, clientcmd.FlagSecretKey, clientcmd.FlagSecretKey+" for the user entry in pi config")
	cmd.Flags().Var(&options.region, clientcmd.FlagRegion, "region for the user entry in pi config")

	return cmd
}

func (o createAuthInfoOptions) run() error {
	err := o.validate()
	if err != nil {
		return err
	}

	config, err := o.configAccess.GetStartingConfig()
	if err != nil {
		return err
	}

	//default cluster info
	startingStanzaCluster, exists := config.Clusters[clientcmd.DefaultCluster]
	if !exists {
		startingStanzaCluster = clientcmdapi.NewCluster()
		cluster := o.modifyDefaultCluster(*startingStanzaCluster)
		config.Clusters[clientcmd.DefaultCluster] = &cluster
	}

	//default context
	startingStanzaContext, exists := config.Contexts[clientcmd.DefaultContext]
	if !exists {
		startingStanzaContext = clientcmdapi.NewContext()
		context := o.modifyDefaultContext(o.name, *startingStanzaContext)
		config.Contexts[clientcmd.DefaultContext] = &context
	}

	//default auth info
	startingStanzaAuth, exists := config.AuthInfos[o.name]
	if !exists {
		startingStanzaAuth = clientcmdapi.NewAuthInfo()
	}
	authInfo := o.modifyAuthInfo(*startingStanzaAuth)
	config.AuthInfos[o.name] = &authInfo

	if config.CurrentContext == "" {
		config.CurrentContext = "default"
	}

	if err := clientcmd.ModifyConfig(o.configAccess, *config, true); err != nil {
		return err
	}

	return nil
}

// cluster builds a Cluster object from the options
func (o *createAuthInfoOptions) modifyDefaultCluster(existingCluster clientcmdapi.Cluster) clientcmdapi.Cluster {
	modifiedCluster := existingCluster

	if modifiedCluster.Server == "" {
		modifiedCluster.Server = clientcmd.DefaultServer
	}

	modifiedCluster.InsecureSkipTLSVerify = true

	return modifiedCluster
}

func (o *createAuthInfoOptions) modifyDefaultContext(user string, existingContext clientcmdapi.Context) clientcmdapi.Context {
	modifiedContext := existingContext

	modifiedContext.Cluster = clientcmd.DefaultCluster
	modifiedContext.AuthInfo = user
	modifiedContext.Namespace = "default"

	return modifiedContext
}

// authInfo builds an AuthInfo object from the options
func (o *createAuthInfoOptions) modifyAuthInfo(existingAuthInfo clientcmdapi.AuthInfo) clientcmdapi.AuthInfo {
	modifiedAuthInfo := existingAuthInfo

	var setToken, setBasic, setCredential bool

	if o.clientCertificate.Provided() {
		certPath := o.clientCertificate.Value()
		if o.embedCertData.Value() {
			modifiedAuthInfo.ClientCertificateData, _ = ioutil.ReadFile(certPath)
			modifiedAuthInfo.ClientCertificate = ""
		} else {
			certPath, _ = filepath.Abs(certPath)
			modifiedAuthInfo.ClientCertificate = certPath
			if len(modifiedAuthInfo.ClientCertificate) > 0 {
				modifiedAuthInfo.ClientCertificateData = nil
			}
		}
	}
	if o.clientKey.Provided() {
		keyPath := o.clientKey.Value()
		if o.embedCertData.Value() {
			modifiedAuthInfo.ClientKeyData, _ = ioutil.ReadFile(keyPath)
			modifiedAuthInfo.ClientKey = ""
		} else {
			keyPath, _ = filepath.Abs(keyPath)
			modifiedAuthInfo.ClientKey = keyPath
			if len(modifiedAuthInfo.ClientKey) > 0 {
				modifiedAuthInfo.ClientKeyData = nil
			}
		}
	}

	if o.token.Provided() {
		modifiedAuthInfo.Token = o.token.Value()
		setToken = len(modifiedAuthInfo.Token) > 0
	}

	if o.username.Provided() {
		modifiedAuthInfo.Username = o.username.Value()
		setBasic = setBasic || len(modifiedAuthInfo.Username) > 0
	}
	if o.password.Provided() {
		modifiedAuthInfo.Password = o.password.Value()
		setBasic = setBasic || len(modifiedAuthInfo.Password) > 0
	}
	if o.authProvider.Provided() {
		newName := o.authProvider.Value()

		// Only overwrite if the existing auth-provider is nil, or different than the newly specified one.
		if modifiedAuthInfo.AuthProvider == nil || modifiedAuthInfo.AuthProvider.Name != newName {
			modifiedAuthInfo.AuthProvider = &clientcmdapi.AuthProviderConfig{
				Name: newName,
			}
		}
	}

	//patch for hyper
	if o.region.Provided() {
		modifiedAuthInfo.Region = o.region.Value()
		setCredential = setCredential || len(modifiedAuthInfo.Region) > 0
	} else {
		modifiedAuthInfo.Region = clientcmd.DefaultRegion
		setCredential = setCredential || len(modifiedAuthInfo.Region) > 0
	}
	if o.accessKey.Provided() {
		modifiedAuthInfo.AccessKey = o.accessKey.Value()
		setCredential = setCredential || len(modifiedAuthInfo.AccessKey) > 0
	}
	if o.secretKey.Provided() {
		modifiedAuthInfo.SecretKey = o.secretKey.Value()
		setCredential = setCredential || len(modifiedAuthInfo.SecretKey) > 0
	}

	if modifiedAuthInfo.AuthProvider != nil {
		if modifiedAuthInfo.AuthProvider.Config == nil {
			modifiedAuthInfo.AuthProvider.Config = make(map[string]string)
		}
		for _, toRemove := range o.authProviderArgsToRemove {
			delete(modifiedAuthInfo.AuthProvider.Config, toRemove)
		}
		for key, value := range o.authProviderArgs {
			modifiedAuthInfo.AuthProvider.Config[key] = value
		}
	}

	// If any auth info was set, make sure any other existing auth types are cleared
	if setToken || setBasic || setCredential {
		if !setToken {
			modifiedAuthInfo.Token = ""
		}
		if !setBasic {
			modifiedAuthInfo.Username = ""
			modifiedAuthInfo.Password = ""
		}
		if !setCredential {
			modifiedAuthInfo.Region = ""
			modifiedAuthInfo.AccessKey = ""
			modifiedAuthInfo.SecretKey = ""
		}
	}

	glog.V(4).Infof("new: %v(%v/%v) existingAuthInfo: %v(%v/%v) modifiedAuthInfo: %v(%v/%v)",
		o.region, o.accessKey, o.secretKey,
		existingAuthInfo.Region, existingAuthInfo.AccessKey, existingAuthInfo.SecretKey,
		modifiedAuthInfo.Region, modifiedAuthInfo.AccessKey, modifiedAuthInfo.SecretKey)

	return modifiedAuthInfo
}

func (o *createAuthInfoOptions) complete(cmd *cobra.Command, out io.Writer) error {
	args := cmd.Flags().Args()
	if len(args) != 1 {
		return fmt.Errorf("Unexpected args: %v", args)
	}

	o.name = args[0]
	return nil
}

func (o createAuthInfoOptions) validate() error {
	if len(o.name) == 0 {
		return errors.New("you must specify a non-empty user name")
	}
	methods := []string{}
	if len(o.token.Value()) > 0 {
		methods = append(methods, fmt.Sprintf("--%v", clientcmd.FlagBearerToken))
	}
	if len(o.username.Value()) > 0 || len(o.password.Value()) > 0 {
		methods = append(methods, fmt.Sprintf("--%v/--%v", clientcmd.FlagUsername, clientcmd.FlagPassword))
	}
	if len(methods) > 1 {
		return fmt.Errorf("you cannot specify more than one authentication method at the same time: %v", strings.Join(methods, ", "))
	}
	if o.embedCertData.Value() {
		certPath := o.clientCertificate.Value()
		keyPath := o.clientKey.Value()
		if certPath == "" && keyPath == "" {
			return fmt.Errorf("you must specify a --%s or --%s to embed", clientcmd.FlagCertFile, clientcmd.FlagKeyFile)
		}
		if certPath != "" {
			if _, err := ioutil.ReadFile(certPath); err != nil {
				return fmt.Errorf("error reading %s data from %s: %v", clientcmd.FlagCertFile, certPath, err)
			}
		}
		if keyPath != "" {
			if _, err := ioutil.ReadFile(keyPath); err != nil {
				return fmt.Errorf("error reading %s data from %s: %v", clientcmd.FlagKeyFile, keyPath, err)
			}
		}
	}

	return nil
}
