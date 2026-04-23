package c

import "b"

var _ = b.Service{Endpoint: "http://example.com"} // OK
var _ = b.Service{}                                // want `missing required field "Endpoint": endpoint URL is mandatory`
var _ = b.Service{Timeout: 30}                     // want `missing required field "Endpoint": endpoint URL is mandatory`
