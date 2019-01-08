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
	"fmt"
	"os"

	"github.com/spf13/pflag"

	"github.com/spf13/cobra"
)

var (
	format string
)

func headersFlag(header Header, flags *pflag.FlagSet) {
	flags.Var(
		header,
		"header",
		"set header to be passed in a http request, can be set multiple times",
	)
}

func formatFlag(flags *pflag.FlagSet) {
	flags.StringVar(
		&format,
		"format",
		"",
		"go template response formatting",
	)
}

// NewRootCommand creates root command a base command for gql
func NewRootCommand(args []string) *cobra.Command {
	var Endpoint string
	header := make(Header)
	path := Peek(args, &Endpoint, header, nil)
	rootCmd := &cobra.Command{
		Use:   "gql",
		Short: "GraphQL command line client",
		Long:  `Simple graphql command line client allowing user to execute GraphQL query against http GraphQL servers`,
	}
	rootCmd.SetArgs(args)
	introspectionCmd := NewIntrospectionCommand(IntrospectionCommandConfig{
		Endpoint: Endpoint,
		Path:     path,
		Header:   header,
	},
	)
	rootCmd.AddCommand(introspectionCmd.Command)
	rootCmd.AddCommand(NewRawCommand())
	rootCmd.AddCommand(NewCompletionCommand(CompletionCommandConfig{}))
	aliasFieldCommand(rootCmd, introspectionCmd.Query.FieldCommand)
	aliasFieldCommand(rootCmd, introspectionCmd.Mutation.FieldCommand)
	aliasFieldCommand(rootCmd, introspectionCmd.Subscription.FieldCommand)
	rootCmd.TraverseChildren = true
	return rootCmd
}

func aliasFieldCommand(cmd *cobra.Command, fc *FieldCommand) {
	if fc != nil {
		cmd.AddCommand(fc.Command)
	}
}

// Execute adds all child commands to the root command and sets flags appropriately.
// This is called by main.main(). It only needs to happen once to the rootCmd.
func Execute() {
	// Peek flags to find
	rootCmd := NewRootCommand(os.Args[1:])
	if err := rootCmd.Execute(); err != nil {
		fmt.Println(err)
		os.Exit(1)
	}
}
