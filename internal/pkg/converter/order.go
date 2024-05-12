package converter

// Name or Section
type order interface {
	Value() string
}

// Section
type orderSection struct {
	value string
}

func (s orderSection) Value() string {
	return s.value
}

// Name
//
// isString is used to quote the default value.
type orderName struct {
	value    string
	isString bool
}

func (s orderName) Value() string {
	return s.value
}
