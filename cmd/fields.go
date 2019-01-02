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
	"fmt"
	"os"
	"strings"

	"github.com/slothking-online/gql/introspection"

	"github.com/agnivade/levenshtein"
	"github.com/spf13/cobra"
)

// fieldsCmd represents the fields command
func NewFieldsCommand(schema introspection.Schema) *cobra.Command {
	return &cobra.Command{
		Use:   "fields",
		Short: "Returns a list of fields on resolve path",
		Long:  `Returns a list of fields that can be referenced on this resolve path.`,
		Run: func(cmd *cobra.Command, args []string) {
			t, ok := schema.TypeForPath(args)
			buf := &bytes.Buffer{}
			if ok {
				// we got exact match in schema, just return fields
				// of that type
				sep := ""
				for _, field := range t.Fields {
					fmt.Fprint(buf, sep)
					fmt.Fprint(buf, field.Name)
					sep = " "
				}
				fmt.Println(buf.String())
				return
			}
			t, ok = schema.TypeForPath(args[:len(args)-1])
			if !ok {
				fmt.Fprintf(os.Stderr, "path not found in schema\n")
				os.Exit(1)
			}
			levMatches := []string{}
			for _, field := range t.Fields {
				distance := levenshtein.ComputeDistance(args[len(args)-1], field.Name)
				// Only accept matches with distance of less
				// than 3 edits
				if distance < 3 {
					levMatches = append(levMatches, field.Name)
				}
			}
			if len(levMatches) == 0 {
				fmt.Fprintf(os.Stderr, "path not found in schema\n")
				os.Exit(1)
			}
			fmt.Fprintf(os.Stderr, "no exact match found\n")
			fmt.Fprintf(os.Stderr, "printing closest matches\n")
			fmt.Println(strings.Join(levMatches, " "))
		},
	}
}
