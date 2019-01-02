// Copyright © 2018 NAME HERE <EMAIL ADDRESS>
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//     http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

package cmd

import (
	"log"
	"net/http"

	"github.com/slothking-online/gql/client"

	"github.com/spf13/cobra"
)

// rawCmd represents the raw command
func NewRawCommand() *cobra.Command {
	var Endpoint string
	header := make(Header)
	rawCmd := &cobra.Command{
		Use:   "raw",
		Short: "Execute raw graphql query",
		Long: `Executes raw GraphQL query against http GraphQL backend.

Takes exactly one argument, which is graphql query string.`,
		Run: func(cmd *cobra.Command, args []string) {
			if len(args) != 1 {
				log.Panicln("command takes exactly one argument")
			}
			var httpHeader http.Header
			for k, v := range header {
				httpHeader.Add(k, v)
			}
			r := client.Raw{
				Query:         args[1],
				Variables:     map[string]interface{}(variables),
				OperationName: operationName,
				Header:        httpHeader,
			}
			cli := client.New(client.Config{
				Endpoint: Endpoint,
			})
			execute(cli, r, nil)
		},
	}
	requiredEndpointFlag(&Endpoint, rawCmd.Flags())
	formatFlag(rawCmd.Flags())
	headersFlag(header, rawCmd.Flags())
	rawCmd.PersistentFlags().Var(
		variables,
		"set",
		"set grapqhl query variable, can be set multiple times",
	)
	rawCmd.PersistentFlags().StringVar(
		&operationName,
		"operation-name",
		"",
		"grapqhl operation name if provided query has more than one operation defined",
	)
	return rawCmd
}
