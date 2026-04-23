package a

type Config struct {
	// required: name must be set for identification
	Name string // want Name:`required: name must be set for identification`
	// required: port is needed for networking
	Port     int // want Port:`required: port is needed for networking`
	Optional string
}

var _ = Config{Name: "x", Port: 80}                   // OK
var _ = Config{Name: "x"}                              // want `missing required field "Port": port is needed for networking`
var _ = Config{Port: 80}                               // want `missing required field "Name": name must be set for identification`
var _ = Config{}                                       // want `missing required field "Name": name must be set for identification` `missing required field "Port": port is needed for networking`
var _ = Config{Name: "x", Port: 80, Optional: "y"}    // OK
var _ = &Config{Name: "x", Port: 80}                  // OK
var _ = &Config{Name: "x"}                             // want `missing required field "Port": port is needed for networking`

// NoRequired has no required fields.
type NoRequired struct {
	A string
	B int
}

var _ = NoRequired{}      // OK
var _ = NoRequired{A: ""} // OK

// Positional syntax should not trigger errors.
var _ = Config{"x", 80, "opt"} // OK
