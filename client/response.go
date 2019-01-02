package client

// Response represents a valid GraphQL remote response
type Response struct {
	// Data returned from remote endpoint.
	// According to spec GraphQL response should have data
	// field if no errors happend
	Data interface{} `json:"data,omitempty"`
	// Errors is an optional list of errors returned by remote endpoint
	Errors Errors `json:"errors,omitempty"`
}
