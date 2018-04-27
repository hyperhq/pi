package cmd

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

import (
	"io"
	"log"
	"os"
	"strings"
	"time"

	"github.com/hyperhq/client-go/tools/clientcmd/api/hyper"
	"github.com/hyperhq/pi"
	"github.com/hyperhq/pi/pkg/pi/cmd/templates"
	cmdutil "github.com/hyperhq/pi/pkg/pi/cmd/util"
	"github.com/hyperhq/pi/pkg/pi/util/i18n"

	"github.com/olekukonko/tablewriter"
	"github.com/spf13/cobra"
)

// NewCmdInfo groups subcommands to get various zones of infos
func NewCmdInfo(f cmdutil.Factory, cmdOut, errOut io.Writer) *cobra.Command {
	cmd := &cobra.Command{
		Use:     "info",
		Short:   i18n.T("Print region and user info"),
		Long:    infoLong,
		Example: infoExample,
		Run: func(cmd *cobra.Command, args []string) {
			err := InfoGeneric(f, cmdOut, cmd, args)
			cmdutil.CheckErr(err)
		},
	}
	cmd.Flags().BoolP("check-update", "c", false, "force to check new version of pi")
	return cmd
}

var (
	infoLong = templates.LongDesc(i18n.T(`Print region and user info`))

	infoExample = templates.Examples(i18n.T(`
	  # Print region and user info
	  pi info`))
)

// InfoGeneric is the implementation of the get info generic command
func InfoGeneric(f cmdutil.Factory, cmdOut io.Writer, cmd *cobra.Command, args []string) error {
	if cfg, err := f.ClientConfig(); err != nil {
		return err
	} else {
		hyperConn := hyper.NewHyperConn(cfg)
		infoCli := hyper.NewInfoCli(hyperConn)
		if _, info, err := infoCli.GetInfo(); err != nil {
			return err
		} else {
			PrintInfoResult(info)
		}
	}

	updater := pi.NewCheckUpdate()
	if cmdutil.GetFlagBool(cmd, "check-update") {
		//force check version
		pi.CheckRelease()
		updater.WriteTime(time.Now())
	} else {
		//check version after 24 hours
		lastUpdate := updater.ReadTime()
		hours := time.Since(lastUpdate).Hours()
		if hours >= 24 {
			if os.Getenv("HYPER_DEBUG") == "true" {
				log.Printf("More than 24 hours(%v) of uncheck version.", int(hours))
			}
			//start check new version
			pi.CheckRelease()
			updater.WriteTime(time.Now())
		} else {
			if os.Getenv("HYPER_DEBUG") == "true" {
				log.Printf("Checked version in 24 hours(%v), skip.", int(hours))
			}
		}
	}

	return nil
}

func PrintInfoResult(result map[string]string) {
	data := [][]string{}
	propertyList := []string{
		"Region Info:",
		"  Region",
		"  AvailabilityZone",
		"  ServiceClusterIPRange",
		"Account Info:",
		"  Email",
		"  TenantID",
		"  DefaultZone",
		"  Resources",
		"Version Info:",
	}
	for _, p := range propertyList {
		data = getProperty(p, result, data)
	}
	data = getVersion(data)

	table := tablewriter.NewWriter(os.Stdout)

	//set table style
	table.SetBorder(false)
	table.SetHeaderLine(false)
	table.SetHeaderAlignment(tablewriter.ALIGN_LEFT)
	table.SetRowLine(false)
	table.SetColumnSeparator("")
	table.SetAlignment(tablewriter.ALIGN_LEFT)

	for _, info := range data {
		table.Append(info)
	}
	table.Render()
}

func getProperty(property string, result map[string]string, data [][]string) [][]string {
	if v, ok := result[strings.TrimSpace(property)]; ok {
		item := []string{property, v}
		data = append(data, item)
	} else {
		item := []string{property, ""}
		data = append(data, item)
	}
	return data
}

func getVersion(data [][]string) [][]string {
	item := []string{"  Version", pi.Version}
	data = append(data, item)

	item = []string{"  Hash", pi.Commit}
	data = append(data, item)

	item = []string{"  Build", pi.Build}
	data = append(data, item)

	return data
}
