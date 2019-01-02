package client

import (
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestGraphqlClientError(t *testing.T) {
	assert := assert.New(t)
	err := Errors{
		Error{
			Path: []interface{}{
				func() {},
			},
		},
	}
	assert.True(strings.HasPrefix(err.Error(), "could not marshal Errors: "))
	err = Errors{
		Error{
			Message: "message1",
		},
		Error{
			Message: "message2",
		},
	}
	assert.JSONEq(`[{"message": "message1"},{"message":"message2"}]`, err.Error())
}
