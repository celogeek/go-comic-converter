package converter

// Name or Section
type converterOrder interface {
	Value() string
}

// Section
type converterOrderSection struct {
	value string
}

func (s converterOrderSection) Value() string {
	return s.value
}

// Name
//
// isString is used to quote the default value.
type converterOrderName struct {
	value    string
	isString bool
}

func (s converterOrderName) Value() string {
	return s.value
}
