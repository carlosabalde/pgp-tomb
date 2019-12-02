package query

type stringEquality struct {
	identifier string
	str        string
}

func (self stringEquality) Eval(identifiers Identifiers) bool {
	return identifiers.Get(self.identifier) == self.str
}

func (self stringEquality) String() string {
	return sprintf("(%s == '%s')", self.identifier, self.str)
}

func Equal(identifier, str string) Query {
	return stringEquality{identifier, str}
}
