package cmd

import (
	"bytes"
	"testing"

	"github.com/aexol/test_util"

	"github.com/stretchr/testify/mock"

	"github.com/stretchr/testify/assert"

	"github.com/slothking-online/gql/introspection"
)

type testCaseArgsCommand struct {
	args        []string
	field       introspection.Field
	ok          bool
	expectedOut []byte
	err         func(*assert.Assertions) test_util.ErrorAssertion
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
func (tt testCaseArgsCommand) test(t *testing.T) {
	if tt.err == nil {
		tt.err = test_util.NoError
	}
	assert := assert.New(t)
	stdout := &bytes.Buffer{}
	schema := new(argsSchemaMock)
	schema.On("FieldForPath", tt.args).Return(tt.field, tt.ok)
	tt.err(assert)(NewArgsCommand(ArgsCommandConfig{
		Schema: schema,
		Config: Config{
			Out: stdout,
		},
	}).RunE(nil, tt.args))
	assert.Equal(tt.expectedOut, stdout.Bytes())
}

func TestArgsCommand(t *testing.T) {
	data := []testCaseArgsCommand{
		{
			args: []string{"a", "b", "c"},
			err:  test_util.Error,
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
