package converter

type Order interface {
	Value() string
}

type OrderSection struct {
	value string
}

func (s OrderSection) Value() string {
	return s.value
}

type OrderName struct {
	value    string
	isString bool
}

func (s OrderName) Value() string {
	return s.value
}
