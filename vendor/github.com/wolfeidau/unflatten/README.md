# unflatten [![GoDoc](https://img.shields.io/badge/godoc-Reference-brightgreen.svg?style=flat)](http://godoc.org/github.com/wolfeidau/unflatten) [![Go Report Card](https://goreportcard.com/badge/github.com/wolfeidau/unflatten)](https://goreportcard.com/report/github.com/wolfeidau/unflatten) [![Build Status](https://travis-ci.org/wolfeidau/unflatten.svg?branch=master)](https://travis-ci.org/wolfeidau/unflatten)

This library can "flatten" and "unflatten" a hierarchy stored in a map[string]interface{}. 

# usage

```go
var m = map[string]interface{}{
	"cpu.usage.0.user": map[string]interface{}{
		"value": 2.3,
	},
	"cpu.usage.0.system": map[string]interface{}{
		"value": 1.2,
	},
}

tree := Unflatten(m, func(k string) []string { return strings.Split(k, ".") })

```

# contributions

Thanks to [Andrew Leap](https://github.com/andyleap) for rewriting this library and reminding me I need to use functions more in golang.

# License

This code is Copyright (c) 2014 Mark Wolfe and licenced under the MIT licence. All rights not explicitly granted in the MIT license are reserved. See the included LICENSE.md file for more details.
