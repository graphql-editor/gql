package cmd

import (
	"bytes"
	"runtime/debug"
	"testing"

	"github.com/stretchr/testify/mock"

	"github.com/stretchr/testify/assert"

	"github.com/slothking-online/gql/introspection"
)

type testCaseArgsCommand struct {
	args         []string
	field        introspection.Field
	ok           bool
	expectedOut  []byte
	expectedErr  []byte
	expectedCode int
}

type argsSchemaMock struct {
	mock.Mock
}

func (a *argsSchemaMock) FieldForPath(p []string) (introspection.Field, bool) {
	call := a.Called(p)
	f := call.Get(0).(introspection.Field)
	ok := call.Bool(1)
	return f, ok
}

type fakeout struct{}

func (_ fakeout) Write([]byte) (int, error) {
	debug.PrintStack()
	return 0, nil
}

func (tt testCaseArgsCommand) test(t *testing.T) {
	assert := assert.New(t)
	stdout := &bytes.Buffer{}
	stderr := &bytes.Buffer{}
	var code int
	exit := func(c int) {
		code = c
	}
	schema := new(argsSchemaMock)
	schema.On("FieldForPath", tt.args).Return(tt.field, tt.ok)
	NewArgsCommand(ArgsCommandConfig{
		Schema: schema,
		Config: Config{
			Out:      stdout,
			Err:      stderr,
			ExitFunc: exit,
		},
	}).Run(nil, tt.args)
	assert.Equal(tt.expectedOut, stdout.Bytes())
	assert.Equal(tt.expectedErr, stderr.Bytes())
	assert.Equal(tt.expectedCode, code)
}

func TestArgsCommand(t *testing.T) {
	data := []testCaseArgsCommand{
		{
			args:         []string{"a", "b", "c"},
			expectedErr:  []byte("path not found in schema\n"),
			expectedCode: 1,
		},
		{
			args: []string{"a", "b", "c"},
			field: introspection.Field{
				Args: []introspection.Arg{
					introspection.Arg{Name: "arg1"},
					introspection.Arg{Name: "arg2"},
				},
			},
			ok:          true,
			expectedOut: []byte("arg1 arg2\n"),
		},
	}
	for _, tt := range data {
		tt.test(t)
	}
}
