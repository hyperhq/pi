package resource

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
	"encoding/json"
	"fmt"
	"io"
	"log"
	"os"
	"strings"

	"github.com/hyperhq/client-go/tools/clientcmd/api/hyper"
	"github.com/hyperhq/pi/pkg/pi/cmd/templates"
	cmdutil "github.com/hyperhq/pi/pkg/pi/cmd/util"
	"github.com/hyperhq/pi/pkg/pi/util/i18n"

	"github.com/golang/glog"
	"github.com/olekukonko/tablewriter"
	"github.com/spf13/cobra"
)

// NewCmdGetFip groups subcommands to get various zones of fips
func NewCmdGetFip(f cmdutil.Factory, cmdOut, errOut io.Writer) *cobra.Command {
	cmd := &cobra.Command{
		Use:     "fip IP",
		Short:   i18n.T("list fips or get a fip"),
		Long:    fipLong,
		Example: fipExample,
		Aliases: []string{"fips"},
		Run: func(cmd *cobra.Command, args []string) {
			err := GetFipGeneric(f, cmdOut, cmd, args)
			cmdutil.CheckErr(err)
		},
	}
	cmdutil.AddApplyAnnotationFlags(cmd)
	cmdutil.AddValidateFlags(cmd)
	cmdutil.AddPrinterFlags(cmd)
	cmdutil.AddGeneratorFlags(cmd, cmdutil.HyperFipV1GeneratorName)
	return cmd
}

var (
	fipLong = templates.LongDesc(i18n.T(`List fips or get a fip.`))

	fipExample = templates.Examples(i18n.T(`
	  # List fips
	  pi get fips

	  # Get a specified fip
	  pi get fip x.x.x.x`))
)

// GetFipGeneric is the implementation of the get fip generic command
func GetFipGeneric(f cmdutil.Factory, cmdOut io.Writer, cmd *cobra.Command, args []string) error {
	ip := IPFromCommandArgs(cmd, args)
	output := cmdutil.GetFlagString(cmd, "output")

	if cfg, err := f.ClientConfig(); err != nil {
		return err
	} else {
		hyperConn := hyper.NewHyperConn(cfg)
		fipCli := hyper.NewFipCli(hyperConn)
		if ip == "" {
			if _, fipList, err := fipCli.ListFips(); err != nil {
				return err
			} else {
				PrintFipListResult(output, fipList)
			}
		} else {
			if _, fip, err := fipCli.GetFip(ip); err != nil {
				return err
			} else {
				PrintFipGetResult(output, *fip)
			}
		}
	}
	return nil
}

func PrintFipListResult(output string, result []hyper.FipListResponse) {
	if output == "table" || output == "" {
		data := [][]string{}
		for _, fip := range result {
			item := []string{fip.Fip, fip.Name, fip.CreatedAt.Format("2006-01-02T15:04:05-07:00")}
			data = append(data, item)
		}
		table := tablewriter.NewWriter(os.Stdout)
		table.SetHeader([]string{"Fip", "Name", "CreatedAt"})

		//set table style
		table.SetBorder(false)
		table.SetHeaderLine(false)
		table.SetHeaderAlignment(tablewriter.ALIGN_LEFT)
		table.SetRowLine(false)
		table.SetColumnSeparator("")
		table.SetAlignment(tablewriter.ALIGN_LEFT)

		for _, fip := range data {
			table.Append(fip)
		}
		table.Render()
	} else if output == "json" {
		if buf, err := json.MarshalIndent(result, "", "  "); err != nil {
			log.Fatal(err)
		} else {
			fmt.Print(string(buf))
		}
	} else {
		glog.Warningf("--output support table,json")
	}
}

func PrintFipGetResult(output string, fip hyper.FipGetResponse) {
	if output == "table" || output == "" {
		data := [][]string{}

		item := []string{fip.Fip, fip.Name, fip.CreatedAt.Format("2006-01-02T15:04:05-07:00"), strings.Join(fip.Pods, ",")}
		data = append(data, item)

		table := tablewriter.NewWriter(os.Stdout)
		table.SetHeader([]string{"Fip", "Name", "CreatedAt", "Pods"})

		//set table style
		table.SetBorder(false)
		table.SetHeaderLine(false)
		table.SetHeaderAlignment(tablewriter.ALIGN_LEFT)
		table.SetRowLine(false)
		table.SetColumnSeparator("")
		table.SetAlignment(tablewriter.ALIGN_LEFT)

		for _, fip := range data {
			table.Append(fip)
		}
		table.Render()
	} else if output == "json" {
		if buf, err := json.MarshalIndent(fip, "", "  "); err != nil {
			log.Fatal(err)
		} else {
			fmt.Print(string(buf))
		}
	} else {
		glog.Warningf("--output support table,json")
	}
}

func IPFromCommandArgs(cmd *cobra.Command, args []string) string {
	if len(args) == 0 {
		return ""
	}
	return args[0]
}
