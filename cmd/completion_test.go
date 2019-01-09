package cmd

import (
	"testing"

	"github.com/stretchr/testify/mock"

	"github.com/aexol/test_util"

	"github.com/spf13/cobra"
	"github.com/stretchr/testify/assert"
)

func TestCommandBuilderFunc(t *testing.T) {
	assert := assert.New(t)
	expectedArgs := []string{"arg1", "arg"}
	expectedCmd := &cobra.Command{}
	var args []string
	cmd := CommandBuilderFunc(func(cargs []string) *cobra.Command {
		args = cargs
		return expectedCmd
	}).New(expectedArgs)
	assert.Equal(expectedArgs, args)
	assert.Equal(expectedCmd, cmd)
}

func TestGetSubcommandCompletions(t *testing.T) {
	c := &cobra.Command{}
	c.AddCommand(&cobra.Command{
		Use:   "somename",
		Short: "descdesc",
	})
	assert := assert.New(t)
	assert.Equal([]completion{
		completion{
			cType:       cmd,
			description: "descdesc",
			name:        "somename",
		},
	}, getSubcommandCompletions(c))
}

type testCaseCompletionCommand struct {
	config      CompletionCommandConfig
	args        []string
	expectedOut []byte
	err         func(*assert.Assertions) test_util.ErrorAssertion
	ok          bool
}

func (tt testCaseCompletionCommand) test(t *testing.T) {
	assert := assert.New(t)
	out := new(mockWriter)
	if tt.err == nil {
		tt.err = test_util.NoError
	}
	if tt.expectedOut != nil {
		out.On("Write", tt.expectedOut).Return(len(tt.expectedOut), nil)
	}
	tt.config.Out = out
	tt.err(assert)(NewCompletionCommand(tt.config).RunE(nil, tt.args))
	if tt.expectedOut == nil {
		out.AssertNotCalled(t, "Write", mock.Anything)
	}
}

type mockCommandBuilder struct {
	mock.Mock
}

func (m *mockCommandBuilder) New(args []string) *cobra.Command {
	called := m.Called(args)
	return called.Get(0).(*cobra.Command)
}

func TestNewCompletionCommand(t *testing.T) {
	makeBuilder := func(
		args []string,
		cmd *cobra.Command,
		setup func(*cobra.Command),
	) CommandBuilder {
		mock := new(mockCommandBuilder)
		mock.On("New", args).Return(cmd)
		if setup != nil {
			setup(cmd)
		}
		return mock
	}
	data := []testCaseCompletionCommand{
		testCaseCompletionCommand{
			args: []string{},
			err:  test_util.Error,
		},
		testCaseCompletionCommand{
			args:        []string{"bash"},
			expectedOut: []byte(bashCompletion + "\n"),
		},
		testCaseCompletionCommand{
			args:        []string{"zsh"},
			expectedOut: []byte(zshCompletion + "\n"),
		},
		testCaseCompletionCommand{
			args: []string{"aa e\" bc"},
			err:  test_util.Error,
		},
		testCaseCompletionCommand{
			args: []string{"cmd was called"},
			config: CompletionCommandConfig{
				CommandBuilder: makeBuilder(
					[]string{"was", "called"},
					&cobra.Command{},
					nil,
				),
			},
		},
		testCaseCompletionCommand{
			args: []string{"cmd --some-opt"},
			config: CompletionCommandConfig{
				CommandBuilder: makeBuilder(
					[]string{"--some-opt"},
					&cobra.Command{},
					nil,
				),
			},
		},
		testCaseCompletionCommand{
			args: []string{"cmd"},
			config: CompletionCommandConfig{
				CommandBuilder: makeBuilder(
					[]string{},
					&cobra.Command{},
					nil,
				),
			},
		},
		testCaseCompletionCommand{
			args: []string{"cmd"},
			config: CompletionCommandConfig{
				CommandBuilder: makeBuilder(
					[]string{},
					&cobra.Command{},
					func(cmd *cobra.Command) {
						cmd.AddCommand(&cobra.Command{
							Use:   "sub1",
							Short: "sub1desc",
						})
						cmd.AddCommand(&cobra.Command{
							Use:   "sub2",
							Short: "sub2desc",
						})
						cmd.Flags().String("flag", "", "flagdesc")
					},
				),
			},
			expectedOut: []byte(`opt:--flag:flagdesc:true
cmd:sub1:sub1desc:false
cmd:sub2:sub2desc:false
`),
		},
		// Test if sub command reached
		testCaseCompletionCommand{
			args: []string{"cmd", "sub1"},
			config: CompletionCommandConfig{
				CommandBuilder: makeBuilder(
					[]string{},
					&cobra.Command{},
					func(cmd *cobra.Command) {
						sub1 := &cobra.Command{
							Use:   "sub1",
							Short: "sub1desc",
						}
						sub1.Flags().String("flag", "", "flagdesc")
					},
				),
			},
			expectedOut: []byte(`opt:--flag:flagdesc:true
`),
		},
		// Test flag order
		testCaseCompletionCommand{
			args: []string{"cmd sub1"},
			config: CompletionCommandConfig{
				CommandBuilder: makeBuilder(
					[]string{"sub1"},
					&cobra.Command{},
					func(cmd *cobra.Command) {
						sub1 := &cobra.Command{
							Use:   "sub1",
							Short: "sub1desc",
						}
						cmd.AddCommand(sub1)
						cmd.PersistentFlags().String("flag", "", "flagdesc")
						sub1.Flags().String("subflag", "", "subflagdesc")
					},
				),
			},
			expectedOut: []byte(`opt:--subflag:subflagdesc:true
opt:--flag:flagdesc:true
`),
		},
	}
	for _, tt := range data {
		tt.test(t)
	}
}

func TestFlagCompletions(t *testing.T) {
	c := &cobra.Command{}
	c.Flags().String("flag1", "", "flag1 desc")
	c.PersistentFlags().String("flag2", "", "flag2 desc")
	sc := &cobra.Command{}
	c.AddCommand(sc)
	sc.Flags().String("flag3", "", "flag3 desc")
	sc.PersistentFlags().String("flag4", "", "flag4 desc")
	assert := assert.New(t)
	assert.Equal([]completion{
		completion{
			cType:       opt,
			description: "flag3 desc",
			name:        "--flag3",
			hasArg:      true,
		},
		completion{
			cType:       opt,
			description: "flag4 desc",
			name:        "--flag4",
			hasArg:      true,
		},
		completion{
			cType:       opt,
			description: "flag2 desc",
			name:        "--flag2",
			hasArg:      true,
		},
	}, getFlagCompletions(sc))
}
