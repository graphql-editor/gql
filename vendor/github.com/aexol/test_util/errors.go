package test_util

import "github.com/stretchr/testify/assert"

type ErrorAssertion func(error, ...interface{}) bool

func Error(a *assert.Assertions) ErrorAssertion {
	return ErrorAssertion(func(err error, args ...interface{}) bool {
		return a.Error(err, args...)
	})
}

func NoError(a *assert.Assertions) ErrorAssertion {
	return ErrorAssertion(func(err error, args ...interface{}) bool {
		return a.NoError(err, args...)
	})
}
