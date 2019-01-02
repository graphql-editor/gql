package cmd

import (
	"bytes"
	"encoding/json"
	"fmt"
	"log"
	"os"
	"text/template"

	"github.com/slothking-online/gql/client"
)

func formattedOutput(b []byte) bool {
	// Try to gracefully format output
	// in case of an error be polite,
	// write out an error to stderr
	// but still return the result of
	// query
	if format == "" {
		return false
	}
	t := template.New("format")
	if _, err := t.Parse(format); err != nil {
		fmt.Fprintln(os.Stderr, err)
		return false
	}
	m := make(map[string]interface{})
	if err := json.Unmarshal(b, &m); err != nil {
		fmt.Fprintln(os.Stderr, err)
		return false
	}
	buf := &bytes.Buffer{}
	if err := t.Execute(buf, m); err != nil {
		fmt.Fprintln(os.Stderr, err)
		return false
	}
	fmt.Println(buf.String())
	return true
}

func execute(cli *client.Client, r client.Raw, out interface{}) {
	data, err := cli.Raw(r, out)
	if err != nil {
		if _, ok := err.(client.Errors); !ok {
			fmt.Fprintln(os.Stderr, err)
			os.Exit(1)
		}
	}
	if data != nil {
		b, err := json.MarshalIndent(data, "", "    ")
		if err != nil {
			log.Panic(err)
		}
		if !formattedOutput(b) {
			fmt.Println(string(b))
		}
	}
	if err != nil {
		b, merr := json.MarshalIndent(err, "", "    ")
		if merr != nil {
			log.Panic(merr)
		}
		fmt.Fprintln(os.Stderr, string(b))
	}
}
