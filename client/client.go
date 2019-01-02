// Package client implements a simple wrapper around http.Client
// to make it more GraphQL friendly.
package client

import (
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"net/url"
	"os"
	"regexp"
	"strings"
)

var (
	multipleSpaces = regexp.MustCompile(`[\s\p{Zs}]{2,}`)
)

// Client is a http Client used to
// execute query on remote endpoint
type Client struct {
	http.Client
	// GraphQL http endpoint
	Endpoint string
}

// Raw GraphQL query,
type Raw struct {
	// Query is a query executed against GraphQL endpoint
	Query string `json:"query,omitempty"`
	// Variables are a list of variables for query
	Variables map[string]interface{} `json:"variables,omitempty"`
	// if query has more than one operation defined
	// OperationName indicates which one should be executed
	OperationName string `json:"operationName,omitempty"`
	// Optional http headers
	Header http.Header `json:"-"`
	// optional request method, defaults to POST
	Method string `json:"-"`
}

func (r Raw) values() (url.Values, error) {
	if r.Query == "" {
		return nil, errors.New("query cannot be empty")
	}
	r.Query = minifyQuery(r.Query)
	values := make(url.Values)
	values.Add("query", r.Query)
	if len(r.Variables) != 0 {
		b, err := json.Marshal(r.Variables)
		if err != nil {
			return nil, err
		}
		values.Add("variables", string(b))
	}
	if r.OperationName != "" {
		values.Add("operationName", r.OperationName)
	}
	return values, nil
}

func minifyQuery(q string) string {
	q = strings.Replace(q, "\n", "", -1)
	q = strings.TrimSpace(q)
	q = multipleSpaces.ReplaceAllString(q, " ")
	return q
}

func (c *Client) buildRequest(r Raw) (*http.Request, error) {
	b, err := json.Marshal(r)
	if err != nil {
		return nil, err
	}
	url, err := url.Parse(c.Endpoint)
	if err != nil || c.Endpoint == "" {
		if c.Endpoint == "" {
			err = fmt.Errorf("endpoint cannot be empty")
		}
		return nil, err
	}
	meth := r.Method
	if meth == "" {
		meth = "POST"
	}
	return http.NewRequest(
		meth,
		url.String(),
		bytes.NewBuffer(b),
	)
}

// Raw executes GraphQL query against GraphQL remote
func (c *Client) Raw(r Raw, out interface{}) (interface{}, error) {
	req, err := c.buildRequest(r)
	if err != nil {
		return nil, err
	}
	if r.Header == nil {
		r.Header = make(http.Header)
	}
	r.Header.Add("Content-Type", "application/json")
	req.Header = r.Header
	resp, err := c.Do(req)
	if err != nil {
		return nil, err
	}
	defer func() {
		if cerr := resp.Body.Close(); cerr != nil {
			fmt.Fprintln(os.Stderr, cerr) // nolint: errcheck
		}
	}()
	var gqlResponse Response
	if out != nil {
		gqlResponse.Data = out
	}
	dec := json.NewDecoder(resp.Body)
	if err = dec.Decode(&gqlResponse); err != nil {
		return nil, err
	}
	if len(gqlResponse.Errors) != 0 {
		err = gqlResponse.Errors
	}
	return gqlResponse.Data, err
}

// Config for GraphQL client
type Config struct {
	// Endpoint is remote GraphQL endpoint
	Endpoint string
	// RoundTripper is an optional http.RoundTripper for client
	// if not set falls back to http.DefaultTransport
	RoundTripper http.RoundTripper
}

// New creates new GraphQL client
func New(cfg Config) *Client {
	cli := &Client{
		Endpoint: cfg.Endpoint,
	}
	if cfg.RoundTripper != nil {
		cli.Client = http.Client{
			Transport: cfg.RoundTripper,
		}
	}
	return cli
}
