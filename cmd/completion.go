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
	"bytes"
	"errors"
	"fmt"

	"github.com/spf13/pflag"

	shellquote "github.com/kballard/go-shellquote"

	"github.com/spf13/cobra"
)

type CommandBuilder interface {
	New([]string) *cobra.Command
}

type CommandBuilderFunc func([]string) *cobra.Command

func (d CommandBuilderFunc) New(args []string) *cobra.Command {
	return d(args)
}

type CompletionCommandConfig struct {
	Config
	CommandBuilder
}

type completionType uint

func (c completionType) String() string {
	switch c {
	case cmd:
		return "cmd"
	case opt:
		return "opt"
	}
	return ""
}

const (
	bashCompletion = `#/usr/bin/env bash

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
    completion="$(gql completion "${COMP_LINE}")"
    cmd="$(echo "${completion}" | grep '^cmd' | cut -d ':' -f 2 | tr "\n" " ")"
    opt="$(echo "${completion}" | grep '^opt' | cut -d ':' -f 2 | tr "\n" " ")"
    COMPREPLY=($(compgen -W "${cmd} ${opt}" -- "${cur}"))
}

complete -F _gql_completions gql`
	zshCompletion = `clean() {
    # Use -E for ERE as it's supported in both
    # FreeBSD sed and GNU sed
	echo "${1}" | \
		cut -c 5- | \
		sed 's/:true$//' | \
        sed 's/:false$//' | \
		sed -E 's/^([^:]*):(.*)/\1\\:"\2"/g' | \
		tr "\n" " " | sed -e 's/^[ \t]*//;s/[ \t]*$//'
}
_gql_parse_fields() {
	echo "${1}" | grep "^cmd:"
}
_gql_parse_opts() {
	echo "${1}" | grep "^opt:"
}
_gql_parse_args() {
	echo "$(_gql_parse_opts "${1}" | grep "^opt:--arg.*:true$")"
}
_gql_args_val() {
	args="$(_gql_parse_args "${1}")"
	echo "$(clean "${args}")"
}
_gql_fields() {
	fields="$(_gql_parse_fields "${1}" | sed -e 's/:true$//g')"
	echo "$(clean "${fields}")"
}
_gql_opts_val() {
	opts_val="$(_gql_parse_opts "${1}" | grep -v "^opt:--arg" | grep ":true$")"
	echo "$(clean "${opts_val}")"
}
_gql_opts_noval() {
	opts_noval="$(_gql_parse_opts "${1}" | grep -v "^opt:--arg" | grep ":false$")"
	echo "$(clean "${opts_noval}")"
}
_gql_completions() {
	completion="$(gql completion "${words}" 2>/dev/null)"
	if [ "$?" != "0" ]; then
		return
	fi
	_alternative \
		'args=:arguments with additional value:(('"$(_gql_args_val "${completion}")"'))' \
		'args=:available fields:(('"$(_gql_fields "${completion}")"'))' \
		'args=:special options with arg:(('"$(_gql_opts_val "${completion}")"'))' \
		'args:special options without arg:(('"$(_gql_opts_noval "${completion}")"'))'
}
compdef _gql_completions gql -p "gql *"`
	cmd completionType = iota
	opt
)

type completion struct {
	cType       completionType
	name        string
	description string
	hasArg      bool
}

func NewCompletionCommand(config CompletionCommandConfig) *cobra.Command {
	compl := &cobra.Command{
		Use:   "completion",
		Short: "Generates bash completion scripts",
		Long: `To load completion run

    . <(gql completion <bash|zsh>)

    To configure your bash shell to load completions for each session add to your bashrc

    # ~/.bashrc or ~/.profile
    . <(gql completion <bash|zsh>)`,
		RunE: func(cmd *cobra.Command, args []string) (err error) {
			if len(args) == 0 {
				return errors.New("command requires one argument")
			}
			switch args[0] {
			case "bash":
				_, err = fmt.Fprintln(config.Output(), bashCompletion)
				return
			case "zsh":
				_, err = fmt.Fprintln(config.Output(), zshCompletion)
				return
			}
			args, err = shellquote.Split(args[0])
			// Strip leading command name
			if err != nil {
				return
			}
			args = args[1:]
			commandBuilder := config.CommandBuilder
			if commandBuilder == nil {
				commandBuilder = CommandBuilderFunc(NewRootCommand)
			}
			cmd = commandBuilder.New(args)

			// Parse all the complation args
			// to create dynamic bash completion
			// in application

			cmd, _, err = cmd.Traverse(args)
			if err != nil {
				return err
			}
			completions := getFlagCompletions(cmd)
			completions = append(completions, getSubcommandCompletions(cmd)...)
			buf := &bytes.Buffer{}
			for _, c := range completions {
				if _, err = fmt.Fprintf(buf, "%s:%s:%s:%t\n", c.cType.String(), c.name, c.description, c.hasArg); err != nil {
					return
				}
			}
			if buf.Len() != 0 {
				_, err = fmt.Fprint(config.Output(), buf.String())
			}
			return
		},
	}
	compl.Flags().SetInterspersed(false)
	return compl
}

func getSubcommandCompletions(c *cobra.Command) []completion {
	sub := make([]completion, 0, len(c.Commands()))
	for _, child := range c.Commands() {
		sub = append(sub, completion{
			cType:       cmd,
			description: child.Short,
			name:        child.Name(),
		})
	}
	return sub
}

func getFlagCompletions(c *cobra.Command) []completion {
	completions := make([]completion, 0)
	addCompletion := func(f *pflag.Flag) {
		completions = append(completions, completion{
			cType:       opt,
			description: f.Usage,
			name:        "--" + f.Name,
			hasArg:      true,
		})
	}
	c.Flags().VisitAll(addCompletion)
	c.PersistentFlags().VisitAll(addCompletion)
	c.InheritedFlags().VisitAll(addCompletion)
	return completions
}
