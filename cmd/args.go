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
	"strings"

	"github.com/slothking-online/gql/introspection"

	"github.com/spf13/cobra"
)

// ArgsSchema defines interface that must be implemented
// by schema for args command
type ArgsSchema interface {
	// FieldForPath returns field for object for path, or false
	// if not found
	FieldForPath([]string) (introspection.Field, bool)
}

type ArgsCommandConfig struct {
	Config
	Schema ArgsSchema
}

// argsCmd represents the args command
func NewArgsCommand(config ArgsCommandConfig) *Command {

	return &Command{
		Command: &cobra.Command{
			Use:   "args",
			Short: "Returns field arguments",
			Long:  `Return all argument names that are defined on this field by GraphQL schema.`,
			Run: func(cmd *cobra.Command, args []string) {
				f, ok := config.Schema.FieldForPath(args)
				if !ok {
					// quick bail, no aproximate matching or anything
					// either field exists or it does not
					fmt.Fprintf(config.Error(), "path not found in schema\n") // nolint: errcheck
					config.Exit(1)
					// in case of exit function
					// not actually exiting, return here
					return
				}
				fmt.Fprintln(config.Output(), strings.Join(f.ArgNames(), " ")) // nolint: errcheck
			},
		},
		Config: config.Config,
	}
}
