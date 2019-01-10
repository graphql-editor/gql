package cmd

import (
	"bytes"
	"encoding/json"
	"fmt"
	"text/template"

	"github.com/slothking-online/gql/client"
)

func formattedOutput(config Config, b []byte) (bool, error) {
	// Try to gracefully format output of
	// the query
	if format == "" {
		return false, nil
	}
	t := template.New("format")
	if _, err := t.Parse(format); err != nil {
		return false, err
	}
	m := make(map[string]interface{})
	if err := json.Unmarshal(b, &m); err != nil {
		return false, err
	}
	buf := &bytes.Buffer{}
	if err := t.Execute(buf, m); err != nil {
		return false, err
	}
	_, err := fmt.Fprintln(config.Output(), buf.String())
	return err == nil, err
}

func execute(config Config, cli *client.Client, r client.Raw, out interface{}) error {
	data, qerr := cli.Raw(r, out)
	if qerr != nil {
		if _, ok := qerr.(client.Errors); !ok {
			return qerr
		}
	}
	if data != nil {
		b, err := json.MarshalIndent(data, "", "    ")
		if err != nil {
			return err
		}
		if ok, err := formattedOutput(config, b); !ok {
			if _, perr := fmt.Fprintln(config.Output(), string(b)); perr != nil {
				return perr
			}
			return err
		}
	}
	if qerr != nil {
		b, merr := json.MarshalIndent(qerr, "", "    ")
		if merr != nil {
			return merr
		}
		fmt.Fprintln(config.Error(), string(b))
	}
	return nil
}
