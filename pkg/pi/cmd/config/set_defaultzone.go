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
	"net/http"
	"path/filepath"
	"strings"

	"github.com/hyperhq/client-go/tools/clientcmd"
	clientcmdapi "github.com/hyperhq/client-go/tools/clientcmd/api"
	"github.com/hyperhq/client-go/tools/clientcmd/api/hyper"
	"github.com/hyperhq/pi/pkg/pi/cmd/templates"
	cmdutil "github.com/hyperhq/pi/pkg/pi/cmd/util"
	"github.com/hyperhq/pi/pkg/pi/util/i18n"

	"github.com/golang/glog"
	restclient "github.com/hyperhq/client-go/rest"
	"github.com/spf13/cobra"
	"k8s.io/apiserver/pkg/util/flag"
)

type setDefaultZoneOptions struct {
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

	server flag.StringFlag
	region flag.StringFlag
	zone   flag.StringFlag
}

var (
	FlagZone = "zone"

	set_defaultzone_long = fmt.Sprintf(templates.LongDesc(`
		Change default zone of user

		Specifying a name that already exists will merge new fields on top of existing values.

			APIServer flags:
			  --%v=server

			Region flags:
			  --%v=region

			DefaultZone flags:
			  --%v=zone

		`), clientcmd.FlagAPIServer, clientcmd.FlagRegion, FlagZone)

	set_defaultzone_example = templates.Examples(`
		# Change default zone of user
		pi config set-default-zone user1 --region=gcp-us-central1 --zone=gcp-us-central1-c`)
)

func NewCmdConfigSetDefaultZone(f cmdutil.Factory, out io.Writer, configAccess clientcmd.ConfigAccess) *cobra.Command {
	options := &setDefaultZoneOptions{configAccess: configAccess}
	return newCmdConfigSetDefaultZone(f, out, options)
}

func newCmdConfigSetDefaultZone(f cmdutil.Factory, out io.Writer, options *setDefaultZoneOptions) *cobra.Command {
	cmd := &cobra.Command{
		Use:     fmt.Sprintf("set-default-zone NAME [--%v=region] [--%v=zone] ", clientcmd.FlagRegion, FlagZone),
		Short:   i18n.T("Change default zone of user"),
		Long:    set_defaultzone_long,
		Example: set_defaultzone_example,
		Run: func(cmd *cobra.Command, args []string) {
			err := options.complete(cmd, out)
			if err != nil {
				cmd.Help()
				cmdutil.CheckErr(err)
			}
			cmdutil.CheckErr(options.run(f))
			fmt.Fprintf(out, "Change default zone of user %q to %q.\n", options.name, options.zone)
		},
	}

	cmd.Flags().Var(&options.server, clientcmd.FlagAPIServer, "apiserver url")
	cmd.Flags().Var(&options.region, clientcmd.FlagRegion, "region for the user")
	cmd.Flags().Var(&options.zone, FlagZone, "default zone for the user")

	return cmd
}

func (o setDefaultZoneOptions) run(f cmdutil.Factory) error {
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

	if o.zone.String() == "" {
		return fmt.Errorf("zone can not be empty")
	}

	if o.region.String() == "" {
		o.region.Set(config.AuthInfos[o.name].Region)
	}
	if o.region.String() == "" {
		return fmt.Errorf("region can not be empty")
	}

	if o.server.String() == "" {
		o.server.Set(startingStanzaCluster.Server)
	}
	if o.server.String() == "" {
		return fmt.Errorf("server can not be empty")
	}

	glog.V(4).Infof("%v %v %v %v\n", o.server, o.name, o.region, o.zone)
	glog.V(4).Infof("%v %v\n", authInfo.AccessKey, authInfo.SecretKey)

	var cfg = &restclient.Config{
		Host: o.server.String(),
		CredentialConfig: restclient.CredentialConfig{
			Region:    o.region.String(),
			AccessKey: authInfo.AccessKey,
			SecretKey: authInfo.SecretKey,
		},
	}
	return o.updateDefaultZone(f, cfg, o.zone.String())
}

func (o *setDefaultZoneOptions) updateDefaultZone(f cmdutil.Factory, cfg *restclient.Config, zone string) error {
	var tenant string
	hyperConn := hyper.NewHyperConn(cfg)
	infoCli := hyper.NewInfoCli(hyperConn)
	if _, info, err := infoCli.GetInfo(); err != nil {
		return fmt.Errorf("failed to get tenant id for user, error:%v", err)
	} else {
		tenant = info["TenantID"]
	}
	if tenant == "" {
		return fmt.Errorf("missing tenant id")
	}
	userCli := hyper.NewUserCli(hyperConn)
	if httpStatus, err := userCli.UpdateDefaultZone(tenant, zone); err != nil {
		return err
	} else {
		if httpStatus != http.StatusOK {
			return fmt.Errorf("failed to update defautl zone, httpstatus=%v", httpStatus)
		}
	}
	return nil
}

// cluster builds a Cluster object from the options
func (o *setDefaultZoneOptions) modifyDefaultCluster(existingCluster clientcmdapi.Cluster) clientcmdapi.Cluster {
	modifiedCluster := existingCluster

	if modifiedCluster.Server == "" {
		modifiedCluster.Server = clientcmd.DefaultServer
	}

	modifiedCluster.InsecureSkipTLSVerify = true

	return modifiedCluster
}

func (o *setDefaultZoneOptions) modifyDefaultContext(user string, existingContext clientcmdapi.Context) clientcmdapi.Context {
	modifiedContext := existingContext

	modifiedContext.Cluster = clientcmd.DefaultCluster
	modifiedContext.AuthInfo = user
	modifiedContext.Namespace = "default"

	return modifiedContext
}

// authInfo builds an AuthInfo object from the options
func (o *setDefaultZoneOptions) modifyAuthInfo(existingAuthInfo clientcmdapi.AuthInfo) clientcmdapi.AuthInfo {
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

	glog.V(4).Infof("new: %v existingAuthInfo: %v(%v/%v) modifiedAuthInfo: %v(%v/%v)",
		o.region,
		existingAuthInfo.Region, existingAuthInfo.AccessKey, existingAuthInfo.SecretKey,
		modifiedAuthInfo.Region, modifiedAuthInfo.AccessKey, modifiedAuthInfo.SecretKey)

	return modifiedAuthInfo
}

func (o *setDefaultZoneOptions) complete(cmd *cobra.Command, out io.Writer) error {
	args := cmd.Flags().Args()
	if len(args) != 1 {
		return fmt.Errorf("please specify the NAME")
	}

	o.name = args[0]
	return nil
}

func (o setDefaultZoneOptions) validate() error {
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
