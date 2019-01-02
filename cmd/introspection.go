// Copyright © 2019 NAME HERE <EMAIL ADDRESS>
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

	"github.com/spf13/cobra"
	"github.com/spf13/pflag"
)

var (
	noCache bool
)

// find GraphQL endpoint option and query path
func Peek(
	cargs []string,
	endpoint *string,
	header Header,
	flagset *pflag.FlagSet) []string {
	if flagset == nil {
		flagset = pflag.NewFlagSet(os.Args[0], pflag.ExitOnError)
		// workaround to ignore help on peek
		flagset.Bool("help", false, "silent no op")
	}
	// Unknown flag is not an error
	// as we are only peeking ahead
	// to find endpoint option and
	// query path.
	flagset.ParseErrorsWhitelist = pflag.ParseErrorsWhitelist{
		UnknownFlags: true,
	}
	endpointFlag(endpoint, flagset)
	headersFlag(header, flagset)
	if err := flagset.Parse(cargs); err != nil {
		fmt.Fprintln(os.Stderr, err)
	}
	args := flagset.Args()
	if len(args) == 0 {
		return args
	}
	// Remove leading gql, completion and intrsopection
	// keywords
	for (len(args) > 0) && (args[0] == "completion" || args[0] == "gql" || args[0] == "introspection") {
		args = args[1:]
	}
	return args
}

func endpointFlag(endpoint *string, flags *pflag.FlagSet) {
	flags.StringVar(endpoint, "endpoint", "", "graphql endpoint")
}

func requiredEndpointFlag(endpoint *string, flags *pflag.FlagSet) {
	endpointFlag(endpoint, flags)
	cobra.MarkFlagRequired(flags, "endpoint")
}

func noCacheFlag(flags *pflag.FlagSet) {
	flags.BoolVar(
		&noCache,
		"no-cache",
		false,
		"do not cache schema introspection result",
	)
}

type IntrospectionCommandConfig struct {
	Path     []string
	Endpoint string
	Header   Header
}

type IntrospectionCommand struct {
	*cobra.Command
	GraphQLRootCommands
	Config IntrospectionCommandConfig
}

func (i *IntrospectionCommand) appendDyn(header Header, endpoint *string, cmd *FieldCommand) {
	if cmd != nil {
		i.AddCommand(cmd.Command)
		requiredEndpointFlag(endpoint, cmd.PersistentFlags())
		formatFlag(cmd.PersistentFlags())
		noCacheFlag(cmd.PersistentFlags())
		headersFlag(header, cmd.PersistentFlags())
	}
}

func NewIntrospectionCommand(config IntrospectionCommandConfig) IntrospectionCommand {
	introspectionCmd := IntrospectionCommand{
		Command: &cobra.Command{
			Use:   "introspection",
			Short: "graphql commands depending on schema introspection",
			Long:  `Root command for commands that require GraphQL schema introspection.`,
		},
		Config: config,
	}
	endpoint := &introspectionCmd.Config.Endpoint
	header := introspectionCmd.Config.Header
	if config.Endpoint != "" {
		var err error
		introspectionCmd.GraphQLRootCommands, err = NewGraphQLRootCommands(GraphQLRootConfig{
			Endpoint: *endpoint,
			Path:     config.Path,
			Header:   header,
		})
		if err != nil {
			fmt.Fprintln(os.Stderr, err)
			os.Exit(1)
		}
		introspectionCmd.appendDyn(header, endpoint, introspectionCmd.Query.FieldCommand)
		introspectionCmd.appendDyn(header, endpoint, introspectionCmd.Mutation.FieldCommand)
		introspectionCmd.appendDyn(header, endpoint, introspectionCmd.Subscription.FieldCommand)
	}
	introspectionCmd.AddCommand(NewFieldsCommand(introspectionCmd.GraphQLRootCommands.Schema))
	introspectionCmd.AddCommand(NewArgsCommand(
		ArgsCommandConfig{
			Schema: introspectionCmd.GraphQLRootCommands.Schema,
		},
	).Command)
	requiredEndpointFlag(endpoint, introspectionCmd.PersistentFlags())
	formatFlag(introspectionCmd.PersistentFlags())
	noCacheFlag(introspectionCmd.PersistentFlags())
	headersFlag(header, introspectionCmd.Flags())
	return introspectionCmd
}
