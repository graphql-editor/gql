package cmd

import (
	"testing"

	"github.com/aexol/test_util"

	"github.com/stretchr/testify/assert"
)

type testCaseFormattedOutput struct {
	fm             string
	in             []byte
	expectedOut    bool
	expectedErr    func(*assert.Assertions) test_util.ErrorAssertion
	expectedWriteB []byte
	writeOutN      int
	writeOutErr    error
}

func (tt testCaseFormattedOutput) test(t *testing.T) {
	assert := assert.New(t)
	if tt.expectedErr == nil {
		tt.expectedErr = test_util.NoError
	}
	format = tt.fm
	writer := new(mockWriter)
	if tt.expectedWriteB != nil {
		writer.On("Write", tt.expectedWriteB).Return(tt.writeOutN, tt.writeOutErr)
	}
	ok, err := formattedOutput(Config{
		Out: writer,
	}, tt.in)
	assert.Equal(tt.expectedOut, ok)
	tt.expectedErr(assert)(err)
}

func TestFormattedOutput(t *testing.T) {
	data := []testCaseFormattedOutput{
		// Test no formatting
		{
			fm:          "",
			in:          []byte(`{"key":"val"}`),
			expectedOut: false,
		},
		// Test erroring out on template compile error
		{
			fm:          "{{if .key}}",
			in:          []byte(`{"key":"val"}`),
			expectedOut: false,
			expectedErr: test_util.Error,
		},
		// Test erroring out on malformed json
		{
			fm:          "{{.key}}",
			in:          []byte(`{"key":"val"`),
			expectedOut: false,
			expectedErr: test_util.Error,
		},
		// Test erroring out on template execution error
		{
			fm:          "{{ .key.val }}",
			in:          []byte(`{"key":null}`),
			expectedOut: false,
			expectedErr: test_util.Error,
		},
		// Success formatting
		{
			fm:             `{{.key}}`,
			in:             []byte(`{"key":"val"}`),
			expectedOut:    true,
			expectedWriteB: []byte("val\n"),
			writeOutN:      3,
		},
	}
	for _, tt := range data {
		tt.test(t)
	}
}

func TestExecute(t *testing.T) {}
