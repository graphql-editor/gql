package client

import (
	"bytes"
	"encoding/json"
	"errors"
	"io/ioutil"
	"net/http"
	"net/url"
	"testing"

	"github.com/aexol/test_util"
	"github.com/stretchr/testify/mock"

	"github.com/stretchr/testify/assert"
)

type testCaseMinifyQuery struct {
	in  string
	out string
}

func (tt testCaseMinifyQuery) test(t *testing.T) {
	assert := assert.New(t)
	out := minifyQuery(tt.in)
	assert.Equal(tt.out, out)
}

func TestMinifyQuery(t *testing.T) {
	data := []testCaseMinifyQuery{
		{
			in:  "    qqq     qqq   \n qqq    ",
			out: "qqq qqq qqq",
		},
	}
	for _, tt := range data {
		tt.test(t)
	}
}

type testCaseRawValues struct {
	in  Raw
	out url.Values
	err func(assert *assert.Assertions) test_util.ErrorAssertion
}

func (tt testCaseRawValues) test(t *testing.T) {
	assert := assert.New(t)
	if tt.err == nil {
		tt.err = test_util.NoError
	}
	values, err := tt.in.values()
	tt.err(assert)(err)
	assert.Equal(tt.out, values)
}

func TestRawValues(t *testing.T) {
	data := []testCaseRawValues{
		{
			in: Raw{
				Query: "some-query",
			},
			out: url.Values(map[string][]string{
				"query": []string{"some-query"},
			}),
		},
		{
			in: Raw{
				Query: "some-query",
				Variables: map[string]interface{}{
					"var": "val",
				},
			},
			out: url.Values(map[string][]string{
				"query":     []string{"some-query"},
				"variables": []string{`{"var":"val"}`},
			}),
		},
		{
			in: Raw{
				Query:         "some-query",
				OperationName: "opName",
			},
			out: url.Values(map[string][]string{
				"query":         []string{"some-query"},
				"operationName": []string{"opName"},
			}),
		},
		{
			in: Raw{
				Query: "some-query",
				Variables: map[string]interface{}{
					"var": "val",
				},
				OperationName: "opName",
			},
			out: url.Values(map[string][]string{
				"query":         []string{"some-query"},
				"variables":     []string{`{"var":"val"}`},
				"operationName": []string{"opName"},
			}),
		},
		{
			in:  Raw{},
			out: nil,
			err: test_util.Error,
		},
		{
			in: Raw{
				Query: "some-query",
				Variables: map[string]interface{}{
					"bad": func() {},
				},
			},
			out: nil,
			err: test_util.Error,
		},
	}
	for _, tt := range data {
		tt.test(t)
	}
}

type testCaseClientRaw struct {
	in        Raw
	resp      Response
	respErr   error
	targetOut interface{}
	out       interface{}
	err       func(assert *assert.Assertions) test_util.ErrorAssertion
	transport *test_util.MockRoundTripper
	Endpoint  string
}

func (tt testCaseClientRaw) test(t *testing.T) {
	b, _ := json.Marshal(tt.resp)
	assert := assert.New(t)
	jsonHeaders := make(http.Header)
	jsonHeaders.Add("Content-Type", "application/json")
	config := Config{
		Endpoint: tt.Endpoint,
	}
	if tt.transport != nil {
		tt.transport.On("RoundTrip", mock.MatchedBy(func(req *http.Request) bool {
			body, _ := ioutil.ReadAll(req.Body)
			inB, _ := json.Marshal(tt.in)
			return req.Method == "POST" &&
				req.Header.Get("Content-Type") == "application/json" &&
				assert.JSONEq(string(inB), string(body))
		})).Return(
			&http.Response{
				Status:     "OK",
				StatusCode: http.StatusOK,
				Body:       ioutil.NopCloser(bytes.NewBuffer(b)),
				Header:     jsonHeaders,
			},
			tt.respErr,
		)
		config.RoundTripper = tt.transport
	}
	if tt.err == nil {
		tt.err = test_util.NoError
	}
	cli := New(config)
	out, err := cli.Raw(tt.in, tt.targetOut)
	tt.err(assert)(err)
	assert.Equal(tt.out, out)
}

func TestClientRaw(t *testing.T) {
	transport := new(test_util.MockRoundTripper)
	data := []testCaseClientRaw{
		{
			in: Raw{
				Query: "some-query",
				Variables: map[string]interface{}{
					"bad": func() {},
				},
			},
			err: test_util.Error,
		},
		{
			in: Raw{
				Query: "some-query",
			},
			resp: Response{
				Data: map[string]interface{}{
					"field": map[string]interface{}{
						"field2": "value",
					},
				},
			},
			out: map[string]interface{}{
				"field": map[string]interface{}{
					"field2": "value",
				},
			},
			transport: transport,
			Endpoint:  "http://example.com/graphql",
		},
		{
			in: Raw{
				Query: "some-query",
			},
			err: test_util.Error,
		},
		{
			in: Raw{
				Query:  "some-query",
				Method: ":INVALID:",
			},
			Endpoint: "http://example.com/graphql",
			err:      test_util.Error,
		},
		{
			in: Raw{
				Query: "some-query",
			},
			respErr:   errors.New("error"),
			transport: transport,
			Endpoint:  "http://example.com/graphql",
			err:       test_util.Error,
		},
	}
	for _, tt := range data {
		tt.test(t)
	}
}
