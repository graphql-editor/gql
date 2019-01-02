package client

import (
	"encoding/json"
	"fmt"
)

// Location where error occured, if specified.
type Location struct {
	// Line is a number indicating at which line did error occur
	Line int `json:"line"`
	// Column is a number indicating at which column did error occur
	Column int `json:"column"`
}

// Error is GraphQL error object.
type Error struct {
	// Message is an error message
	Message string `json:"message"`
	// Locations is a list of
	Locations  []Location             `json:"locations,omitempty"`
	Path       []interface{}          `json:"path,omitempty"`
	Extensions map[string]interface{} `json:"extensions,omitempty"`
}

// Errors is a list of errors returned by remote GraphQL endpoint
type Errors []Error

// Error formats Errors as json array
func (g Errors) Error() string {
	b, err := json.Marshal(g)
	if err != nil {
		return fmt.Sprintf("could not marshal Errors: %s", err.Error())
	}
	return string(b)
}
