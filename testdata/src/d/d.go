package d

// Empty struct.
type Empty struct{}

var _ = Empty{} // OK

// Struct with only optional fields.
type AllOptional struct {
	X int
	Y string
}

var _ = AllOptional{}     // OK
var _ = AllOptional{X: 1} // OK

// Struct with embedded field.
type Base struct {
	// required: base name is required
	BaseName string // want BaseName:`required: base name is required`
}

type Derived struct {
	Base
	// required: derived value is required
	Value int // want Value:`required: derived value is required`
}

var _ = Derived{Value: 1}             // OK (embedded field Base is not required)
var _ = Derived{}                     // want `missing required field "Value": derived value is required`
var _ = Derived{Base: Base{}, Value: 1} // want `missing required field "BaseName": base name is required`
