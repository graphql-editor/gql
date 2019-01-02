package cmd

import (
	"encoding/json"
	"errors"
	"strings"

	"github.com/wolfeidau/unflatten"
)

type Variables map[string]interface{}

func (v Variables) String() string {
	b, err := json.Marshal(v)
	if err != nil {
		panic(err)
	}
	return string(b)
}

func (v Variables) Type() string {
	return "Variables"
}

func (v Variables) Set(val string) error {
	ss := strings.Split(val, "=")
	if len(ss) != 2 {
		return errors.New("must be in {key}={value} format")
	}
	var i interface{}
	err := json.Unmarshal([]byte(ss[1]), &i)
	if err != nil {
		i = ss[1]
	}
	v[ss[0]] = i
	return nil
}

func (v Variables) Unflatten() map[string]interface{} {
	return unflatten.Unflatten(v, func(k string) []string { return strings.Split(k, ".") })
}

type Header map[string]string

func (h Header) String() string {
	b, err := json.Marshal(h)
	if err != nil {
		panic(err)
	}
	return string(b)
}

func (h Header) Type() string {
	return "Header"
}

func (h Header) Set(val string) error {
	ss := strings.Split(val, "=")
	if len(ss) != 2 {
		return errors.New("must be in {header}={value} format")
	}
	h[ss[0]] = ss[1]
	return nil
}

var (
	query         string
	variables     = Variables{}
	operationName string
)
