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
	"net/http"
	"os"
	"strings"

	"github.com/hyperhq/client-go/tools/clientcmd/api/hyper"
	"github.com/hyperhq/pi/pkg/pi/cmd/templates"
	cmdutil "github.com/hyperhq/pi/pkg/pi/cmd/util"
	"github.com/hyperhq/pi/pkg/pi/util/i18n"

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
	cmd.Flags().StringP("output", "o", "", "Output format. One of: json|ip")
	return cmd
}

var (
	fipLong = templates.LongDesc(i18n.T(`List fips or get a fip.`))

	fipExample = templates.Examples(i18n.T(`
	  # List fips
	  pi get fips

	  # Get a specified fip
	  pi get fip x.x.x.x

	  # Show ip only
	  pi get fip -o ip`))
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
				if len(fipList) == 0 {
					fmt.Println("No resources found.")
				} else {
					return PrintFipResult(output, true, fipList)
				}
			}
		} else {
			if httpStatus, fip, err := fipCli.GetFip(ip); err != nil {
				return err
			} else {
				if httpStatus == http.StatusNotFound {
					fmt.Println("No resources found.")
				} else {
					return PrintFipResult(output, false, []hyper.FipResponse{*fip})
				}
			}
		}
	}
	return nil
}

func PrintFipResult(output string, isList bool, result []hyper.FipResponse) error {
	if output == "" {
		data := [][]string{}
		for _, fip := range result {
			item := []string{fip.Fip, fip.Name, fip.CreatedAt.Format("2006-01-02T15:04:05-07:00"), strings.Join(fip.Services, ",")}
			data = append(data, item)
		}
		table := tablewriter.NewWriter(os.Stdout)
		table.SetHeader([]string{"Fip", "Name", "CreatedAt", "Services"})

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
		var (
			buf []byte
			err error
		)
		if !isList {
			buf, err = json.MarshalIndent(result[0], "", "  ")
			if err != nil {
				log.Fatal(err)
			}
		} else {
			buf, err = json.MarshalIndent(result, "", "  ")
			if err != nil {
				log.Fatal(err)
			}
		}
		fmt.Print(string(buf))
	} else if output == "ip" {
		for _, fip := range result {
			fmt.Printf("fips/%v\n", fip.Fip)
		}
	} else {
		err := fmt.Errorf("error: output format \"%v\" not recognized", output)
		return err
	}
	return nil
}

func IPFromCommandArgs(cmd *cobra.Command, args []string) string {
	if len(args) == 0 {
		return ""
	}
	return args[0]
}
