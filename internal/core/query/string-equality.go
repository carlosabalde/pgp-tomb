package query

type stringEquality struct {
	identifier string
	str        string
}

func (e stringEquality) Eval(p Params) bool {
	return p.Get(e.identifier) == e.str
}

func (e stringEquality) String() string {
	return sprintf("(%s == '%s')", e.identifier, e.str)
}

func Equal(identifier, str string) Query {
	return stringEquality{identifier, str}
}
