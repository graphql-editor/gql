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
	"strings"

	"github.com/spf13/pflag"

	shellquote "github.com/kballard/go-shellquote"

	"github.com/spf13/cobra"
)

func NewCompletionCommand() *cobra.Command {
	compl := &cobra.Command{
		Use:   "completion",
		Short: "Generates bash completion scripts",
		Long: `To load completion run

    . <(gql completion)

    To configure your bash shell to load completions for each session add to your bashrc

    # ~/.bashrc or ~/.profile
    . <(gql completion)
    `,
		Run: func(cmd *cobra.Command, args []string) {
			if len(args) == 0 {
				fmt.Println(`#/usr/bin/env bash

# gql completion script
# aside from a few commands, everything
# depends on graphql schema that is being loaded.
# Because of that command is  completly dynamic and
# most of it's functionality is unknown
# until the moment of execution.
# Whole completion solving is left to
# the gql, this script being only
# a proxy calling it

_gql_completions() {
    cur="${COMP_WORDS[COMP_CWORD]}"
    COMPREPLY=($(compgen -W "$(gql completion "${COMP_LINE}")" -- "${cur}"))
}
complete -F _gql_completions gql`)
				return
			}
			var err error
			args, err = shellquote.Split(args[0])
			// Strip leading command name
			args = args[1:]
			if err != nil {
				fmt.Fprintln(os.Stderr, "")
				fmt.Fprintln(os.Stderr, err)
				os.Exit(1)
			}
			cmd = NewRootCommand(args)

			// Parse all the complation args
			// to create dynamic bash completion
			// in application

			cmd, _, err = cmd.Traverse(args)
			if err != nil {
				fmt.Fprintln(os.Stderr, err)
				os.Exit(1)
			}
			subCommands := cmd.Commands()
			complReply := make([]string, 0, len(subCommands))
			for _, child := range cmd.Commands() {
				complReply = append(complReply, child.Name())
			}
			addFlagArg := func(f *pflag.Flag) {
				complReply = append(complReply, "--"+f.Name)
			}
			cmd.Flags().VisitAll(addFlagArg)
			cmd.PersistentFlags().VisitAll(addFlagArg)
			fmt.Println(strings.Join(complReply, " "))
		},
	}
	compl.Flags().SetInterspersed(false)
	return compl
}
