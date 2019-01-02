package cmd

import (
	"os"
	"testing"

	"github.com/stretchr/testify/mock"

	"github.com/stretchr/testify/assert"
)

type mockReader struct {
	mock.Mock
}

func (m *mockReader) Read(p []byte) (int, error) {
	called := m.Called(p)
	return called.Int(0), called.Error(1)
}

type mockWriter struct {
	mock.Mock
}

func (m *mockWriter) Write(p []byte) (int, error) {
	called := m.Called(p)
	return called.Int(0), called.Error(1)
}

func TestConfig(t *testing.T) {
	assert := assert.New(t)
	assert.Exactly(Config{}.Input(), os.Stdin)
	assert.Exactly(Config{}.Output(), os.Stdout)
	assert.Exactly(Config{}.Error(), os.Stderr)
	assert.Exactly(Config{
		In: new(mockReader),
	}.Input(), new(mockReader))
	assert.Exactly(Config{
		Out: new(mockWriter),
	}.Output(), new(mockWriter))
	assert.Exactly(Config{
		Err: new(mockWriter),
	}.Error(), new(mockWriter))
	osExitCalledWith := -1
	oldExit := osExit
	osExit = func(c int) {
		osExitCalledWith = c
	}
	Config{}.Exit(1)
	assert.Equal(1, osExitCalledWith)
	osExit = oldExit
	osExitCalledWith = -1
	Config{
		ExitFunc: func(c int) {
			osExitCalledWith = c
		},
	}.Exit(1)
	assert.Equal(1, osExitCalledWith)
}
